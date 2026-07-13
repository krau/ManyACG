package orb

import (
	"bytes"
	"image"
	"math"
	"os"
	"sync"

	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

type Options struct {
	NFeatures   int     // target keypoints across all levels (default 500)
	ScaleFactor float64 // pyramid scale factor (default 1.2)
	NLevels     int     // pyramid levels (default 8)
	IniThFAST   int     // initial FAST threshold (default 20)
	MinThFAST   int     // fallback FAST threshold (default 7)
	MaxSize     MaxSize // downscale bound before extraction (0 = no bound)
}

type MaxSize struct {
	Height int
	Width  int
}

var DefaultMaxSize = MaxSize{Height: 1080, Width: 768}

func (o *Options) applyDefaults() {
	if o.NFeatures <= 0 {
		o.NFeatures = 500
	}
	if o.ScaleFactor <= 1 {
		o.ScaleFactor = 1.2
	}
	if o.NLevels <= 0 {
		o.NLevels = 8
	}
	if o.IniThFAST <= 0 {
		o.IniThFAST = 20
	}
	if o.MinThFAST <= 0 {
		o.MinThFAST = 7
	}
}

type Extractor struct {
	opts            Options
	scaleFactors    []float64
	invScaleFactors []float64
	featuresPerLvl  []int
	blurBufPool     sync.Pool
}

func New(opts Options) *Extractor {
	opts.applyDefaults()
	e := &Extractor{opts: opts}
	n := opts.NLevels
	e.scaleFactors = make([]float64, n)
	e.invScaleFactors = make([]float64, n)
	e.scaleFactors[0] = 1
	for i := 1; i < n; i++ {
		e.scaleFactors[i] = e.scaleFactors[i-1] * opts.ScaleFactor
	}
	for i := range n {
		e.invScaleFactors[i] = 1.0 / e.scaleFactors[i]
	}
	e.featuresPerLvl = make([]int, n)
	factor := 1.0 / opts.ScaleFactor
	desired := float64(opts.NFeatures) * (1 - factor) / (1 - math.Pow(factor, float64(n)))
	sum := 0
	for level := 0; level < n-1; level++ {
		e.featuresPerLvl[level] = int(math.Round(desired))
		sum += e.featuresPerLvl[level]
		desired *= factor
	}
	last := max(opts.NFeatures-sum, 0)
	e.featuresPerLvl[n-1] = last
	e.blurBufPool = sync.Pool{
		New: func() any { return &grayImage{} },
	}
	return e
}

func (e *Extractor) Close() error { return nil }

func (e *Extractor) getBlurBufs(w, h int) (*grayImage, *grayImage) {
	tmp := e.blurBufPool.Get().(*grayImage)
	blurred := e.blurBufPool.Get().(*grayImage)
	ensureSize(tmp, w, h)
	ensureSize(blurred, w, h)
	return tmp, blurred
}

func (e *Extractor) putBlurBufs(tmp, blurred *grayImage) {
	e.blurBufPool.Put(tmp)
	e.blurBufPool.Put(blurred)
}

func ensureSize(g *grayImage, w, h int) {
	if cap(g.pix) >= w*h {
		g.pix = g.pix[:w*h]
	} else {
		g.pix = make([]byte, w*h)
	}
	g.w = w
	g.h = h
	g.stride = w
}

func (e *Extractor) buildPyramid(g0 *grayImage) []*grayImage {
	levels := make([]*grayImage, e.opts.NLevels)
	levels[0] = g0
	for i := 1; i < e.opts.NLevels; i++ {
		scale := e.invScaleFactors[i]
		w := int(math.Round(float64(g0.w) * scale))
		h := int(math.Round(float64(g0.h) * scale))
		if w < 2*edgeThreshold+2 || h < 2*edgeThreshold+2 {
			levels = levels[:i]
			break
		}
		levels[i] = resizeGray(levels[i-1], w, h)
	}
	return levels
}

func (e *Extractor) detectLevel(img *grayImage, wantN int, scoreBuf []int) []keyPoint {
	const cell = 35
	minB := edgeThreshold - 3
	maxBX := img.w - edgeThreshold + 3
	maxBY := img.h - edgeThreshold + 3
	if maxBX <= minB || maxBY <= minB {
		return nil
	}
	width := maxBX - minB
	height := maxBY - minB
	nCols := width / cell
	nRows := height / cell
	if nCols < 1 {
		nCols = 1
	}
	if nRows < 1 {
		nRows = 1
	}
	wCell := int(math.Ceil(float64(width) / float64(nCols)))
	hCell := int(math.Ceil(float64(height) / float64(nRows)))

	toDistribute := make([]keyPoint, 0, nRows*nCols*4)
	for i := 0; i < nRows; i++ {
		iniY := minB + i*hCell
		maxY := iniY + hCell + 6
		if iniY >= maxBY-3 {
			continue
		}
		if maxY > maxBY {
			maxY = maxBY
		}
		for j := 0; j < nCols; j++ {
			iniX := minB + j*wCell
			maxX := iniX + wCell + 6
			if iniX >= maxBX-6 {
				continue
			}
			if maxX > maxBX {
				maxX = maxBX
			}
			cellKps := detectFASTRegion(img, iniX, iniY, maxX, maxY, e.opts.IniThFAST, scoreBuf)
			if len(cellKps) == 0 {
				cellKps = detectFASTRegion(img, iniX, iniY, maxX, maxY, e.opts.MinThFAST, scoreBuf)
			}
			for _, kp := range cellKps {
				kp.x += float32(iniX)
				kp.y += float32(iniY)
				toDistribute = append(toDistribute, kp)
			}
		}
	}

	kps := distributeOctTree(toDistribute, minB, maxBX, minB, maxBY, wantN)
	return kps
}

func detectFASTRegion(img *grayImage, x0, y0, x1, y1, thr int, scoreBuf []int) []keyPoint {
	sub := &grayImage{
		pix:    img.pix,
		w:      x1 - x0,
		h:      y1 - y0,
		stride: img.stride,
	}
	off := y0*img.stride + x0
	sub.pix = img.pix[off:]
	return detectFAST(sub, thr, true, scoreBuf)
}

func (e *Extractor) Detect(img image.Image) [][]byte {
	g0 := toGray(img)
	return e.detectFromGray(g0)
}

func (e *Extractor) DetectImage(img image.Image) ([][]byte, error) {
	return e.Detect(img), nil
}

func (e *Extractor) DetectBytes(data []byte) ([][]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	g0 := toGray(img)
	return e.detectFromGray(g0), nil
}

func (e *Extractor) detectFromGray(g0 *grayImage) [][]byte {
	if e.opts.MaxSize.Height > 0 && e.opts.MaxSize.Width > 0 {
		if g0.w > e.opts.MaxSize.Width && g0.h > e.opts.MaxSize.Height {
			sw := float64(e.opts.MaxSize.Width) / float64(g0.w)
			sh := float64(e.opts.MaxSize.Height) / float64(g0.h)
			scale := sw
			if sh > scale {
				scale = sh
			}
			nw := int(float64(g0.w) * scale)
			nh := int(float64(g0.h) * scale)
			g0 = resizeGray(g0, nw, nh)
		}
	}
	pyr := e.buildPyramid(g0)

	var allOut [][]byte
	for level := range pyr {
		img := pyr[level]
		want := e.opts.NFeatures
		if level < len(e.featuresPerLvl) {
			want = e.featuresPerLvl[level]
		}
		scoreBuf := make([]int, img.w*img.h)
		kps := e.detectLevel(img, want, scoreBuf)
		if len(kps) == 0 {
			continue
		}
		for i := range kps {
			kps[i].angle = icAngle(img, int(kps[i].x+0.5), int(kps[i].y+0.5))
		}
		tmp, blurred := e.getBlurBufs(img.w, img.h)
		gaussianBlur7(img, tmp, blurred)

		nValid := 0
		for _, kp := range kps {
			cx, cy := int(kp.x+0.5), int(kp.y+0.5)
			if cx >= edgeThreshold && cy >= edgeThreshold &&
				cx < img.w-edgeThreshold && cy < img.h-edgeThreshold {
				nValid++
			}
		}
		flat := make([]byte, nValid*DescriptorBytes)
		out := make([][]byte, 0, nValid)
		idx := 0
		for _, kp := range kps {
			cx, cy := int(kp.x+0.5), int(kp.y+0.5)
			if cx < edgeThreshold || cy < edgeThreshold ||
				cx >= img.w-edgeThreshold || cy >= img.h-edgeThreshold {
				continue
			}
			desc := flat[idx*DescriptorBytes : (idx+1)*DescriptorBytes]
			computeDescriptor(blurred, cx, cy, kp.angle, desc)
			out = append(out, desc)
			idx++
		}
		e.putBlurBufs(tmp, blurred)
		allOut = append(allOut, out...)
	}
	return allOut
}

func (e *Extractor) DetectFile(path string) ([][]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return e.Detect(img), nil
}
