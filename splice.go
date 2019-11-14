package splice

import (
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"

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
	// first find the offset we will be using
	// just top left corner of bounds
	pt := t.Bounds[0]
	offset := image.Pt(pt[0], pt[1])
	s.ResizeImg(t.Bounds)

	targImg := *t.Img
	sourceImg := *s.Img
	b := targImg.Bounds()
	newImg := image.NewRGBA(b)

	draw.Draw(newImg, b, targImg, image.ZP, draw.Src)
	draw.Draw(newImg, sourceImg.Bounds().Add(offset), sourceImg, image.ZP, draw.Over)

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

		skib := &Source{Img: &currImg}

		newImg := Imgs(skib, t)

		newFrame := image.NewPaletted(newImg.Bounds(), palette.Plan9)
		draw.FloydSteinberg.Draw(newFrame, newFrame.Bounds(), newImg, image.ZP)

		// try to figure out why drawing image.Paletted takes so long
		// draw.Draw(newFrame, newFrame.Bounds(), newImg, image.ZP, draw.Src)

		newGif.Image = append(newGif.Image, newFrame)
	}
	return newGif
}

// ResizeImg resizes the source image to fit the
// bounds on target
func (s *Source) ResizeImg(bounds *[4][2]int) {
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

	height := uint(maxY - minY)
	width := uint(maxX - minX)

	newImg := resize.Resize(width, height, *s.Img, resize.Lanczos3)
	s.Img = &newImg
}
