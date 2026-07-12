package orb

import (
	"image"
	"math"
)

const (
	// halfPatchSize is half of the ORB circular patch (full patch is 31x31).
	// Intensity-centroid angle and BRIEF pattern are built around this radius.
	halfPatchSize = 15
	edgeThreshold = 19
)

type grayImage struct {
	pix    []byte
	w, h   int
	stride int
}

func newGray(w, h int) *grayImage {
	return &grayImage{pix: make([]byte, w*h), w: w, h: h, stride: w}
}

func toGray(img image.Image) *grayImage {
	b := img.Bounds()
	w, hgt := b.Dx(), b.Dy()
	g := newGray(w, hgt)
	switch src := img.(type) {
	case *image.YCbCr:
		// webp decoder returns YCbCr — use Y channel directly when 4:4:4
		if src.SubsampleRatio == image.YCbCrSubsampleRatio444 {
			for y := range hgt {
				copy(g.pix[y*w:(y+1)*w], src.Y[y*src.YStride:y*src.YStride+w])
			}
		} else {
			// 4:2:0 or 4:2:2 — Cb/Cr are subsampled; use the Y channel
			// directly as luminance. This is exact for grayscale purposes
			// since Y IS the luma component of YCbCr.
			for y := range hgt {
				copy(g.pix[y*w:(y+1)*w], src.Y[y*src.YStride:y*src.YStride+w])
			}
		}
	case *image.NRGBA:
		for y := range hgt {
			for x := range w {
				i := y*src.Stride + x*4
				g.pix[y*w+x] = luma(src.Pix[i], src.Pix[i+1], src.Pix[i+2])
			}
		}
	case *image.RGBA:
		for y := range hgt {
			for x := range w {
				i := y*src.Stride + x*4
				g.pix[y*w+x] = luma(src.Pix[i], src.Pix[i+1], src.Pix[i+2])
			}
		}
	case *image.Gray:
		for y := range hgt {
			copy(g.pix[y*w:(y+1)*w], src.Pix[y*src.Stride:y*src.Stride+w])
		}
	default:
		for y := range hgt {
			for x := range w {
				r, gg, bb, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
				g.pix[y*w+x] = luma(byte(r>>8), byte(gg>>8), byte(bb>>8))
			}
		}
	}
	return g
}

func luma(r, g, b byte) byte {
	return byte((19595*uint32(r) + 38470*uint32(g) + 7471*uint32(b) + 1<<15) >> 16)
}

func resizeGray(src *grayImage, w, h int) *grayImage {
	dst := newGray(w, h)
	if w == 0 || h == 0 {
		return dst
	}
	scaleX := float64(src.w) / float64(w)
	scaleY := float64(src.h) / float64(h)

	// Precompute y mapping: source y0, y1, and fixed-point weight for each dst row.
	y0Tab := make([]int, h)
	y1Tab := make([]int, h)
	fyTab := make([]int, h)
	for dy := range h {
		sy := (float64(dy)+0.5)*scaleY - 0.5
		y0 := int(math.Floor(sy))
		fyTab[dy] = int((sy - float64(y0)) * 256)
		y1 := y0 + 1
		if y0 < 0 {
			y0 = 0
		}
		if y1 < 0 {
			y1 = 0
		}
		if y0 >= src.h {
			y0 = src.h - 1
		}
		if y1 >= src.h {
			y1 = src.h - 1
		}
		y0Tab[dy] = y0
		y1Tab[dy] = y1
	}

	// Precompute x mapping.
	x0Tab := make([]int, w)
	x1Tab := make([]int, w)
	fxTab := make([]int, w)
	for dx := range w {
		sx := (float64(dx)+0.5)*scaleX - 0.5
		x0 := int(math.Floor(sx))
		fxTab[dx] = int((sx - float64(x0)) * 256)
		x1 := x0 + 1
		if x0 < 0 {
			x0 = 0
		}
		if x1 < 0 {
			x1 = 0
		}
		if x0 >= src.w {
			x0 = src.w - 1
		}
		if x1 >= src.w {
			x1 = src.w - 1
		}
		x0Tab[dx] = x0
		x1Tab[dx] = x1
	}

	for dy := range h {
		y0 := y0Tab[dy]
		y1 := y1Tab[dy]
		fy := fyTab[dy]
		row0 := src.pix[y0*src.stride:]
		row1 := src.pix[y1*src.stride:]
		outRow := dst.pix[dy*dst.stride : dy*dst.stride+w]
		for dx := range w {
			x0 := x0Tab[dx]
			x1 := x1Tab[dx]
			fx := fxTab[dx]
			p00 := int(row0[x0])
			p01 := int(row0[x1])
			p10 := int(row1[x0])
			p11 := int(row1[x1])
			top := p00 + (p01-p00)*fx>>8
			bot := p10 + (p11-p10)*fx>>8
			outRow[dx] = byte(top + (bot-top)*fy>>8)
		}
	}
	return dst
}

var gaussianKernel = func() [7]int {
	sigma := 2.0
	var k [7]float64
	var sum float64
	for i := -3; i <= 3; i++ {
		v := math.Exp(-float64(i*i) / (2 * sigma * sigma))
		k[i+3] = v
		sum += v
	}
	var ki [7]int
	for i := range k {
		ki[i] = int(k[i]/sum*256 + 0.5)
	}
	return ki
}()

func gaussianBlur7(src *grayImage, tmp, dst *grayImage) {
	w, h := src.w, src.h
	k := gaussianKernel
	k0, k1, k2, k3 := k[0], k[1], k[2], k[3]

	// horizontal pass: src -> tmp
	for y := range h {
		row := src.pix[y*src.stride : y*src.stride+w]
		out := tmp.pix[y*tmp.stride : y*tmp.stride+w]
		for x := 3; x < w-3; x++ {
			out[x] = byte((k3*int(row[x]) +
				k2*(int(row[x-1])+int(row[x+1])) +
				k1*(int(row[x-2])+int(row[x+2])) +
				k0*(int(row[x-3])+int(row[x+3]))) >> 8)
		}
		for x := range 3 {
			out[x] = blurAt1DInt(row, x, w, k)
		}
		for x := w - 3; x < w; x++ {
			out[x] = blurAt1DInt(row, x, w, k)
		}
	}

	// vertical pass: tmp -> dst
	st := tmp.stride
	for y := 3; y < h-3; y++ {
		outRow := dst.pix[y*dst.stride : y*dst.stride+w]
		r0 := tmp.pix[(y-3)*st:]
		r1 := tmp.pix[(y-2)*st:]
		r2 := tmp.pix[(y-1)*st:]
		rc := tmp.pix[y*st:]
		r4 := tmp.pix[(y+1)*st:]
		r5 := tmp.pix[(y+2)*st:]
		r6 := tmp.pix[(y+3)*st:]
		for x := range w {
			outRow[x] = byte((k3*int(rc[x]) +
				k2*(int(r2[x])+int(r4[x])) +
				k1*(int(r1[x])+int(r5[x])) +
				k0*(int(r0[x])+int(r6[x]))) >> 8)
		}
	}
	for y := range 3 {
		blurRowVertInt(tmp, dst, y, h, k)
	}
	for y := h - 3; y < h; y++ {
		blurRowVertInt(tmp, dst, y, h, k)
	}
}

func blurAt1DInt(row []byte, x, w int, k [7]int) byte {
	reflect := func(i int) int {
		if w == 1 {
			return 0
		}
		for i < 0 || i >= w {
			if i < 0 {
				i = -i
			}
			if i >= w {
				i = 2*(w-1) - i
			}
		}
		return i
	}
	acc := k[3]*int(row[reflect(x)]) +
		k[2]*(int(row[reflect(x-1)])+int(row[reflect(x+1)])) +
		k[1]*(int(row[reflect(x-2)])+int(row[reflect(x+2)])) +
		k[0]*(int(row[reflect(x-3)])+int(row[reflect(x+3)]))
	acc >>= 8
	if acc < 0 {
		acc = 0
	}
	if acc > 255 {
		acc = 255
	}
	return byte(acc)
}

func blurRowVertInt(tmp, dst *grayImage, y, h int, k [7]int) {
	w := tmp.w
	reflect := func(i int) int {
		if h == 1 {
			return 0
		}
		for i < 0 || i >= h {
			if i < 0 {
				i = -i
			}
			if i >= h {
				i = 2*(h-1) - i
			}
		}
		return i
	}
	st := tmp.stride
	out := dst.pix[y*dst.stride : y*dst.stride+w]
	for x := range w {
		acc := k[3]*int(tmp.pix[reflect(y)*st+x]) +
			k[2]*(int(tmp.pix[reflect(y-1)*st+x])+int(tmp.pix[reflect(y+1)*st+x])) +
			k[1]*(int(tmp.pix[reflect(y-2)*st+x])+int(tmp.pix[reflect(y+2)*st+x])) +
			k[0]*(int(tmp.pix[reflect(y-3)*st+x])+int(tmp.pix[reflect(y+3)*st+x]))
		acc >>= 8
		if acc < 0 {
			acc = 0
		}
		if acc > 255 {
			acc = 255
		}
		out[x] = byte(acc)
	}
}
