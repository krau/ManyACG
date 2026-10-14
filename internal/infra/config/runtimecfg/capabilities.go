package runtimecfg

import (
	"fmt"
	"slices"
)

type Capability string

const (
	CapabilityRead     Capability = "read"
	CapabilityWrite    Capability = "write"
	CapabilityFetch    Capability = "fetch"
	CapabilityDownload Capability = "download"
)

type Capabilities []Capability

func (c Capabilities) Validate() error {
	for i, capability := range c {
		switch capability {
		case CapabilityRead, CapabilityWrite, CapabilityFetch, CapabilityDownload:
		default:
			return fmt.Errorf("app.capabilities contains unknown capability %q, valid values are read, write, fetch, download", capability)
		}
		if slices.Contains(c[:i], capability) {
			return fmt.Errorf("app.capabilities contains duplicate capability %q", capability)
		}
	}
	return nil
}

func (c Capabilities) Has(capability Capability) bool {
	return slices.Contains(c, capability)
}
