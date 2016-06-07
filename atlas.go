// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"github.com/go-gl/gl/v3.2-compatibility/gl"
	"image"
	"image/png"
	"os"
	"unsafe"
)

// A node represents an area of an atlas texture which
// has been allocated for use.
type atlasNode struct {
	x int32 // region x
	y int32 // region y + height
	z int32 // region width
}

// A region denotes an allocated chunk of space in an atlas.
type AtlasRegion struct {
	X int32
	Y int32
	W int32
	H int32
}

// A texture atlas is used to tightly pack arbitrarily many small images
// into a single texture.
//
// The actual implementation is based on the article by Jukka Jylänki:
// "A Thousand Ways to Pack the Bin - A Practical Approach to Two-Dimensional
// Rectangle Bin Packing", February 27, 2010.
//
// More precisely, this is an implementation of the
// 'Skyline Bottom-Left' algorithm.
type TextureAtlas struct {
	nodes   []atlasNode // Allocated nodes.
	data    []byte      // Atlas pixel data.
	used    uint        // Allocated surface size.
	width   int32         // Width (in pixels) of the underlying texture.
	height  int32         // Height (in pixels) of the underlying texture.
	depth   int32         // Color depth of the underlying texture.
	texture uint32  // Glyph texture.
}

// NewAtlas creates a new texture atlas.
//
// The given width, height and depth determine the size and depth of
// the underlying texture.
//
// depth should be 1, 3 or 4 and it will specify if the texture is
// created with Alpha, RGB or RGBA channels.
// The image data supplied through Atlas.Set() should be of the same format.
func NewTextureAtlas(width, height, depth int32) *TextureAtlas {
	switch depth {
	case 1, 3, 4:
	default:
		panic("Invalid depth value")
	}

	a := new(TextureAtlas)
	a.width = width
	a.height = height
	a.depth = depth
	a.used = 0
	a.data = make([]byte, width*height*depth)

	// We want a one pixel border around the whole atlas to avoid
	// any artefacts when sampling our texture.
	a.nodes = append(a.nodes, atlasNode{1, 1, width - 2})
	gl.GenTextures(1, &a.texture)
	return a
}

// Release clears all atlas resources.
func (a *TextureAtlas) Release() {
	a.data = nil
	a.nodes = nil
	gl.DeleteTextures(1, &a.texture)
	a.texture = 0
	a.width = 0
	a.height = 0
	a.depth = 0
	a.used = 0
}

// Clear removes all allocated regions from the atlas.
// This invalidates any previously allocated regions.
func (a *TextureAtlas) Clear() {
	a.used = 0
	a.nodes = a.nodes[:1]

	// We want a one pixel border around the whole atlas to avoid
	// any artefacts when sampling our texture.
	a.nodes[0].x = 1
	a.nodes[0].y = 1
	a.nodes[0].z = a.width - 2

	pix := a.data
	for i := range pix {
		pix[i] = 0
	}
}

// Bind binds the atlas texture, so it can be used for rendering.
func (a *TextureAtlas) Bind(target uint32) {
	gl.BindTexture(target, a.texture)
}

// Unbind unbinds the current texture.
// Note that this applies to any texture currently active.
// If this is not the atlas texture, it will still perform the action.
func (a *TextureAtlas) Unbind(target uint32) {
	gl.BindTexture(target, 0)
}

// Commit creates the actual texture from the atlas image data.
// This should be called after all regions have been defined and set,
// and before you start using the texture for display.
func (a *TextureAtlas) Commit(target uint32) {
	gl.Enable(target)

	gl.BindTexture(target, a.texture)

	gl.TexParameteri(target, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(target, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(target, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(target, gl.TEXTURE_MIN_FILTER, gl.LINEAR)

	switch a.depth {
	case 4:
		gl.TexImage2D(target, 0, gl.RGBA, a.width, a.height,
			0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&a.data))

	case 3:
		gl.TexImage2D(target, 0, gl.RGB, a.width, a.height,
			0, gl.RGB, gl.UNSIGNED_BYTE, unsafe.Pointer(&a.data))

	case 1:
		gl.TexImage2D(target, 0, gl.ALPHA, a.width, a.height,
			0, gl.ALPHA, gl.UNSIGNED_BYTE, unsafe.Pointer(&a.data))
	}

}

// Allocate allocates a new region of the given dimensions in the atlas.
// It returns false if the allocation failed. This can happen when the
// specified dimensions exceed atlas bounds, or the atlas is full.
func (a *TextureAtlas) Allocate(width, height int32) (AtlasRegion, bool) {
	var region AtlasRegion
	region.X = 0
	region.Y = 0
	region.W = width
	region.H = height

	bestIndex := -1
	bestWidth := int32(1<<31 - 1)
	bestHeight := int32(1<<31 - 1)

	for index := range a.nodes {
		y := a.fit(index, width, height)

		if y < 0 {
			continue
		}

		node := a.nodes[index]

		if ((y + height) < bestHeight) || (((y + height) == bestHeight) && (node.z < bestWidth)) {
			bestHeight = y + height
			bestIndex = index
			bestWidth = node.z
			region.X = node.x
			region.Y = y
		}
	}

	if bestIndex == -1 {
		return region, false
	}

	// Insert the node at bestIndex
	a.nodes = append(a.nodes, atlasNode{})
	copy(a.nodes[bestIndex+1:], a.nodes[bestIndex:])
	a.nodes[bestIndex] = atlasNode{region.X, region.Y + height, width}

	// Adjust subsequent nodes.
	for i := bestIndex + 1; i < len(a.nodes); i++ {
		curr := &a.nodes[i]
		prev := &a.nodes[i-1]

		if curr.x >= prev.x+prev.z {
			break
		}

		shrink := prev.x + prev.z - curr.x
		curr.x += shrink
		curr.z -= shrink

		if curr.z > 0 {
			break
		}

		copy(a.nodes[i:], a.nodes[i+1:])
		a.nodes = a.nodes[:len(a.nodes)-1]
		i--
	}

	a.merge()
	a.used += uint(width * height)
	return region, true
}

// Set pastes the given data into the atlas buffer at the given coordinates.
// It assumes there is enough space available for the data to fit.
func (a *TextureAtlas) Set(region AtlasRegion, src []byte, stride int32) {
	depth := a.depth
	x := region.X
	y := region.Y
	height := region.H
	dst := a.data

	for i := int32(0); i < height; i++ {
		dp := ((y+i)*a.width + x) * depth
		sp := i * stride
		copy(
			dst[dp:dp+stride],
			src[sp:sp+stride],
		)
	}
}

// Save saves the texture as a PNG image.
func (a *TextureAtlas) Save(file string) (err error) {
	fd, err := os.Create(file)
	if err != nil {
		return
	}

	defer fd.Close()

	rect := image.Rect(0, 0, int(a.width), int(a.height))

	switch a.depth {
	case 1:
		img := image.NewAlpha(rect)
		copy(img.Pix, a.data)
		err = png.Encode(fd, img)

	case 3:
		img := image.NewRGBA(rect)
		copy(img.Pix, a.data)
		err = png.Encode(fd, img)

	case 4:
		img := image.NewRGBA(rect)
		copy(img.Pix, a.data)
		err = png.Encode(fd, img)
	}

	return
}

// Width returns the underlying texture width in pixels.
func (a *TextureAtlas) Width() int32 { return a.width }

// Height returns the underlying texture height in pixels.
func (a *TextureAtlas) Height() int32 { return a.height }

// Depth returns the underlying texture color depth.
func (a *TextureAtlas) Depth() int32 { return a.depth }

// fit checks if the given dimensions fit in the given node.
// If not, it checks any subsequent nodes for a match.
// It returns the height for the last checked node.
// Returns -1 if the width or height exceed texture capacity.
func (a *TextureAtlas) fit(index int, width, height int32) int32 {
	node := a.nodes[index]

	if node.x+width > a.width-1 {
		return -1
	}

	y := node.y
	remainder := width

	for remainder > 0 {
		node = a.nodes[index]

		if node.y > y {
			y = node.y
		}

		if y+height > a.height-1 {
			return -1
		}

		remainder -= node.z
		index++
	}

	return y
}

// merge merges nodes where possible.
// This is the case when two regions overlap.
func (a *TextureAtlas) merge() {
	for i := 0; i < len(a.nodes)-1; i++ {
		node := &a.nodes[i]
		next := a.nodes[i+1]

		if node.y != next.y {
			continue
		}

		node.z += next.z

		copy(a.nodes[i+1:], a.nodes[i+2:])
		a.nodes = a.nodes[:len(a.nodes)-1]
		i--
	}
}
