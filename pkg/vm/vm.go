package vm

type VirtualMachine struct {
	// todo(rodro): define state fields
}

func NewVirtualMachine(program *Program) *VirtualMachine {
	return &VirtualMachine{}
}

func (vm *VirtualMachine) Run() error {
    return nil
}
