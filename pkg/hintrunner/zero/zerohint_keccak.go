package zero

import (
	"math"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

func newKeccakWriteArgsHint(inputs, low, high hinter.ResOperander) hinter.Hinter {
	name := "KeccakWriteArgs"
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			// segments.write_arg(ids.inputs, [ids.low % 2 ** 64, ids.low // 2 ** 64])
			// segments.write_arg(ids.inputs + 2, [ids.high % 2 ** 64, ids.high // 2 ** 64])

			low, err := hinter.ResolveAsFelt(vm, low)
			if err != nil {
				return err
			}

			high, err := hinter.ResolveAsFelt(vm, high)
			if err != nil {
				return err
			}

			inputsPtr, err := hinter.ResolveAsAddress(vm, inputs)
			if err != nil {
				return err
			}

			var lowBig big.Int
			var highBig big.Int
			low.BigInt(&lowBig)
			high.BigInt(&highBig)

			var maxUint64Big big.Int
			maxUint64Big = *maxUint64Big.SetUint64(math.MaxUint64)

			lowResultBig := new(big.Int).Set(&lowBig)
			lowResultBigLow := lowResultBig
			lowResultBigLow.And(lowResultBigLow, &maxUint64Big)
			lowResultFeltLow := new(fp.Element).SetBigInt(lowResultBigLow)
			mvLowLow := mem.MemoryValueFromFieldElement(lowResultFeltLow)

			lowResultBigHigh := lowResultBig
			lowResultBigHigh.Rsh(lowResultBigHigh, 64)
			lowResultFeltHigh := new(fp.Element).SetBigInt(lowResultBigHigh)
			mvLowHigh := mem.MemoryValueFromFieldElement(lowResultFeltHigh)

			highResultBig := new(big.Int).Set(&lowBig)
			highResultBigLow := highResultBig
			highResultBigLow.And(highResultBigLow, &maxUint64Big)
			highResultFeltLow := new(fp.Element).SetBigInt(highResultBigLow)
			mvHighLow := mem.MemoryValueFromFieldElement(highResultFeltLow)

			highResulBigHigh := highResultBig
			highResulBigHigh.Rsh(highResulBigHigh, 64)
			highResultFeltHigh := new(fp.Element).SetBigInt(highResulBigHigh)
			mvHighHigh := mem.MemoryValueFromFieldElement(highResultFeltHigh)

			err = vm.Memory.Write(inputsPtr.SegmentIndex, inputsPtr.Offset, &mvLowLow)
			if err != nil {
				return err
			}

			err = vm.Memory.Write(inputsPtr.SegmentIndex, inputsPtr.Offset+1, &mvLowHigh)
			if err != nil {
				return err
			}

			err = vm.Memory.Write(inputsPtr.SegmentIndex, inputsPtr.Offset+2, &mvHighLow)
			if err != nil {
				return err
			}

			err = vm.Memory.Write(inputsPtr.SegmentIndex, inputsPtr.Offset+3, &mvHighHigh)
			if err != nil {
				return err
			}

			return nil
		},
	}
}

func createKeccakWriteArgsHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	inputs, err := resolver.GetResOperander("inputs")
	if inputs != nil {
		return nil, err
	}

	low, err := resolver.GetResOperander("low")
	if low != nil {
		return nil, err
	}

	high, err := resolver.GetResOperander("high")
	if high != nil {
		return nil, err
	}

	return newKeccakWriteArgsHint(inputs, low, high), nil
}
