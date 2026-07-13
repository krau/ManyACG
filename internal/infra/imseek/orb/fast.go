package orb

type keyPoint struct {
	x, y     float32
	response float32
	angle    float32
}

var fastCircle = [16][2]int{
	{0, -3}, {1, -3}, {2, -2}, {3, -1}, {3, 0}, {3, 1}, {2, 2}, {1, 3},
	{0, 3}, {-1, 3}, {-2, 2}, {-3, 1}, {-3, 0}, {-3, -1}, {-2, -2}, {-1, -3},
}

func hasArc9(mask uint32) bool {
	m := mask | (mask << 16)
	m &= m >> 1
	m &= m >> 2
	m &= m >> 4
	m &= m >> 1
	return m != 0
}

func detectFAST(img *grayImage, threshold int, nonmaxSuppression bool, scores []int) []keyPoint {
	w, h := img.w, img.h
	stride := img.stride
	pix := img.pix

	clear(scores[:w*h])

	var off [16]int
	for i := range fastCircle {
		off[i] = fastCircle[i][1]*stride + fastCircle[i][0]
	}

	var kps []keyPoint

	for y := 3; y < h-3; y++ {
		rowOff := y * stride
		scoreRow := y * w
		for x := 3; x < w-3; x++ {
			center := int(pix[rowOff+x])
			hi := center + threshold
			lo := center - threshold
			p := rowOff + x

			var bc, dc int
			v0 := int(pix[p+off[0]])
			if v0 > hi {
				bc++
			} else if v0 < lo {
				dc++
			}
			v4 := int(pix[p+off[4]])
			if v4 > hi {
				bc++
			} else if v4 < lo {
				dc++
			}
			v8 := int(pix[p+off[8]])
			if v8 > hi {
				bc++
			} else if v8 < lo {
				dc++
			}
			v12 := int(pix[p+off[12]])
			if v12 > hi {
				bc++
			} else if v12 < lo {
				dc++
			}
			if bc < 2 && dc < 2 {
				continue
			}

			var vals [16]int
			var brightMask, darkMask uint32
			vals[0], vals[4], vals[8], vals[12] = v0, v4, v8, v12
			if v0 > hi {
				brightMask |= 1
			} else if v0 < lo {
				darkMask |= 1
			}
			if v4 > hi {
				brightMask |= 1 << 4
			} else if v4 < lo {
				darkMask |= 1 << 4
			}
			if v8 > hi {
				brightMask |= 1 << 8
			} else if v8 < lo {
				darkMask |= 1 << 8
			}
			if v12 > hi {
				brightMask |= 1 << 12
			} else if v12 < lo {
				darkMask |= 1 << 12
			}
			for i := 1; i < 16; i++ {
				if i == 4 || i == 8 || i == 12 {
					continue
				}
				v := int(pix[p+off[i]])
				vals[i] = v
				if v > hi {
					brightMask |= 1 << uint(i)
				} else if v < lo {
					darkMask |= 1 << uint(i)
				}
			}

			if !hasArc9(brightMask) && !hasArc9(darkMask) {
				continue
			}

			s := 0
			for i := range 16 {
				d := vals[i] - center
				if d < 0 {
					d = -d
				}
				s += d
			}

			if nonmaxSuppression {
				scores[scoreRow+x] = s
			} else {
				kps = append(kps, keyPoint{x: float32(x), y: float32(y), response: float32(s)})
			}
		}
	}

	if !nonmaxSuppression {
		return kps
	}

	for y := 3; y < h-3; y++ {
		scoreRow := y * w
		for x := 3; x < w-3; x++ {
			s := scores[scoreRow+x]
			if s == 0 {
				continue
			}
			if s > scores[scoreRow+x-1] && s > scores[scoreRow+x+1] &&
				s > scores[(y-1)*w+x-1] && s > scores[(y-1)*w+x] && s > scores[(y-1)*w+x+1] &&
				s > scores[(y+1)*w+x-1] && s > scores[(y+1)*w+x] && s > scores[(y+1)*w+x+1] {
				kps = append(kps, keyPoint{x: float32(x), y: float32(y), response: float32(s)})
			}
		}
	}
	return kps
}
