// Copyright 2012 The go-gl Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package glh

import (
	"log"
	"github.com/go-gl/gl/v4.5-core/gl"
)

type Shader struct {
	Type    uint32
	Program string
}

func (s Shader) Compile() uint32 {
	return MakeShader(s.Type, s.Program)
}

func NewProgram(shaders ...Shader) uint32 {
	program := gl.CreateProgram()
	for _, shader := range shaders {
		program.AttachShader(shader.Compile())
	}

	program.Link()
	OpenGLSentinel()

	linkstat := program.Get(gl.LINK_STATUS)
	if linkstat != 1 {
		log.Panic("Program link failed, status=", linkstat,
			"Info log: ", program.GetInfoLog())
	}

	program.Validate()
	valstat := program.Get(gl.VALIDATE_STATUS)
	if valstat != 1 {
		log.Panic("Program validation failed: ", valstat)
	}
	return program
}

func MakeShader(shader_type uint32, source string) uint32 {

	shader := gl.CreateShader(shader_type)
	shader.Source(source)
	shader.Compile()
	OpenGLSentinel()

	compstat := shader.Get(gl.COMPILE_STATUS)
	if compstat != 1 {
		log.Print("vert shader compilation status: ", compstat)
		log.Print("Info log: ", shader.GetInfoLog())
		log.Panic("Problem creating shader?")
	}
	return shader
}
