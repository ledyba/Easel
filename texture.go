package easel

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Texture2D ...
type Texture2D struct {
	texID uint32
}

func newTexture2DFromBytes(data []byte) (*Texture2D, image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, err
	}
	texID, err := loadTexture(img)
	if err != nil {
		return nil, img, err
	}
	return &Texture2D{
		texID: texID,
	}, img, nil
}

func newTexture2DFromImage(img image.Image) (*Texture2D, error) {
	texID, err := loadTexture(img)
	if err != nil {
		return nil, err
	}
	return &Texture2D{
		texID: texID,
	}, nil
}

func (tex *Texture2D) bind() error {
	gl.BindTexture(gl.TEXTURE_2D, tex.texID)
	return checkGLError("Error on binding texture")
}

func (tex *Texture2D) unbind() {
	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// Destroy ...
func (tex *Texture2D) Destroy() {
	gl.DeleteTextures(1, &tex.texID)
}

func loadTexture(img image.Image) (uint32, error) {
	rgba := image.NewRGBA(img.Bounds())
	if rgba.Stride != rgba.Rect.Size().X*4 {
		return 0, fmt.Errorf("unsupported stride")
	}
	draw.Draw(rgba, rgba.Bounds(), img, image.Point{0, 0}, draw.Src)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	return texture, checkGLError("Error on loading image")
}
