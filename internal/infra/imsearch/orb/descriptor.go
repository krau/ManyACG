package orb

import "math"

const DescriptorBytes = 32

const degPerRad = 180.0 / math.Pi

var umax = computeUmax()

func computeUmax() [halfPatchSize + 1]int {
	var u [halfPatchSize + 1]int
	vmax := int(math.Floor(halfPatchSize*math.Sqrt2/2 + 1))
	vmin := int(math.Ceil(halfPatchSize * math.Sqrt2 / 2))
	hp2 := float64(halfPatchSize * halfPatchSize)
	for v := 0; v <= vmax; v++ {
		u[v] = int(math.Round(math.Sqrt(hp2 - float64(v*v))))
	}
	v0 := 0
	for v := halfPatchSize; v >= vmin; v-- {
		for u[v0] == u[v0+1] {
			v0++
		}
		u[v] = v0
		v0++
	}
	return u
}

func icAngle(img *grayImage, cx, cy int) float32 {
	stride := img.stride
	pix := img.pix
	base := cy * stride

	var m01, m10 int
	// center row v=0
	for u := -halfPatchSize; u <= halfPatchSize; u++ {
		m10 += u * int(pix[base+cx+u])
	}
	for v := 1; v <= halfPatchSize; v++ {
		d := umax[v]
		rowPlus := (cy+v)*stride + cx
		rowMinus := (cy-v)*stride + cx
		vSum := 0
		for u := -d; u <= d; u++ {
			plus := int(pix[rowPlus+u])
			minus := int(pix[rowMinus+u])
			vSum += plus - minus
			m10 += u * (plus + minus)
		}
		m01 += v * vSum
	}
	return float32(math.Atan2(float64(m01), float64(m10)) * degPerRad)
}

func computeDescriptor(img *grayImage, cx, cy int, angleDeg float32, desc []byte) {
	angle := float64(angleDeg) / degPerRad
	a := math.Cos(angle)
	b := math.Sin(angle)
	stride := img.stride
	pix := img.pix
	base := cy*stride + cx

	for i := range 32 {
		baseIdx := i * 16
		var val byte
		for j := range 8 {
			p0 := baseIdx + j*2
			p1 := p0 + 1
			px0 := float64(bitPattern31[p0*2])
			py0 := float64(bitPattern31[p0*2+1])
			px1 := float64(bitPattern31[p1*2])
			py1 := float64(bitPattern31[p1*2+1])

			// point 0: y=round(px*b+py*a), x=round(px*a-py*b)
			y0 := int(px0*b + py0*a + 0.5)
			x0 := int(px0*a - py0*b + 0.5)
			t0 := int(pix[base+x0+y0*stride])

			// point 1
			y1 := int(px1*b + py1*a + 0.5)
			x1 := int(px1*a - py1*b + 0.5)
			t1 := int(pix[base+x1+y1*stride])

			if t0 < t1 {
				val |= 1 << uint(j)
			}
		}
		desc[i] = val
	}
}
