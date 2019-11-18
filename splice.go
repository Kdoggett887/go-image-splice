package splice

import (
	"bytes"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"

	"gopkg.in/gographics/imagick.v2/imagick"

	"github.com/nfnt/resize"
)

// Source is the source image that will be spliced onto
// the Target image
type Source struct {
	Img *image.Image
}

// Imgs takes a source and target and draws the source onto the target
// given the bounds in target
func Imgs(s *Source, t *Target) image.Image {
	targImg := *t.Img
	b := targImg.Bounds()

	// have to make source image as big as target
	// so distortion doesn't cut it off
	// TODO: figure out a way to keep aspect ratio while
	// resizing
	edges := &[4][2]int{
		{b.Min.X, b.Min.Y},
		{b.Max.X, b.Min.Y},
		{b.Min.X, b.Max.Y},
		{b.Max.X, b.Max.Y},
	}
	s.ResizeImg(edges) // to full size
	s.TransformPerspective(t.Bounds, b)

	sourceImg := *s.Img
	newImg := image.NewRGBA(b)

	draw.Draw(newImg, b, targImg, image.ZP, draw.Src)
	draw.Draw(newImg, b, sourceImg, image.ZP, draw.Over)

	return newImg
}

// GifToImg takes a source gif and draws each frame onto the target
// then returns the new gif
// Walks through each frame of g and splices that ontop of
// the target. Replaces curr frame with the spliced version
func GifToImg(g *gif.GIF, t *Target) *gif.GIF {
	// gif frames are like a book, you have to paint over the last one
	// so rgbFrame needs to be updated everytime and passed back in
	newGif := &gif.GIF{Delay: g.Delay}
	rgbFrame := image.NewRGBA(g.Image[0].Bounds())
	draw.Draw(rgbFrame, rgbFrame.Bounds(), g.Image[0], image.ZP, draw.Src)

	for _, gifFrame := range g.Image {
		// update rgbFrame, then assign to currFrame which is appended
		// to images
		draw.Draw(rgbFrame, rgbFrame.Bounds(), gifFrame, image.ZP, draw.Over)
		currFrame := *rgbFrame
		currImg := currFrame.SubImage(currFrame.Bounds())

		currSrc := &Source{Img: &currImg}
		newImg := Imgs(currSrc, t)

		// if target is a large image this will be a bottleneck
		newFrame := image.NewPaletted(newImg.Bounds(), palette.Plan9)
		draw.FloydSteinberg.Draw(newFrame, newFrame.Bounds(), newImg, image.ZP)

		newGif.Image = append(newGif.Image, newFrame)
	}
	return newGif
}

// ResizeImg resizes the source image to fit the
// bounds on target
func (s *Source) ResizeImg(bounds *[4][2]int) {
	edges := boundsMinMax(bounds)
	minX, maxX, minY, maxY := edges[0], edges[1], edges[2], edges[3]

	height := uint(maxY - minY)
	width := uint(maxX - minX)

	newImg := resize.Resize(width, height, *s.Img, resize.Lanczos3)
	s.Img = &newImg
}

// TransformPerspective performs a perspective transformation
// on a source image based on the bounds passed in
// our source image will be a rectangle if this is called after
// ResizeImg, so this will warp the perspective to the generic
// quadrilateral bounds
func (s *Source) TransformPerspective(bounds *[4][2]int, targetRect image.Rectangle) {
	img := *s.Img

	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	buf := new(bytes.Buffer)
	err := png.Encode(buf, img)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = mw.ReadImageBlob(buf.Bytes())
	if err != nil {
		fmt.Println(err)
		return
	}

	mw.SetImageVirtualPixelMethod(imagick.VIRTUAL_PIXEL_TRANSPARENT)
	mw.SetImageAlphaChannel(imagick.ALPHA_CHANNEL_SET)

	minX, minY := 0, 0
	maxX := img.Bounds().Dx()
	maxY := img.Bounds().Dy()

	currPts := []image.Point{
		image.Point{minX, minY},
		image.Point{maxX, minY},
		image.Point{minX, maxY},
		image.Point{maxX, maxY},
	}

	newPts := make([]image.Point, 0, 4)
	for _, pt := range bounds {
		newPts = append(newPts, image.Point{pt[0], pt[1]})
	}

	// the input for DistortionImage is a slice of length
	// 16. Where points are grouped as origX1, origY1, newX1, newY1, ...
	pts := make([]float64, 16)
	for i, j := 0, 0; i < len(currPts); i, j = i+1, j+4 {
		origX := float64(currPts[i].X)
		origY := float64(currPts[i].Y)
		newX := float64(newPts[i].X)
		newY := float64(newPts[i].Y)

		pts[j] = origX
		pts[j+1] = origY
		pts[j+2] = newX
		pts[j+3] = newY
	}

	err = mw.DistortImage(imagick.DISTORTION_BILINEAR_FORWARD, pts, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	newImg, _, err := image.Decode(bytes.NewReader(mw.GetImageBlob()))
	if err != nil {
		fmt.Println(err)
		return
	}

	s.Img = &newImg
}

// boundsMinMax finds the minX, maxX, minY, maxY
// of a set of bounds and returns them in an array in that
// order
func boundsMinMax(bounds *[4][2]int) [4]int {
	var edges [4]int

	var minY int
	var maxY int
	var minX int
	var maxX int

	for i, pts := range bounds {
		if i == 0 {
			minX, maxX = pts[0], pts[0]
			minY, maxY = pts[1], pts[1]
		} else {
			if pts[0] < minX {
				minX = pts[0]
			}

			if pts[0] > maxX {
				maxX = pts[0]
			}

			if pts[1] < minY {
				minY = pts[1]
			}

			if pts[1] > maxY {
				maxY = pts[1]
			}
		}
	}

	edges[0] = minX
	edges[1] = maxX
	edges[2] = minY
	edges[3] = maxY
	return edges
}
