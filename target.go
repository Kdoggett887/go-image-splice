package splice

import "image"

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
			// either 0 or 1
			if x > xAvg {
				newBounds[1] = point
			} else {
				newBounds[0] = point
			}
		} else {
			// either 2 or 3
			if x > xAvg {
				newBounds[3] = point
			} else {
				newBounds[2] = point
			}
		}
	}

	t.Bounds = &newBounds
}
