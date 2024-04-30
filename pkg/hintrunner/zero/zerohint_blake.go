package zero

import (
	"fmt"
	"math"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

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

//	\\from starkware.cairo.common.cairo_blake2s.blake2s_utils import IV, blake2s_compress
//	\\
//	\\_n_packed_instances = int(ids.N_PACKED_INSTANCES)
//	\\assert 0 <= _n_packed_instances < 20
//	\\_blake2s_input_chunk_size_felts = int(ids.INPUT_BLOCK_FELTS)
//	\\assert 0 <= _blake2s_input_chunk_size_felts < 100
//	\\
//	\\message = [0] * _blake2s_input_chunk_size_felts
//	\\modified_iv = [IV[0] ^ 0x01010020] + IV[1:]
//	\\output = blake2s_compress(
//	\\    message=message,
//	\\    h=modified_iv,
//	\\    t0=0,
//	\\    t1=0,
//	\\    f0=0xffffffff,
//	\\    f1=0,
//	\\)
//	\\padding = (modified_iv + message + [0, 0xffffffff] + output) * (_n_packed_instances - 1)
//	\\segments.write_arg(ids.blake2s_ptr_end, padding)
//
// ;
func newBlake2sFinalizeHint(blake2sPtrEnd, nPackedInstances, inputBlockFelt hinter.ResOperander) hinter.Hinter {
	name := "Blake2sFinalize"
	return &GenericZeroHinter{
		Name: name,
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_blake2s.blake2s_utils import IV, blake2s_compress
			//> _n_packed_instances = int(ids.N_PACKED_INSTANCES)
			//> assert 0 <= _n_packed_instances < 20
			//> _blake2s_input_chunk_size_felts = int(ids.INPUT_BLOCK_FELTS)
			//> assert 0 <= _blake2s_input_chunk_size_felts < 100
			//>
			//> message = [0] * _blake2s_input_chunk_size_felts
			//> modified_iv = [IV[0] ^ 0x01010020] + IV[1:]
			//> output = blake2s_compress(
			//>     message=message,
			//>     h=modified_iv,
			//>     t0=0,
			//>     t1=0,
			//>     f0=0xffffffff,
			//>     f1=0,
			//> )
			//> padding = (modified_iv + message + [0, 0xffffffff] + output) * (_n_packed_instances - 1)
			//> segments.write_arg(ids.blake2s_ptr_end, padding)

			blake2sPtrEnd, err := hinter.ResolveAsAddress(vm, blake2sPtrEnd)
			if err != nil {
				return err
			}
			nPackedInstancesElement, err := hinter.ResolveAsFelt(vm, nPackedInstances)
			if err != nil {
				return err
			}
			nPackedInstances := nPackedInstancesElement.Uint64()

			// assert 0 <= _n_packed_instances < 20
			if nPackedInstances >= 20 {
				return fmt.Errorf("n_packed_instances should be in range [0, 20), got %d", nPackedInstances)
			}

			inputBlockFeltElement, err := hinter.ResolveAsFelt(vm, inputBlockFelt)
			if err != nil {
				return err
			}

			inputBlockFelt := inputBlockFeltElement.Uint64()

			if inputBlockFelt >= 100 {
				return fmt.Errorf("inputBlockFelt should be in range [0, 100), got %d", inputBlockFelt)
			}

			message := make([]uint32, inputBlockFelt)
			modifiedIv := utils.IV()
			modifiedIv[0] = modifiedIv[0] ^ 0x01010020
			output := utils.Blake2sCompress(modifiedIv, message, 0, 0, 0xffffffff, 0)
			padding := modifiedIv[:]
			padding = append(padding, message[:]...)
			padding = append(padding, 0, 0xffffffff)
			padding = append(padding, output[:]...)
			fullPadding := padding
			for i := uint8(1); i < uint8(nPackedInstances); i++ {
				fullPadding = append(fullPadding, padding...)
			}
			for i, val := range fullPadding {
				mv := mem.MemoryValueFromInt(val)
				memWithOffset, err := blake2sPtrEnd.AddOffset(int16(i))
				if err != nil {
					return err
				}
				err = vm.Memory.Write(blake2sPtrEnd.SegmentIndex, memWithOffset.Offset, &mv)
				if err != nil {
					return err
				}
			}
			return err
		},
	}

}

func createBlake2sFinalizeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	blake2sPtrEnd, err := resolver.GetResOperander("blake2s_ptr_end")
	if err != nil {
		return nil, err
	}
	nPackedInstances, err := resolver.GetResOperander("N_PACKED_INSTANCES")
	if err != nil {
		return nil, err
	}
	message, err := resolver.GetResOperander("INPUT_BLOCK_FELTS")
	if err != nil {
		return nil, err
	}

	return newBlake2sFinalizeHint(blake2sPtrEnd, nPackedInstances, message), nil
}
