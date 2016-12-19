package main

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"os"

	"fmt"
	"github.com/blackjack/webcam"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	vgdraw "github.com/gonum/plot/vg/draw"

	"github.com/gonum/plot/vg/vgimg"
	"github.com/saljam/mjpeg"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
)

type frameSizes []webcam.FrameSize

var (
	p   *plot.Plot
	img *image.RGBA
)

func init() {
	var err error
	p, err = plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "X"
	p.Y.Label.Text = "Y"
	p.BackgroundColor = color.Transparent

}

const (
	listenAddr  = "localhost:8080"
	dpi         = 96
	mjpegFormat = webcam.PixelFormat(1196444237)
)

func main() {

	fmt.Printf("go to http://%v/camera\n", listenAddr)
	stream := mjpeg.NewStream()
	go updateJpeg(stream)
	http.Handle("/camera", stream)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}

func initWebcam() (*webcam.Webcam, error) {
	var cam *webcam.Webcam
	var err error
	cam, err = webcam.Open("/dev/video0")
	if err != nil {
		return nil, err
	}
	// Check if the webcam supports MJPEG
	formatDesc := cam.GetSupportedFormats()
	if _, ok := formatDesc[mjpegFormat]; !ok {
		return nil, fmt.Errorf("Webcam does not support MJPEG")
	}
	frames := frameSizes(cam.GetSupportedFrameSizes(mjpegFormat))
	log.Println(frames)

	size := frames[2]

	_, w, h, err := cam.SetImageFormat(mjpegFormat, uint32(size.MaxWidth), uint32(size.MaxHeight))

	if err != nil {
		return nil, err
	}
	fmt.Fprintf(os.Stderr, "Resulting image format: (%dx%d)\n", w, h)

	return cam, nil
}

func updateJpeg(s *mjpeg.Stream) {
	cam, err := initWebcam()
	if err != nil {
		panic(err)
	}
	defer cam.Close()
	err = cam.StartStreaming()
	if err != nil {
		log.Fatal(err)
	}
	for {
		timeout := uint32(5) //5 seconds
		err = cam.WaitForFrame(timeout)

		switch err.(type) {
		case nil:
		case *webcam.Timeout:
			fmt.Fprint(os.Stderr, err.Error())
		default:
			panic(err.Error())
		}
		frame, err := cam.ReadFrame()
		if err == nil {
			f, err := processFrame(frame)
			if err == nil {
				s.UpdateJPEG(f)
			}
		}
	}
}

func processFrame(frame []byte) ([]byte, error) {
	// Convert the frame into an image
	r := bytes.NewReader(frame)
	img, err := jpeg.Decode(r)
	if err != nil {
		return nil, err
	}
	imgB := img.Bounds()
	// Create the desination image as a image.RGBA object
	finalDst := image.NewRGBA(imgB)

	// Copy the photo in the finalDst
	draw.Draw(finalDst, finalDst.Bounds(), img, imgB.Min, draw.Over)
	draw.Draw(finalDst, finalDst.Bounds(), getPlot(), imgB.Min, draw.Over)

	out := new(bytes.Buffer)
	jpeg.Encode(out, finalDst, &jpeg.Options{jpeg.DefaultQuality})
	return out.Bytes(), nil
}

func getPlot() image.Image {
	// Graph
	err := plotutil.AddLinePoints(p,
		"First", randomPoints(15),
		"Second", randomPoints(15),
		"Third", randomPoints(15))
	if err != nil {
		panic(err)
	}
	if err != nil {
		panic(err)
	}

	// Draw the plot to an in-memory image.
	img = image.NewRGBA(image.Rect(0, 0, 5*dpi, 3*dpi))
	c := vgimg.NewWith(vgimg.UseImage(img))
	p.Draw(vgdraw.New(c))
	return img
}

// randomPoints returns some random x, y points.
func randomPoints(n int) plotter.XYs {
	pts := make(plotter.XYs, n)
	for i := range pts {
		if i == 0 {
			pts[i].X = rand.Float64()
		} else {
			pts[i].X = pts[i-1].X + rand.Float64()
		}
		pts[i].Y = pts[i].X + 10*rand.Float64()
	}
	return pts
}
