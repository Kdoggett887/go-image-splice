package main

import (
	splice "go-image-splice"
	"image/gif"
	"image/jpeg"
	"os"
)

func main() {
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
		println("could not decode source img")
	}
	defer source.Close()

	targetPath := os.Args[2]
	target, err := os.Open(targetPath)
	if err != nil {
		println("Error opening target img")
		os.Exit(1)
	}

	targetImg, err := jpeg.Decode(target)
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
	bounds := [4][2]int{
		{334, 378},
		{712, 366},
		{348, 640},
		{730, 581},
	}
	t := splice.NewTarget(&targetImg, &bounds)
	t.SortBounds()
	s := &splice.Source{Img: &sourceImg}

	newImg := splice.Imgs(s, t)

	output, err := os.Create("result.jpg")
	if err != nil {
		println("couldn't make file")
	}
	defer output.Close()

	jpeg.Encode(output, newImg, &jpeg.Options{Quality: jpeg.DefaultQuality})

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
		println(err)
	}
}
