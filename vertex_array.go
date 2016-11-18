package easel

import "github.com/go-gl/gl/v4.1-core/gl"

// VertexArray ...
type VertexArray struct {
	id     uint32
	length uint32
}

func newVertexArray() *VertexArray {
	va := &VertexArray{}
	gl.CreateVertexArrays(1, &va.id)
	return va
}

// Bind ...
func (va *VertexArray) Bind() error {
	gl.BindVertexArray(va.id)
	return checkGLError("Error while binding vertex array")
}

// Unbind ...
func (va *VertexArray) Unbind() {
	gl.BindVertexArray(0)
}

// Length returns the length of this VertexArray
func (va *VertexArray) Length() uint32 {
	return va.length
}