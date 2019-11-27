package splice

import (
	"image"
	"image/draw"
)

// Target is the target file that we want to splice
// a smaller image into. The bounds respresent the corners
// of the quadrilateral we are splicing onto.
// Bounds are in [x y] pairs.
type Target struct {
	Img    *image.Image
	Bounds *[4][2]int
}

// NewTarget is a factory for Target
func NewTarget(img *image.Image, bounds *[4][2]int) *Target {
	return &Target{
		Img:    img,
		Bounds: bounds,
	}
}

// SortBounds sorts the bounds of a new target to be
// in the order of top left, top right, bottom left,
// bottom right.
// Method for sorting is finding average in each dimension
// and creating axes and putting each point in quadrants.
// Note this will be invalid in the case that the bottom right
// point is higher than the top left point.
// In images 0,0 is top left corner and (h,w) is bottom right
func (t *Target) SortBounds() {
	var xAvg float64
	var yAvg float64
	var newBounds = [4][2]int{}

	for _, point := range t.Bounds {
		xAvg += float64(point[0])
		yAvg += float64(point[1])

	}
	xAvg /= 4
	yAvg /= 4

	for _, point := range t.Bounds {
		x := float64(point[0])
		y := float64(point[1])

		if y > yAvg {
			// either 2 or 3
			if x > xAvg {
				newBounds[3] = point
			} else {
				newBounds[2] = point
			}
		} else {
			// either 0 or 1
			if x > xAvg {
				newBounds[1] = point
			} else {
				newBounds[0] = point
			}
		}
	}

	t.Bounds = &newBounds
}

// BoundsMinMax finds the minX, maxX, minY, maxY
// of a set of bounds and returns them in an array in that
// order
func (t *Target) bbox() [4]int {
	var edges [4]int

	var minY int
	var maxY int
	var minX int
	var maxX int

	for i, pts := range t.Bounds {
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

// AddTransparency converts every pixel within the bounds
// to a transparent pixel. This is done so we can use the
// transparent versions of gifs and utilize the disposal methods
// instead of rendering the background for every frame.
// Currently using a naive method where a minimum bounding box
// is made around the geometry and then each pixel is checked to see
// if it is contained inside the bbox
func (t *Target) AddTransparency() {
	bbox := t.bbox()
	img := image.NewRGBA((*t.Img).Bounds())
	draw.Draw(img, img.Bounds(), *t.Img, image.ZP, draw.Src)

	for x := bbox[0]; x <= bbox[1]; x++ {
		for y := bbox[2]; y <= bbox[3]; y++ {
			if t.ptIn(x, y) {
				img.Set(x, y, image.Transparent)
			}
		}
	}
	newImg := img.SubImage(img.Bounds())
	t.Img = &newImg
}

// ptIn takes x y coords and determines if they are in
// the bounds of target.
// Uses raycasting method to determine intersection counts
// then from there determines if its in.
func (t *Target) ptIn(x, y int) bool {
	// have to make the 4 segments contained in the bounds
	// will go index (0 -> 1, 1 -> 3, 3 -> 2, 2 -> 0)
	segments := [][2][2]int{
		[2][2]int{t.Bounds[0], t.Bounds[1]},
		[2][2]int{t.Bounds[1], t.Bounds[3]},
		[2][2]int{t.Bounds[3], t.Bounds[2]},
		[2][2]int{t.Bounds[2], t.Bounds[0]},
	}
	intersections := 0
	for _, seg := range segments {
		if rayIntersectsSeg(x, y, seg) {
			intersections++
		}
	}

	if intersections > 0 && !(intersections%2 == 0) {
		return true
	}
	return false
}

// checks if x, y extending right intersects the two points
// representing a segment as [[x1,y1], [x2, y2]]
// https://www.ecse.rpi.edu/Homepages/wrf/Research/Short_Notes/pnpoly.html
func rayIntersectsSeg(x, y int, seg [2][2]int) bool {
	return (seg[0][1] > y) != (seg[1][1] > y) &&
		x < (seg[1][0]-seg[0][0])*(y-seg[0][1])/(seg[1][1]-seg[0][1])+seg[0][0]
}
