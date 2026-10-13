package runtimecfg

import "fmt"

type Mode string

const (
	ModeStandalone Mode = "standalone"
	ModeWriter     Mode = "writer"
	ModeReader     Mode = "reader"
)

func (m Mode) Valid() error {
	switch m {
	case "", ModeStandalone, ModeWriter, ModeReader:
		return nil
	default:
		return fmt.Errorf("app.mode must be standalone, writer or reader, got %q", m)
	}
}

func (m Mode) ReadOnly() bool {
	return m == ModeReader
}

func (m Mode) Standalone() bool {
	return m == "" || m == ModeStandalone
}
