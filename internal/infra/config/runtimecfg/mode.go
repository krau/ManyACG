package runtimecfg

import "fmt"

type Mode string

const ModeReader Mode = "reader"

func (m Mode) Valid() error {
	switch m {
	case "", ModeReader:
		return nil
	default:
		return fmt.Errorf("app.mode must be empty or reader, got %q", m)
	}
}

func (m Mode) ReadOnly() bool {
	return m == ModeReader
}
