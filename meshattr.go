// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh
import (
	"github.com/go-gl/gl/v3.2-compatibility/gl"
	"unsafe"
)

// Pre-defined attribute names.
//
// These are used by all render modes other than RenderShader
// to deterine what the purpose of a given attribute is.
const (
	mbPositionKey = "position"
	mbColorKey    = "color"
	mbNormalKey   = "normal"
	mbTexCoordKey = "texcoord"
	mbIndexKey    = "index"
)

// An Attr describes the type and size of a single vertex component.
// These tell the MeshBuffer how to interpret mesh data.
type Attr struct {
	data    interface{} // Data store.
	name    string      // Attribute name.
	vbo     uint32   // gl.Buffer   // Vertex buffer identity.
	target  uint32   // Buffer type.
	usage   uint32   // Usage type of this attribute.
	typ     uint32   // Attribute type.
	size    int32         // Component size (number of elements).
	stride  int32         // Size of component in bytes.
	gpuSize int         // Size of data on GPU.
	invalid bool        // Do we require re-committing?
}

// NewAttr creates a new mesh attribute for the given size,
// type, usage value and name.
//
// In RenderShader mode, the name is the variable by which the component
// can be referenced from a shader program.
//
// In other modes, the MeshBuffer uses this name to identify the attribute's
// purpose. In these cases, it is advised to use the NewIndexAttr,
// NewPositionAttr, etc. wrappers.
func NewAttr(name string, size int32, typ, usage uint32) *Attr {
	a := new(Attr)
	a.name = name

	if size == 0 {
		return a
	}

	a.target = gl.ARRAY_BUFFER
	a.size = size
	a.usage = usage
	a.typ = typ

	a.Clear()
	a.stride = Sizeof(typ)
	return a
}

// NewPositionAttr creates a new vertex position attribute.
func NewPositionAttr(size int32, typ, usage uint32) *Attr {
	return NewAttr(mbPositionKey, size, typ, usage)
}

// NewColorAttr creates a new vertex color attribute.
func NewColorAttr(size int32, typ, usage uint32) *Attr {
	return NewAttr(mbColorKey, size, typ, usage)
}

// NewNormalAttr creates a new surface normal attribute.
func NewNormalAttr(size int32, typ, usage uint32) *Attr {
	return NewAttr(mbNormalKey, size, typ, usage)
}

// NewTexCoordAttr creates a new vertex texture coordinate attribute.
func NewTexCoordAttr(size int32, typ, usage uint32) *Attr {
	return NewAttr(mbTexCoordKey, size, typ, usage)
}

// NewIndexAttr creates a new index attribute.
func NewIndexAttr(size int32, typ, usage uint32) *Attr {
	a := NewAttr(mbIndexKey, size, typ, usage)
	a.target = gl.ELEMENT_ARRAY_BUFFER
	return a
}

// init initializes some of the attribute fields.
// These will be defined by the mesh buffer.
func (a *Attr) init(mode RenderMode) {
	switch mode {
	case RenderClassic, RenderArrays:
		// No VBO in classic and vertex array modes.
	default:
		gl.GenBuffers(1, &a.vbo)
	}
}

// release release attribute resources.
func (a *Attr) release() {
	if a.vbo != 0 {
		gl.DeleteBuffers(1, &a.vbo)
		a.vbo = 0
	}

	a.gpuSize = 0
	a.data = nil
}

// Clear clears the attribute buffer.
func (a *Attr) Clear() {
	a.gpuSize = 0

	switch a.typ {
	case 0: // Null attribute -- do nothing
	case gl.BYTE:
		a.data = []int8{}
	case gl.UNSIGNED_BYTE:
		a.data = []uint8{}
	case gl.SHORT:
		a.data = []int16{}
	case gl.UNSIGNED_SHORT:
		a.data = []uint16{}
	case gl.INT:
		a.data = []int32{}
	case gl.UNSIGNED_INT:
		a.data = []uint32{}
	case gl.FLOAT:
		a.data = []float32{}
	case gl.DOUBLE:
		a.data = []float64{}
	default:
		panic("Invalid attribute type")
	}
}

// Name returns the attribute name.
func (a *Attr) Name() string { return a.name }

// Data returns the atribute data store.
//
// This value should be asserted to a concrete type, which would
// be a slice of the attrbute's type. E.g.: []float32, []uint8, etc.
func (a *Attr) Data() interface{} { return a.data }

// Invalid returns true if the data store needs to be re-committed.
func (a *Attr) Invalid() bool { return a.invalid }

// Invalidate marks the data store as invalid.
// This should be done any time the data is modified externally.
// It will be re-committed on the next render pass.
func (a *Attr) Invalidate() { a.invalid = true }

// Size returns the number of elements in a vertext component for this attribute.
func (a *Attr) Size() int32 { return a.size }

// Stride returns the stride value for the data type this attribute holds.
func (a *Attr) Stride() int32 { return a.stride }

// Type returns the data type of the attribute.
func (a *Attr) Type() uint32 { return a.typ }

// bind binds the buffer.
func (a *Attr) bind() { gl.BindBuffer(a.target, a.vbo)}

// unbind unbinds the buffer.
func (a *Attr) unbind() { gl.BindBuffer(a.target, 0)}

// Target returns the buffer target.
func (a *Attr) Target() uint32 { return a.target }

// SetTarget sets the buffer target.
func (a *Attr) SetTarget(t uint32) { a.target = t }

// Len returns the number of elements in the data store.
func (a *Attr) Len() int32 {
	switch v := a.data.(type) {
	case []int8:
		return int32(len(v))
	case []uint8:
		return int32(len(v))
	case []int16:
		return int32(len(v))
	case []uint16:
		return int32(len(v))
	case []int32:
		return int32(len(v))
	case []uint32:
		return int32(len(v))
	case []float32:
		return int32(len(v))
	case []float64:
		return int32(len(v))
	}

	return 0
}

// buffer buffers the mesh data on the GPU.
// This calls glBufferData or glBufferSubData where appropriate.
func (a *Attr) buffer() {
	switch v := a.data.(type) {
	case []int8:
		size := len(v) * int(a.stride)

		if size != a.gpuSize {
			gl.BufferData(a.target, size, unsafe.Pointer(&v[0]), a.usage)
			a.gpuSize = size
		} else {
			gl.BufferSubData(a.target, 0, size, unsafe.Pointer(&v[0]))
		}
	case []uint8:
		size := len(v) * int(a.stride)

		if size != a.gpuSize {
			gl.BufferData(a.target, size, unsafe.Pointer(&v[0]), a.usage)
			a.gpuSize = size
		} else {
			gl.BufferSubData(a.target, 0, size, unsafe.Pointer(&v[0]))
		}
	case []int16:
		size := len(v) * int(a.stride)

		if size != a.gpuSize {
			gl.BufferData(a.target, size, unsafe.Pointer(&v[0]), a.usage)
			a.gpuSize = size
		} else {
			gl.BufferSubData(a.target, 0, size, unsafe.Pointer(&v[0]))
		}
	case []uint16:
		size := len(v) * int(a.stride)

		if size != a.gpuSize {
			gl.BufferData(a.target, size, unsafe.Pointer(&v[0]), a.usage)
			a.gpuSize = size
		} else {
			gl.BufferSubData(a.target, 0, size, unsafe.Pointer(&v[0]))
		}
	case []int32:
		size := len(v) * int(a.stride)

		if size != a.gpuSize {
			gl.BufferData(a.target, size, unsafe.Pointer(&v[0]), a.usage)
			a.gpuSize = size
		} else {
			gl.BufferSubData(a.target, 0, size, unsafe.Pointer(&v[0]))
		}
	case []uint32:
		size := len(v) * int(a.stride)

		if size != a.gpuSize {
			gl.BufferData(a.target, size, unsafe.Pointer(&v[0]), a.usage)
			a.gpuSize = size
		} else {
			gl.BufferSubData(a.target, 0, size, unsafe.Pointer(&v[0]))
		}
	case []float32:
		size := len(v) * int(a.stride)

		if size != a.gpuSize {
			gl.BufferData(a.target, size, unsafe.Pointer(&v[0]), a.usage)
			a.gpuSize = size
		} else {
			gl.BufferSubData(a.target, 0, size, unsafe.Pointer(&v[0]))
		}
	case []float64:
		size := len(v) * int(a.stride)

		if size != a.gpuSize {
			gl.BufferData(a.target, size, unsafe.Pointer(&v[0]), a.usage)
			a.gpuSize = size
		} else {
			gl.BufferSubData(a.target, 0, size, unsafe.Pointer(&v[0]))
		}

	}

	a.invalid = false
}

// increment increments the value at the given range by the supplied value.
// This is used internally by the mesh buffer.
func (a *Attr) increment(start int, value float64) {
	switch v := a.data.(type) {
	case []int8:
		for i := start; i < len(v); i++ {
			v[i] += int8(value)
		}
	case []uint8:
		for i := start; i < len(v); i++ {
			v[i] += uint8(value)
		}
	case []int16:
		for i := start; i < len(v); i++ {
			v[i] += int16(value)
		}
	case []uint16:
		for i := start; i < len(v); i++ {
			v[i] += uint16(value)
		}
	case []int32:
		for i := start; i < len(v); i++ {
			v[i] += int32(value)
		}
	case []uint32:
		for i := start; i < len(v); i++ {
			v[i] += uint32(value)
		}
	case []float32:
		for i := start; i < len(v); i++ {
			v[i] += float32(value)
		}
	case []float64:
		for i := start; i < len(v); i++ {
			v[i] += value
		}
	}

	a.invalid = true
}

// append appends the given slice to the data store.
// We expect a slice of the appropriate type. E.g.: []uint8, []float32, etc.
func (a *Attr) append(data interface{}) int32 {
	var n int

	switch va := a.data.(type) {
	case []int8:
		vb := data.([]int8)
		a.data = append(va, vb...)
		n = len(vb)

	case []uint8:
		vb := data.([]uint8)
		a.data = append(va, vb...)
		n = len(vb)

	case []int16:
		vb := data.([]int16)
		a.data = append(va, vb...)
		n = len(vb)

	case []uint16:
		vb := data.([]uint16)
		a.data = append(va, vb...)
		n = len(vb)

	case []int32:
		vb := data.([]int32)
		a.data = append(va, vb...)
		n = len(vb)

	case []uint32:
		vb := data.([]uint32)
		a.data = append(va, vb...)
		n = len(vb)

	case []float32:
		vb := data.([]float32)
		a.data = append(va, vb...)
		n = len(vb)

	case []float64:
		vb := data.([]float64)
		a.data = append(va, vb...)
		n = len(vb)
	}

	a.invalid = true
	return int32(n)
}

// Ptr returns a pointer to the element indicated by index.
// Used in RenderArrays mode.
func (a *Attr) ptr(index int) unsafe.Pointer {
	switch v := a.data.(type) {
	case []int8:
		return unsafe.Pointer(&v[index])
	case []uint8:
		return unsafe.Pointer(&v[index])
	case []int16:
		return unsafe.Pointer(&v[index])
	case []uint16:
		return unsafe.Pointer(&v[index])
	case []int32:
		return unsafe.Pointer(&v[index])
	case []uint32:
		return unsafe.Pointer(&v[index])
	case []float32:
		return unsafe.Pointer(&v[index])
	case []float64:
		return unsafe.Pointer(&v[index])
	}

	return unsafe.Pointer(nil)
}

// vertex draws vertices.
// Used in classic render mode.
func (a *Attr) vertex(i int) {
	i *= int(a.size)

	switch a.size {
	case 2:
		switch v := a.data.(type) {
		case []int16:
			gl.Vertex2s(v[i], v[i+1])
		case []int32:
			gl.Vertex2i(v[i], v[i+1])
		case []float32:
			gl.Vertex2f(v[i], v[i+1])
		case []float64:
			gl.Vertex2d(v[i], v[i+1])
		}
	case 3:
		switch v := a.data.(type) {
		case []int16:
			gl.Vertex3s(v[i], v[i+1], v[i+2])
		case []int32:
			gl.Vertex3i(v[i], v[i+1], v[i+2])
		case []float32:
			gl.Vertex3f(v[i], v[i+1], v[i+2])
		case []float64:
			gl.Vertex3d(v[i], v[i+1], v[i+2])
		}
	case 4:
		switch v := a.data.(type) {
		case []int16:
			gl.Vertex4s(v[i], v[i+1], v[i+2], v[i+3])
		case []int32:
			gl.Vertex4i(v[i], v[i+1], v[i+2], v[i+3])
		case []float32:
			gl.Vertex4f(v[i], v[i+1], v[i+2], v[i+3])
		case []float64:
			gl.Vertex4d(v[i], v[i+1], v[i+2], v[i+3])
		}
	}
}

// texcoord defines vertex texture coordinates.
// Used in classic render mode.
func (a *Attr) texcoord(i int) {
	i *= int(a.size)

	switch a.size {
	case 1:
		switch v := a.data.(type) {
		case []int16:
			gl.TexCoord1s(v[i])
		case []int32:
			gl.TexCoord1i(v[i])
		case []float32:
			gl.TexCoord1f(v[i])
		case []float64:
			gl.TexCoord1d(v[i])
		}
	case 2:
		switch v := a.data.(type) {
		case []int16:
			gl.TexCoord2s(v[i], v[i+1])
		case []int32:
			gl.TexCoord2i(v[i], v[i+1])
		case []float32:
			gl.TexCoord2f(v[i], v[i+1])
		case []float64:
			gl.TexCoord2d(v[i], v[i+1])
		}
	case 3:
		switch v := a.data.(type) {
		case []int16:
			gl.TexCoord3s(v[i], v[i+1], v[i+2])
		case []int32:
			gl.TexCoord3i(v[i], v[i+1], v[i+2])
		case []float32:
			gl.TexCoord3f(v[i], v[i+1], v[i+2])
		case []float64:
			gl.TexCoord3d(v[i], v[i+1], v[i+2])
		}
	case 4:
		switch v := a.data.(type) {
		case []int16:
			gl.TexCoord4s(v[i], v[i+1], v[i+2], v[i+3])
		case []int32:
			gl.TexCoord4i(v[i], v[i+1], v[i+2], v[i+3])
		case []float32:
			gl.TexCoord4f(v[i], v[i+1], v[i+2], v[i+3])
		case []float64:
			gl.TexCoord4d(v[i], v[i+1], v[i+2], v[i+3])
		}
	}

}

// normal defines surface normals.
// Used in classic render mode.
func (a *Attr) normal(i int) {
	if a.size != 3 {
		return
	}

	i *= int(a.size)

	switch v := a.data.(type) {
	case []int8:
		gl.Normal3b(v[i], v[i+1], v[i+2])
	case []int16:
		gl.Normal3s(v[i], v[i+1], v[i+2])
	case []int32:
		gl.Normal3i(v[i], v[i+1], v[i+2])
	case []float32:
		gl.Normal3f(v[i], v[i+1], v[i+2])
	case []float64:
		gl.Normal3d(v[i], v[i+1], v[i+2])
	}
}

// Used in classic render mode.
// Defines vertex colors.
func (a *Attr) color(i int) {
	i *= int(a.size)

	switch a.size {
	case 3:
		switch v := a.data.(type) {
		case []int8:
			gl.Color3b(v[i], v[i+1], v[i+2])
		case []uint8:
			gl.Color3ub(v[i], v[i+1], v[i+2])
		case []int16:
			gl.Color3s(v[i], v[i+1], v[i+2])
		case []int32:
			gl.Color3i(v[i], v[i+1], v[i+2])
		case []float32:
			gl.Color3f(v[i], v[i+1], v[i+2])
		case []float64:
			gl.Color3d(v[i], v[i+1], v[i+2])
		}
	case 4:
		switch v := a.data.(type) {
		case []int8:
			gl.Color4b(v[i], v[i+1], v[i+2], v[i+3])
		case []uint8:
			gl.Color4ub(v[i], v[i+1], v[i+2], v[i+3])
		case []int16:
			gl.Color4s(v[i], v[i+1], v[i+2], v[i+3])
		case []int32:
			gl.Color4i(v[i], v[i+1], v[i+2], v[i+3])
		case []float32:
			gl.Color4f(v[i], v[i+1], v[i+2], v[i+3])
		case []float64:
			gl.Color4d(v[i], v[i+1], v[i+2], v[i+3])
		}
	}
}

// index returns the index at the given offset.
// Used in classic render mode.
func (a *Attr) index(offset int) int {
	switch v := a.data.(type) {
	case []int8:
		return int(v[offset*int(a.size)])
	case []uint8:
		return int(v[offset*int(a.size)])
	case []int16:
		return int(v[offset*int(a.size)])
	case []uint16:
		return int(v[offset*int(a.size)])
	case []int32:
		return int(v[offset*int(a.size)])
	case []uint32:
		return int(v[offset*int(a.size)])
	case []float32:
		return int(v[offset*int(a.size)])
	case []float64:
		return int(v[offset*int(a.size)])
	}

	return 0
}
