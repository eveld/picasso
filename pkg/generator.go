package pkg

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

// Generate turns a template into an image.
func Generate(template *Template, outputPath string) error {
	dc := gg.NewContext(template.Output.Width, template.Output.Height)

	for _, layer := range template.Layers {
		switch layer.Type {
		case "image":
			// Decode the base64 string.
			content, err := base64.StdEncoding.DecodeString(layer.Content)
			if err != nil {
				return err
			}

			// Convert it into an image.
			img, _, err := image.Decode(bytes.NewReader([]byte(content)))
			if err != nil {
				return err
			}

			if layer.Width == 0 && layer.Height == 0 {
				layer.Width = img.Bounds().Dx()
				layer.Height = img.Bounds().Dy()
			}

			resizedImg := imaging.Resize(img, layer.Width, layer.Height, imaging.Lanczos)

			// Draw the image.
			dc.DrawImage(resizedImg, layer.X, layer.Y)
		case "text":
			size := float64(layer.Size)
			px := size

			dc.SetColor(color.White)

			if err := dc.LoadFontFace(layer.Font, size); err != nil {
				return err
			}

			if layer.Width != 0 {
				dc.DrawStringWrapped(layer.Content, float64(layer.X), float64(layer.Y)+px, 0, 0, float64(layer.Width), 1.5, gg.AlignLeft)
			} else {
				dc.DrawString(layer.Content, float64(layer.X), float64(layer.Y)+px)
			}
		}
	}

	if err := dc.SavePNG(outputPath); err != nil {
		return err
	}

	return nil
}
