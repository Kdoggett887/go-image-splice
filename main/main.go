package main

import (
	splice "go-image-splice"
	"image/jpeg"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		println("not enough args")
		os.Exit(1)
	}
	sourcePath := os.Args[1]
	source, err := os.Open(sourcePath)
	if err != nil {
		println("Error opening scope")
		panic("bah")
	}

	sourceImg, err := jpeg.Decode(source)
	if err != nil {
		println("could not decode source img")
	}

	targetPath := os.Args[2]
	target, err := os.Open(targetPath)
	if err != nil {
		println("Error opening drinks")
		panic("bah")
	}

	targetImg, err := jpeg.Decode(target)
	if err != nil {
		println("could not decode target img")
	}

	bounds := [4][2]int{
		{0, 400},
		{0, 0},
		{600, 400},
		{600, 0},
	}
	t := splice.NewTarget(&targetImg, &bounds)
	s := &splice.Source{Img: &sourceImg}

	newImg := splice.SpliceImgs(s, t)

	output, err := os.Create("result.jpg")
	if err != nil {
		println("couldn't make file")
	}

	jpeg.Encode(output, newImg, &jpeg.Options{jpeg.DefaultQuality})

}
