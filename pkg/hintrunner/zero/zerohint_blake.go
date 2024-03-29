package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
	"math"
	"math/big"
)

func newBlake2sAddUint256BigendHint(low, high, data hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Blake2sAddUint256Bigend",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> B = 32
			//> MASK = 2 ** 32 - 1
			//> segments.write_arg(ids.data, [(ids.high >> (B * (3 - i))) & MASK for i in range(4)])
			//> segments.write_arg(ids.data + 4, [(ids.low >> (B * (3 - i))) & MASK for i in range(4)])

			low, err := hinter.ResolveAsFelt(vm, low)
			if err != nil {
				return err
			}
			high, err := hinter.ResolveAsFelt(vm, high)
			if err != nil {
				return err
			}
			dataPtr, err := hinter.ResolveAsAddress(vm, data)
			if err != nil {
				return err
			}

			var lowBig big.Int
			var highBig big.Int
			low.BigInt(&lowBig)
			high.BigInt(&highBig)

			const b uint64 = 32
			mask := new(big.Int).SetUint64(math.MaxUint32)

			for i := uint64(0); i < 4; i++ {
				shift := uint(b * (3 - i))

				highResultBig := new(big.Int).Set(&highBig)
				highResultBig.Rsh(highResultBig, shift).And(highResultBig, mask)
				highResultFelt := new(fp.Element).SetBigInt(highResultBig)
				mvHigh := mem.MemoryValueFromFieldElement(highResultFelt)
				err = vm.Memory.Write(dataPtr.SegmentIndex, dataPtr.Offset+i, &mvHigh)
				if err != nil {
					return err
				}

				lowResultBig := new(big.Int).Set(&lowBig)
				lowResultBig.Rsh(lowResultBig, shift).And(lowResultBig, mask)
				lowResultFelt := new(fp.Element).SetBigInt(lowResultBig)
				mvLow := mem.MemoryValueFromFieldElement(lowResultFelt)
				err = vm.Memory.Write(dataPtr.SegmentIndex, dataPtr.Offset+i+4, &mvLow)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createBlake2sAddUint256BigendHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	low, err := resolver.GetResOperander("low")
	if err != nil {
		return nil, err
	}
	high, err := resolver.GetResOperander("high")
	if err != nil {
		return nil, err
	}
	data, err := resolver.GetResOperander("data")
	if err != nil {
		return nil, err
	}

	return newBlake2sAddUint256BigendHint(low, high, data), nil
}

func newBlake2sAddUint256Hint(low, high, data hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Blake2sAddUint256",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> B = 32
			//> MASK = 2 ** 32 - 1
			//> segments.write_arg(ids.data, [(ids.low >> (B * i)) & MASK for i in range(4)])
			//> segments.write_arg(ids.data + 4, [(ids.high >> (B * i)) & MASK for i in range(4)])

			low, err := hinter.ResolveAsFelt(vm, low)
			if err != nil {
				return err
			}
			high, err := hinter.ResolveAsFelt(vm, high)
			if err != nil {
				return err
			}
			dataPtr, err := hinter.ResolveAsAddress(vm, data)
			if err != nil {
				return err
			}

			var lowBig big.Int
			var highBig big.Int
			low.BigInt(&lowBig)
			high.BigInt(&highBig)

			const b uint64 = 32
			mask := new(big.Int).SetUint64(math.MaxUint32)

			for i := uint64(0); i < 4; i++ {
				shift := uint(b * i)

				lowResultBig := new(big.Int).Set(&lowBig)
				lowResultBig.Rsh(lowResultBig, shift).And(lowResultBig, mask).Mod(lowResultBig, fp.Modulus())
				lowResultFelt := new(fp.Element).SetBigInt(lowResultBig)
				mvLow := mem.MemoryValueFromFieldElement(lowResultFelt)
				err = vm.Memory.Write(dataPtr.SegmentIndex, dataPtr.Offset+i, &mvLow)
				if err != nil {
					return err
				}

				highResultBig := new(big.Int).Set(&highBig)
				highResultBig.Rsh(highResultBig, shift).And(highResultBig, mask).Mod(highResultBig, fp.Modulus())
				highResultFelt := new(fp.Element).SetBigInt(highResultBig)
				mvHigh := mem.MemoryValueFromFieldElement(highResultFelt)
				err = vm.Memory.Write(dataPtr.SegmentIndex, dataPtr.Offset+i+4, &mvHigh)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createBlake2sAddUint256Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	low, err := resolver.GetResOperander("low")
	if err != nil {
		return nil, err
	}
	high, err := resolver.GetResOperander("high")
	if err != nil {
		return nil, err
	}
	data, err := resolver.GetResOperander("data")
	if err != nil {
		return nil, err
	}

	return newBlake2sAddUint256Hint(low, high, data), nil
}
