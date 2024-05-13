package zero

import (
	"math"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// Blake2sAddUint256 hint serializes a `uint256` number in a Blake2s compatible way
//
// `newBlake2sAddUint256Hint` takes 3 operanders as arguments
//   - `low` and `high` are the low and high parts of a `uint256` variable,
//     each of them being a `felt` interpreted as a `uint128`
//   - `data` is a pointer to the starting address in memory where to write the result of the hint
//
// `newBlake2sAddUint256Hint` splits each part of the `uint256` in 4 `u32` and writes the result in memory
// This hint is available in Big-Endian or Little-Endian representation
func newBlake2sAddUint256Hint(low, high, data hinter.ResOperander, bigend bool) hinter.Hinter {
	name := "Blake2sAddUint256"
	if bigend {
		name += "Bigend"
	}
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> B = 32
			//> MASK = 2 ** 32 - 1
			//
			//> non-bigend version
			//> segments.write_arg(ids.data, [(ids.low >> (B * i)) & MASK for i in range(4)])
			//> segments.write_arg(ids.data + 4, [(ids.high >> (B * i)) & MASK for i in range(4)])
			//
			//> bigend version
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

			var shift uint
			var highOffset uint64
			var lowOffset uint64
			for i := uint64(0); i < 4; i++ {
				if bigend {
					shift = uint(b * (3 - i))
					highOffset = dataPtr.Offset + i
					lowOffset = dataPtr.Offset + i + 4
				} else {
					shift = uint(b * i)
					highOffset = dataPtr.Offset + i + 4
					lowOffset = dataPtr.Offset + i
				}

				highResultBig := new(big.Int).Set(&highBig)
				highResultBig.Rsh(highResultBig, shift).And(highResultBig, mask)
				highResultFelt := new(fp.Element).SetBigInt(highResultBig)
				mvHigh := mem.MemoryValueFromFieldElement(highResultFelt)
				err = vm.Memory.Write(dataPtr.SegmentIndex, highOffset, &mvHigh)
				if err != nil {
					return err
				}

				lowResultBig := new(big.Int).Set(&lowBig)
				lowResultBig.Rsh(lowResultBig, shift).And(lowResultBig, mask)
				lowResultFelt := new(fp.Element).SetBigInt(lowResultBig)
				mvLow := mem.MemoryValueFromFieldElement(lowResultFelt)
				err = vm.Memory.Write(dataPtr.SegmentIndex, lowOffset, &mvLow)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createBlake2sAddUint256Hinter(resolver hintReferenceResolver, bigend bool) (hinter.Hinter, error) {
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

	return newBlake2sAddUint256Hint(low, high, data, bigend), nil
}
