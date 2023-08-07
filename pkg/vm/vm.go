package vm

import f "github.com/NethermindEth/juno/core/felt"

type VirtualMachine struct {
	Ap, Fp, Pc Relocatable
	Mem        map[Relocatable]Relocatable
}

func NewVirtualMachine(bytecode *[]f.Felt) *VirtualMachine {
	mem := map[Relocatable]Relocatable{}

	for pc, instr := range *bytecode {
		mem[*NewRelocatable(1, new(f.Felt).SetUint64(uint64(pc)))] = *new(Relocatable).SetFelt(&instr)
	}

	return &VirtualMachine{
		Ap:  *NewRelocatable(2, &f.Zero),
		Fp:  *NewRelocatable(2, &f.Zero),
		Pc:  *NewRelocatable(1, &f.Zero),
		Mem: mem,
	}
}

func (vm *VirtualMachine) Run() error {
	return nil
}
