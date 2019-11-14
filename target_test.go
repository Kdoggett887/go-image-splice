package splice

import (
	"testing"
)

func permutations(bounds *[4][2]int) [][4][2]int {
	var helper func([4][2]int, int)
	perms := [][4][2]int{}

	helper = func(arr [4][2]int, n int) {
		if n == 1 {
			var tmp [4][2]int

			copy(tmp[:], arr[:])
			perms = append(perms, tmp)
		} else {
			for i := 0; i < n; i++ {
				helper(arr, n-1)
				if n%2 == 1 {
					tmp := arr[i]
					arr[i] = arr[n-1]
					arr[n-1] = tmp
				} else {
					tmp := arr[0]
					arr[0] = arr[n-1]
					arr[n-1] = tmp
				}
			}
		}
	}

	helper(*bounds, len(*bounds))

	return perms
}

func TestSortBounds(t *testing.T) {
	points := &[4][2]int{
		{0, 1},
		{1, 1},
		{0, 0},
		{1, 0},
	}

	perms := permutations(points)

	for _, bounds := range perms {
		target := NewTarget(nil, &bounds)
		target.SortBounds()

		if *target.Bounds != *points {
			t.Error("Bounds not properly sorted")
		}
	}
}
