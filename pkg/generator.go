package pkg

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"io/ioutil"
	"strings"

	"github.com/AndreKR/multiface"
	"github.com/disintegration/imaging"
	"github.com/flopp/go-findfont"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/mod/semver"
)

func ReadFont(inputPath string) (*truetype.Font, error) {
	foundPath, err := findfont.Find(inputPath)
	if err != nil {
		return nil, err
	}

	// load the font with the freetype library
	fontData, err := ioutil.ReadFile(foundPath)
	if err != nil {
		return nil, err
	}
	font, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	return font, nil
}

func CreateFont(fontPaths []string, size float64) (font.Face, error) {
	face := new(multiface.Face)
	opts := &truetype.Options{Size: size, DPI: 72}

	for _, fontPath := range fontPaths {
		font, err := ReadFont(fontPath)
		if err != nil {
			return nil, err
		}
		fc := truetype.NewFace(font, opts)
		face.AddTruetypeFace(fc, font)
	}

	return face, nil
}

// Generate turns a template into an image.
func Generate(template *Template, version string, outputPath string) error {
	dc := gg.NewContext(template.Output.Width, template.Output.Height)

	if len(template.Picasso) > 0 {
		templateVersion := template.Picasso[len(template.Picasso)-1].Version
		if semver.Compare(version, templateVersion) < 0 {
			return fmt.Errorf("a newer version of picasso is needed to render this template: want %s, found %s", templateVersion, version)
		}
	}

	for _, layer := range template.Layers {
		switch layer.Type {
		case "rectangle":
			if strings.HasPrefix(layer.Color, "#") {
				dc.SetHexColor(layer.Color)
			} else if strings.HasPrefix(layer.Color, "rgb(") {
				fmt.Println("rgb is not yet implemented")
			} else if strings.HasPrefix(layer.Color, "rgba(") {
				fmt.Println("rgba is not yet implemented")
			} else if strings.HasPrefix(layer.Color, "linear-gradient(") {
				// TODO: convert degrees/angle into start/stop coordinate
				// g := gg.NewLinearGradient(float64(layer.Color.Start.X), float64(layer.Color.Start.Y), float64(layer.Color.End.X), float64(layer.Color.End.Y))
				// for _, stop := range layer.Color.Stops {
				// 	c, err := ParseHexColor(stop.Value)
				// 	if err != nil {
				// 		return err
				// 	}
				// 	g.AddColorStop(float64(stop.Position), c)
				// }
				// dc.SetFillStyle(g)
				fmt.Println("linear-gradient is not yet implemented")
			} else if strings.HasPrefix(layer.Color, "radial-gradient(") {
				fmt.Println("radial-gradient is not yet implemented")
			} else {
				// add a check for color names here later
				fmt.Println("other colors are not yet implemented")
			}

			dc.DrawRectangle(float64(layer.X), float64(layer.Y), float64(layer.Width), float64(layer.Height))
			dc.Fill()

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

			if strings.HasPrefix(layer.Color, "#") {
				dc.SetHexColor(layer.Color)
			} else {
				dc.SetColor(color.White)
			}

			// if err := dc.LoadFontFace(layer.Font, size); err != nil {
			// 	return err
			// }

			fontPaths := strings.Split(layer.Font, ",")
			face, err := CreateFont(fontPaths, size)
			if err != nil {
				return err
			}

			dc.SetFontFace(face)

			ax := 0.0
			ay := 0.0
			align := gg.AlignLeft

			if layer.Anchor != "" {
				switch layer.Anchor {
				case "CENTER":
					ax = 0.5
					ay = 0.5
					align = gg.AlignCenter
				case "TOP":
					ax = 0.5
					ay = 0.0
					align = gg.AlignCenter
				case "TOP_LEFT":
					ax = 0.0
					ay = 0.0
					align = gg.AlignLeft
				case "TOP_RIGHT":
					ax = 1.0
					ay = 0.0
					align = gg.AlignRight
				case "BOTTOM":
					ax = 0.5
					ay = 1.0
					align = gg.AlignCenter
				case "BOTTOM_LEFT":
					ax = 0.0
					ay = 1.0
					align = gg.AlignLeft
				case "BOTTOM_RIGHT":
					ax = 1.0
					ay = 1.0
					align = gg.AlignRight
				default:
					ax = 0.0
					ay = 0.0
					align = gg.AlignLeft
				}
			}

			if layer.Width != 0 {
				dc.DrawStringWrapped(layer.Content, float64(layer.X), float64(layer.Y)+px, 0, 0, float64(layer.Width), 1.5, align)
			} else {
				dc.DrawStringAnchored(layer.Content, float64(layer.X), float64(layer.Y)+px, ax, ay)
			}
		}
	}

	if err := dc.SavePNG(outputPath); err != nil {
		return err
	}

	return nil
}
