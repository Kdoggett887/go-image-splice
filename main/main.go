package main

import (
	"fmt"
	splice "go-image-splice"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"runtime"
)

// NProcs is number of go processes
const NProcs = 4

func main() {

	runtime.GOMAXPROCS(NProcs)

	if len(os.Args) < 4 {
		println("not enough args")
		os.Exit(1)
	}
	sourcePath := os.Args[1]
	source, err := os.Open(sourcePath)
	if err != nil {
		println("Error opening scope")
		os.Exit(1)
	}

	sourceImg, err := jpeg.Decode(source)
	if err != nil {
		println("could not decode source img", sourceImg)
	}
	defer source.Close()

	targetPath := os.Args[2]
	target, err := os.Open(targetPath)
	if err != nil {
		println("Error opening target img")
		os.Exit(1)
	}

	targetImg, err := png.Decode(target)
	if err != nil {
		println("could not decode target img")
	}
	defer target.Close()

	// bounds := [4][2]int{
	// 	{134, 178},
	// 	{512, 166},
	// 	{148, 440},
	// 	{530, 381},
	// }

	// target setup
	bounds := [4][2]int{
		{438, 174},
		{863, 311},
		{448, 745},
		{845, 646},
	}
	t := splice.NewTarget(&targetImg, &bounds)
	t.SortBounds()

	// jpeg splice test
	// s := &splice.Source{Img: &sourceImg}

	// newImg := splice.Imgs(s, t)

	// output, err := os.Create("result.jpg")
	// if err != nil {
	// 	println("couldn't make file")
	// }
	// defer output.Close()

	// jpeg.Encode(output, newImg, &jpeg.Options{Quality: jpeg.DefaultQuality})

	// gif splice testing
	gifPath := os.Args[3]
	source2, err := os.Open(gifPath)
	if err != nil {
		println("Error opening source gif")
		os.Exit(1)
	}

	sourceGif, err := gif.DecodeAll(source2)
	if err != nil {
		println("could not decode source gif")
	}
	defer source2.Close()

	newGif := splice.GifToImg(sourceGif, t)

	output2, err := os.Create("gif-result.gif")
	if err != nil {
		println("couldn't make gif")
	}
	defer output2.Close()

	err = gif.EncodeAll(output2, newGif)
	if err != nil {
		println(err.Error())
	}

	// Gif overlay testing
	// println(time.Now().String())
	// newGif = writeGif(sourceGif)
	// println(time.Now().String())

	// output3, err := os.Create("gif-copy.gif")
	// if err != nil {
	// 	println("couldn't make gif file")
	// }
	// defer output3.Close()

	// err = gif.EncodeAll(output3, newGif)
	// if err != nil {
	// 	println(err.Error())
	// }

	// target transparency testing
	// t.AddTransparency()
	// output4, err := os.Create("tronk-trans.png")
	// if err != nil {
	// 	println("couldn't make png file")
	// }
	// defer output4.Close()

	// err = png.Encode(output4, *t.Img)
	// if err != nil {
	// 	println(err.Error())
	// }
}

func writeGif(g *gif.GIF) *gif.GIF {
	fmt.Println(g.Delay, g.Disposal, g.BackgroundIndex)
	firstFrame := g.Image[0]
	for i := range g.Image {
		// tmp image is used here to keep the same dimensions for each frame
		tmp := image.NewNRGBA(firstFrame.Bounds())
		dst := image.NewPaletted(tmp.Bounds(), g.Image[i].Palette)
		draw.FloydSteinberg.Draw(dst, dst.Bounds(), g.Image[i], image.ZP)
		g.Image[i] = dst
	}

	return g
}
