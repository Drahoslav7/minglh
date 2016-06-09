// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"unsafe"

	"github.com/go-gl/gl/v3.2-compatibility/gl"
	// "github.com/go-gl/glu"
)

// Sizeof yields the byte size for GL type specified by the given enum.
func Sizeof(gtype uint32) int32 {
	switch gtype {
	case gl.BOOL:
		// var v gl.GLboolean
		// return int32(unsafe.Sizeof(v))
		return 1

	case gl.BYTE:
		// var v gl.GLbyte
		// return int32(unsafe.Sizeof(v))
		return 1

	case gl.UNSIGNED_BYTE:
		// var v gl.GLubyte
		// return int32(unsafe.Sizeof(v))
		return 1

	case gl.SHORT:
		// var v gl.GLshort
		// return int32(unsafe.Sizeof(v))
		return 2

	case gl.UNSIGNED_SHORT:
		// var v gl.GLushort
		// return int32(unsafe.Sizeof(v))
		return 2

	case gl.INT:
		// var v gl.GLint
		// return int32(unsafe.Sizeof(v))
		return 4

	case gl.UNSIGNED_INT:
		// var v gl.GLuint
		// return int32(unsafe.Sizeof(v))
		return 4

	case gl.FLOAT:
		// var v gl.GLfloat
		// return int32(unsafe.Sizeof(v))
		return 4

	case gl.DOUBLE:
		// var v gl.GLdouble
		// return int32(unsafe.Sizeof(v))
		return 8
	}

	panic("Unsupported type")
}

// func ErrorString(error gl.GLenum) (string, error) {
// 	e := unsafe.Pointer(C.gluErrorString(C.GLenum(error)))
// 	if e == nil {
// 		return "", errors.New("Invalid GL error code")
// 	}
// 	return C.GoString((*C.char)(e)), nil
// }

// Used as "defer OpenGLSentinel()()" checks the gl error code on call and exit
func OpenGLSentinel() func() {
	check := func() {
		e := gl.GetError()
		if e != gl.NO_ERROR {
			// s, err := ErrorString(e)
			// if err != nil {
			// 	log.Panic("Invalid error code: ", err)
			// }
			log.Panic("Encountered GLError: ", e, /* " = ", s */)
		}
	}
	check()
	return check
}

// Returns w, h of viewport
func GetViewportWH() (int, int) {
	var viewport [4]int32
	gl.GetIntegerv(gl.VIEWPORT, &viewport[0])
	return int(viewport[2]), int(viewport[3])
}

func GetViewportWHD() (float64, float64) {
	w, h := GetViewportWH()
	return float64(w), float64(h)
}

// // Returns x, y in window co-ordinates at 0 in the z direction
// func WindowToProj(x, y int) (float64, float64) {
// 	var projmat, modelmat [16]float64
// 	var viewport [4]int32

// 	gl.GetDoublev(gl.PROJECTION_MATRIX, projmat[:])
// 	gl.GetDoublev(gl.MODELVIEW_MATRIX, modelmat[:])

// 	gl.GetIntegerv(gl.VIEWPORT, viewport[:])
// 	// Need to convert so that y is at lower left
// 	y = int(viewport[3]) - y

// 	px, py, _ := glu.UnProject(float64(x), float64(y), 0,
// 		&modelmat, &projmat, &viewport)

// 	return px, py
// }

// // Returns x, y in window co-ordinates at 0 in the z direction
// func ProjToWindow(x, y float64) (float64, float64) {
// 	var projmat, modelmat [16]float64
// 	var viewport [4]int32

// 	gl.GetDoublev(gl.PROJECTION_MATRIX, projmat[:])
// 	gl.GetDoublev(gl.MODELVIEW_MATRIX, modelmat[:])
// 	gl.GetIntegerv(gl.VIEWPORT, viewport[:])

// 	px, py, _ := glu.Project(float64(x), float64(y), 0,
// 		&modelmat, &projmat, &viewport)

// 	//return int(px), int(viewport[3]) - int(py)
// 	return px, float64(viewport[3]) - py
// }

// Draws lines of unit length along the X, Y, Z axis in R, G, B
func DrawAxes() {
	gl.Begin(gl.LINES)

	gl.Color3d(1, 0, 0)
	gl.Vertex3d(0, 0, 0)
	gl.Vertex3d(1, 0, 0)

	gl.Color3d(0, 1, 0)
	gl.Vertex3d(0, 0, 0)
	gl.Vertex3d(0, 1, 0)

	gl.Color3d(0, 0, 1)
	gl.Vertex3d(0, 0, 0)
	gl.Vertex3d(0, 0, 1)

	gl.End()
}

// Draws a cross on the screen with known lengths, useful for understanding
// screen dimensions
func DebugLines() {
	gl.MatrixMode(gl.PROJECTION)
	gl.PushMatrix()
	//gl.LoadIdentity()
	//gl.Ortho(-2.1, 6.1, -4, 8, 1, -1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.PushMatrix()
	gl.LoadIdentity()

	gl.LoadIdentity()
	gl.LineWidth(5)
	gl.Color4f(1, 1, 0, 1)
	gl.Begin(gl.LINES)
	gl.Vertex2d(0, -1.6)
	gl.Vertex2d(0, 0.8)
	gl.Vertex2d(-0.8, 0)
	gl.Vertex2d(0.8, 0)
	gl.End()
	gl.PopMatrix()

	gl.MatrixMode(gl.PROJECTION)
	gl.PopMatrix()
	gl.MatrixMode(gl.MODELVIEW)
}

// Emit Vertices of a square with texture co-ordinates which wind anti-clockwise
func Squarei(x, y, w, h int32) {
	var u, v, u2, v2 int32 = 0, 1, 1, 0

	gl.TexCoord2i(u, v)
	gl.Vertex2i(x, y)

	gl.TexCoord2i(u2, v)
	gl.Vertex2i(x+w, y)

	gl.TexCoord2i(u2, v2)
	gl.Vertex2i(x+w, y+h)

	gl.TexCoord2i(u, v2)
	gl.Vertex2i(x, y+h)
}

// Draw a Quad with integer co-ordinates (Using Squarei)
func DrawQuadi(x, y, w, h int32) {
	With(Primitive{gl.QUADS}, func() {
		Squarei(x, y, w, h)
	})
}

// Same as Squarei, double co-ordinates
func Squared(x, y, w, h float64) {
	var u, v, u2, v2 int32 = 0, 1, 1, 0

	gl.TexCoord2i(u, v)
	gl.Vertex2d(x, y)

	gl.TexCoord2i(u2, v)
	gl.Vertex2d(x+w, y)

	gl.TexCoord2i(u2, v2)
	gl.Vertex2d(x+w, y+h)

	gl.TexCoord2i(u, v2)
	gl.Vertex2d(x, y+h)
}

// Same as DrawQuadi, double co-ordinates
func DrawQuadd(x, y, w, h float64) {
	With(Primitive{gl.QUADS}, func() {
		Squared(x, y, w, h)
	})
}

func CaptureRGBA(im *image.RGBA) {
	b := im.Bounds()
	gl.ReadBuffer(gl.BACK_LEFT)
	gl.ReadPixels(0, 0, int32(b.Dx()), int32(b.Dy()), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&im.Pix))
}

// Note: You may want to call ClearAlpha(1) first..
func CaptureToPng(filename string) {
	w, h := GetViewportWH()
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	CaptureRGBA(im)

	fd, err := os.Create(filename)
	if err != nil {
		log.Panic("Err: ", err)
	}
	defer fd.Close()

	png.Encode(fd, im)
}

func ColorC(c color.Color) {
	if c == nil {
		panic("nil color passed to ColorC")
	}
	r, g, b, a := c.RGBA()
	gl.Color4us(uint16(r), uint16(g), uint16(b), uint16(a))
}

// Clear the alpha channel in the color buffer
func ClearAlpha(alpha_value float32) {
	With(Attrib{gl.COLOR_BUFFER_BIT}, func() {
		gl.ColorMask(false, false, false, true)
		gl.ClearColor(0, 0, 0, alpha_value)
		gl.Clear(gl.COLOR_BUFFER_BIT)
	})
}
