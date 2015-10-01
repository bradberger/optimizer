package optimizer

import (
	"fmt"
	"github.com/bradberger/resize"
	"github.com/bradberger/webp"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
)

// Options provides a Client-Hinting compatible set of options for image encoding.
type Options struct {
	Mime          string
	Width         uint
	Height        uint
	Dpr           float64
	Quality       int
	Downlink      float64
	ViewportWidth float64
	SaveData      bool
	Interpolation resize.InterpolationFunction
	Optimized     bool
}

func (o *Options) Optimize() {

	if !o.Optimized {

		if o.Width > 0 {
			o.Width = uint(float64(o.Width) * o.Dpr)
		}

		if o.Dpr == 0 {
			o.Dpr = 1
		}

		// If quality not explicity set, we'll try to optimize it.
		if o.Quality == 0 {

			o.Quality = int(100 - o.Dpr*30)

			if o.Downlink > 0 && o.Downlink < 1 {
				o.Quality = int(float32(o.Quality) * float32(o.Downlink))
			}

			if o.SaveData {
				o.Quality = int(float32(o.Quality) * float32(0.75))
			}

		}

		o.Optimized = true

	}

}

// Encode encodes the image with the given options.
func Encode(w io.Writer, i image.Image, o Options) error {

	o.Optimize()

	if o.Width > 0 {
		i = resize.Resize(uint(float64(o.Width)*o.Dpr), o.Height, i, resize.Bicubic)
	}

	// Now write the result.
	switch {
	case o.Mime == "image/jpeg":
		jpeg.Encode(w, i, &jpeg.Options{Quality: o.Quality})
	case o.Mime == "image/png":
		png.Encode(w, i)
	case o.Mime == "image/webp":
		webp.Encode(w, i, &webp.Options{Quality: float32(o.Quality)})
	case o.Mime == "image/gif":
		gif.Encode(w, i, nil)
	default:
		return fmt.Errorf("Format %s is not supported", o.Mime)
	}

	return nil

}
