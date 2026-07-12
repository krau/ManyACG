package orb

import "testing"

func BenchmarkDetect640x480(b *testing.B) {
	img := makeTextured(1, 640, 480)
	e := New(Options{NFeatures: 500})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = e.Detect(img)
	}
}
