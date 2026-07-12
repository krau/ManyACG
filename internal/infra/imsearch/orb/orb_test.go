package orb

import (
	"image"
	"math/rand"
	"testing"
)

func makeTextured(seed int64, w, h int) *image.NRGBA {
	rng := rand.New(rand.NewSource(seed))
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = 255
	}
	for range 300 {
		x0 := rng.Intn(w)
		y0 := rng.Intn(h)
		bw := rng.Intn(40) + 5
		bh := rng.Intn(40) + 5
		r := byte(rng.Intn(256))
		g := byte(rng.Intn(256))
		bl := byte(rng.Intn(256))
		for y := y0; y < y0+bh && y < h; y++ {
			for x := x0; x < x0+bw && x < w; x++ {
				i := y*img.Stride + x*4
				img.Pix[i], img.Pix[i+1], img.Pix[i+2], img.Pix[i+3] = r, g, bl, 255
			}
		}
	}
	return img
}

func TestUmax(t *testing.T) {
	// umax[0] should be halfPatchSize (15), and values should be non-increasing.
	if umax[0] != halfPatchSize {
		t.Fatalf("umax[0]=%d want %d", umax[0], halfPatchSize)
	}
	for v := 1; v <= halfPatchSize; v++ {
		if umax[v] > umax[v-1] {
			t.Fatalf("umax not monotonic at %d: %d > %d", v, umax[v], umax[v-1])
		}
	}
}

func TestExtractProducesDescriptors(t *testing.T) {
	img := makeTextured(1, 640, 480)
	e := New(Options{NFeatures: 500})
	codes := e.Detect(img)
	if len(codes) == 0 {
		t.Fatal("no descriptors extracted")
	}
	if len(codes) > 500*2 {
		t.Fatalf("far more descriptors than requested: %d", len(codes))
	}
	for _, c := range codes {
		if len(c) != DescriptorBytes {
			t.Fatalf("descriptor size %d want %d", len(c), DescriptorBytes)
		}
	}
	t.Logf("extracted %d descriptors", len(codes))
}

func TestExtractDeterministic(t *testing.T) {
	img := makeTextured(2, 512, 512)
	e := New(Options{NFeatures: 400})
	a := e.Detect(img)
	b := e.Detect(img)
	if len(a) != len(b) {
		t.Fatalf("nondeterministic count: %d vs %d", len(a), len(b))
	}
	// Descriptor sets must be identical (order may differ due to goroutines, so
	// compare as multiset via simple sort of a hash).
	ha := hashCodes(a)
	hb := hashCodes(b)
	for k, v := range ha {
		if hb[k] != v {
			t.Fatalf("descriptor multiset differs at key %x", k)
		}
	}
}

func hashCodes(codes [][]byte) map[[DescriptorBytes]byte]int {
	m := make(map[[DescriptorBytes]byte]int)
	for _, c := range codes {
		var k [DescriptorBytes]byte
		copy(k[:], c)
		m[k]++
	}
	return m
}

func TestDescriptorDistinguishesImages(t *testing.T) {
	// Two different images should produce mostly-different descriptor sets.
	e := New(Options{NFeatures: 300})
	a := e.Detect(makeTextured(10, 480, 480))
	b := e.Detect(makeTextured(20, 480, 480))
	if len(a) == 0 || len(b) == 0 {
		t.Fatal("empty descriptors")
	}
	shared := 0
	hb := hashCodes(b)
	for _, c := range a {
		var k [DescriptorBytes]byte
		copy(k[:], c)
		if hb[k] > 0 {
			shared++
		}
	}
	frac := float64(shared) / float64(len(a))
	if frac > 0.1 {
		t.Fatalf("too many identical descriptors between different images: %.2f", frac)
	}
}

func TestFASTFindsCorner(t *testing.T) {
	// A single bright square on dark background: corners should be detected.
	g := newGray(64, 64)
	for y := 20; y < 44; y++ {
		for x := 20; x < 44; x++ {
			g.pix[y*g.stride+x] = 255
		}
	}
	kps := detectFAST(g, 20, true, make([]int, g.w*g.h))
	if len(kps) == 0 {
		t.Fatal("FAST found no corners on a square")
	}
}
