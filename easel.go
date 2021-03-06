package easel

import (
	"image"

	log "github.com/Sirupsen/logrus"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

// Easel ...
type Easel struct {
	window *glfw.Window
}

// NewEasel ...
func NewEasel() *Easel {
	glfw.WindowHint(glfw.Visible, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	w, err := glfw.CreateWindow(640, 480, "Easel", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	w.MakeContextCurrent()
	defer glfw.DetachCurrentContext()
	err = gl.Init()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Easel Created.")
	log.Debugf("  ** OpenGL Info **")
	log.Debugf("    OpenGL Version: %s", gl.GoStr(gl.GetString(gl.VERSION)))
	log.Debugf("    GLSL Version:   %s", gl.GoStr(gl.GetString(gl.SHADING_LANGUAGE_VERSION)))
	log.Debugf("    OpenGL Vendor:  %s", gl.GoStr(gl.GetString(gl.VENDOR)))
	log.Debugf("    Renderer:       %s", gl.GoStr(gl.GetString(gl.RENDERER)))
	log.Debugf("    ** Extensions **")
	for i := uint32(0); i < gl.NUM_EXTENSIONS; i++ {
		str := gl.GetStringi(gl.EXTENSIONS, i)
		code := gl.GetError()
		if code == gl.INVALID_VALUE {
			break
		}
		if str != nil {
			log.Debugf("      - %s", gl.GoStr(str))
		}
	}
	return &Easel{
		window: w,
	}
}

// Destroy ...
func (e *Easel) Destroy() {
	e.window.Destroy()
}

// MakeCurrent ...
func (e *Easel) MakeCurrent() {
	e.window.MakeContextCurrent()
}

// DetachCurrent ...
func (e *Easel) DetachCurrent() {
	glfw.DetachCurrentContext()
}

// NewPalette ...
func (e *Easel) NewPalette() (*Palette, error) {
	var err error
	va, err := newVertexArray()
	if err != nil {
		return nil, err
	}
	var fb uint32
	gl.GenFramebuffers(1, &fb)
	err = checkGLError("Error while generating framebuffer")
	if err != nil {
		return nil, err
	}
	p := &Palette{
		easel:         e,
		program:       nil,
		vertexArray:   va,
		frameBufferID: fb,
	}
	return p, nil
}

// SwapBuffers ...
func (e *Easel) SwapBuffers() {
	e.window.SwapBuffers()
}

// CompileProgram ...
func (e *Easel) CompileProgram(vertex, fragment string) (*Program, error) {
	progID, err := compileProgram(vertex, fragment)
	if err != nil {
		return nil, err
	}
	return newProgram(progID), nil
}

// LoadTexture2D ...
func (e *Easel) LoadTexture2D(data []byte) (*Texture2D, image.Image, error) {
	return newTexture2DFromBytes(data)
}

// CreateTexture2D ...
func (e *Easel) CreateTexture2D(img image.Image) (*Texture2D, error) {
	return newTexture2DFromImage(img)
}
