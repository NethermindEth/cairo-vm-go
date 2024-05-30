package builtins

import mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"

type Poseidon struct{}

func (p *Poseidon) CheckWrite(segment *mem.Segment, offset uint64, value *mem.MemoryValue) error {
	return nil
}

func (p *Poseidon) InferValue(segment *mem.Segment, offset uint64) error {
	return nil
}

func (p *Poseidon) String() string {
	return "poseidon"
}
