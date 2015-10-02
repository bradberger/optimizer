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
	"net/http"
	"strconv"
	"strings"
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

// SetFromRequest sets the options based on headers/parameters in the http request
func (o *Options) SetFromRequest(r *http.Request) {

	// Get the mime type.
	if strings.Contains(r.Header.Get("Accept"), "image/webp") {
		o.Mime = "image/webp"
	}

	// Get the DPR
	dpr, err := strconv.ParseFloat(r.Header.Get("DPR"), 64)
	if err != nil {
		dpr, err = strconv.ParseFloat(r.FormValue("dpr"), 64)
	}
	if err != nil {
		dpr = 1.0
	}
	o.Dpr = dpr

	// Set SaveData flag
	if r.Header.Get("Save-Data") == "1" || r.FormValue("save-data") == "1" {
		o.SaveData = true
	} else {
		o.SaveData = false
	}

	// Get the Viewport Width
	viewport, err := strconv.ParseFloat(r.Header.Get("Viewport-Width"), 64)
	if err != nil {
		viewport, err = strconv.ParseFloat(r.FormValue("viewport-width"), 64)
	}
	o.ViewportWidth = viewport

	// Set the image width.
	width, err := strconv.Atoi(r.Header.Get("Width"))
	if err != nil {
		width, _ = strconv.Atoi(r.FormValue("width"))
	}
	if width > 0 {
		o.Width = uint(width)
	}

	// Set Downlink
	downlink, err := strconv.ParseFloat(r.Header.Get("Downlink"), 64)
	if err != nil {
		downlink, err = strconv.ParseFloat(r.FormValue("downlink"), 64)
	}
	if err != nil {
		downlink = 0
	}
	o.Downlink = downlink

}

// Optimize sets the Quality dependent on various factors
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
