package splice

import (
	"image"
	"image/draw"

	"github.com/nfnt/resize"
)

// Source is the source image that will be spliced onto
// the Target image
type Source struct {
	Img *image.Image
}

func SpliceImgs(s *Source, t *Target) image.Image {
	// first find the offset we will be using
	// just top left corner of bounds
	pt := t.Bounds[0]
	offset := image.Pt(pt[0], pt[1])
	println(offset.X)

	s.ResizeImg(t.Bounds)

	targImg := *t.Img
	sourceImg := *s.Img
	b := targImg.Bounds()
	newImg := image.NewRGBA(b)

	draw.Draw(newImg, b, targImg, image.ZP, draw.Src)
	draw.Draw(newImg, sourceImg.Bounds().Add(offset), sourceImg, image.ZP, draw.Over)

	return newImg
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
