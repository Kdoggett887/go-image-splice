package splice

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/fogleman/gg"
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
	// add transparency to target so we can safely write over it
	t.AddTransparency()
	fmt.Println(g.Delay, g.Disposal, g.BackgroundIndex)
	newGif := &gif.GIF{
		Image:           g.Image,
		Delay:           g.Delay,
		Disposal:        g.Disposal,
		BackgroundIndex: g.BackgroundIndex,
	}

	// setup color palette
	var gifPalette color.Palette
	gifPalette = palette.WebSafe
	gifPalette = append(gifPalette, image.Transparent)

	proc := make(chan bool)
	for i, gifFrame := range g.Image {
		go func(i int, gifFrame *image.Paletted) {
			currFrame := image.NewRGBA(gifFrame.Bounds())
			draw.Draw(currFrame, currFrame.Bounds(), gifFrame, image.ZP, draw.Src)

			currImg := currFrame.SubImage(currFrame.Bounds())
			currSrc := &Source{Img: &currImg}
			newImg := Imgs(currSrc, t)

			// if target is a large image this will be a bottleneck
			newFrame := image.NewPaletted(newImg.Bounds(), gifPalette)
			draw.FloydSteinberg.Draw(newFrame, newFrame.Bounds(), newImg, image.ZP)
			newGif.Image[i] = newFrame
			proc <- true
		}(i, gifFrame)
	}

	for i := 0; i < len(g.Image); i++ {
		<-proc
	}

	return newGif
}

// GifToFrames splices the gif onto the target and saves
// the generated images to a timestamped folder that is a child of
// the dir. Frames are named as frame-%04d.png (ex. frame-0001.png).
// Returns path to pngdir or error.
func GifToFrames(g *gif.GIF, t *Target, dir string) (string, error) {
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	pngdir := dir + timestamp
	err := os.MkdirAll(pngdir, os.ModePerm)
	if err != nil {
		fmt.Println("error creating dir", err.Error())
	}

	fmt.Println("frames to process:", len(g.Image))

	firstFrame := g.Image[0]
	currFrame := image.NewNRGBA64(firstFrame.Bounds())
	draw.Draw(currFrame, currFrame.Bounds(), g.Image[0], image.ZP, draw.Src)

	for i, gifImage := range g.Image {
		fmt.Println("starting frame:", i)
		draw.Draw(currFrame, currFrame.Bounds(), gifImage, image.ZP, draw.Over)
		frame := *currFrame

		currImg := currFrame.SubImage(frame.Bounds())
		currSrc := &Source{Img: &currImg}
		newImg := Imgs(currSrc, t)

		ctx := gg.NewContextForImage(newImg)
		err = ctx.SavePNG(fmt.Sprintf("%s%s%04d%s", pngdir, "/frame-", i, ".png"))

		if err != nil {
			fmt.Println("error saving png", err.Error())
			return pngdir, err
		}
	}

	return pngdir, nil
}

// GifToPngSlice takes a gif and encodes it into a list of PNG buffers
func GifToPngSlice(g *gif.GIF, t *Target) ([]*bytes.Buffer, error) {
	fmt.Println("frames to process:", len(g.Image))
	frames := make([]*bytes.Buffer, len(g.Image))

	firstFrame := g.Image[0]
	currFrame := image.NewNRGBA64(firstFrame.Bounds())
	draw.Draw(currFrame, currFrame.Bounds(), g.Image[0], image.ZP, draw.Src)

	for i, gifImage := range g.Image {
		fmt.Println("starting frame:", i)
		draw.Draw(currFrame, currFrame.Bounds(), gifImage, image.ZP, draw.Over)
		frame := *currFrame

		currImg := currFrame.SubImage(frame.Bounds())
		currSrc := &Source{Img: &currImg}
		newImg := Imgs(currSrc, t)

		buf := new(bytes.Buffer)
		err := png.Encode(buf, newImg)
		if err != nil {
			return nil, err
		}

		frames[i] = buf
	}

	return frames, nil
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
	ptStr := boundaryString(s.Img, bounds)

	cmd := exec.Command("convert", "-", "-virtual-pixel", "transparent",
		"-alpha", "set", "-distort", "BilinearForward", ptStr, "-")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("err on stdin", err.Error())
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("stdout err", err.Error())
	}

	// read image in once we exec
	go func() {
		defer stdin.Close()

		inBuf := new(bytes.Buffer)
		err = png.Encode(inBuf, img)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = stdin.Write(inBuf.Bytes())
		if err != nil {
			fmt.Println("error writing buf to stdin", err.Error())
			return
		}
	}()

	err = cmd.Start()
	if err != nil {
		fmt.Println("error while running", err.Error())
		return
	}

	outBuf := new(bytes.Buffer)
	_, err = outBuf.ReadFrom(stdout)
	if err != nil {
		fmt.Println("error reading stdout", err.Error())
		return
	}

	err = cmd.Wait()
	if err != nil {
		fmt.Println("err at wait", err.Error())
		return
	}

	newImg, _, err := image.Decode(bytes.NewReader(outBuf.Bytes()))
	if err != nil {
		fmt.Println("error while decoding image", err.Error())
		return
	}

	s.Img = &newImg
}

func boundaryString(img *image.Image, bounds *[4][2]int) string {
	minX, minY := 0, 0
	maxX := (*img).Bounds().Dx()
	maxY := (*img).Bounds().Dy()

	currPts := []image.Point{
		image.Point{minX, minY},
		image.Point{maxX, minY},
		image.Point{minX, maxY},
		image.Point{maxX, maxY},
	}

	// the input for DistortionImage is a slice of length
	// 16. Where points are grouped as origX1, origY1, newX1, newY1, ...
	pts := make([]string, 16)
	for i, j := 0, 0; i < len(currPts); i, j = i+1, j+4 {
		origX := strconv.Itoa(currPts[i].X)
		origY := strconv.Itoa(currPts[i].Y)
		newX := strconv.Itoa((*bounds)[i][0])
		newY := strconv.Itoa((*bounds)[i][1])

		pts[j] = origX
		pts[j+1] = origY
		pts[j+2] = newX
		pts[j+3] = newY
	}

	ptStr := strings.Join(pts, ",")
	return ptStr
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
