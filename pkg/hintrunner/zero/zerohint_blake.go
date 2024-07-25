package zero

import (
	"math"
	"math/big"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	hintrunnerUtils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
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

// Blake2sFinalize hint finalizes the Blake2s hash computation, ie it verifies
// that the results of blake2s() are valid
//
// `newBlake2sFinalizeHint` takes 1 operander as argument
//   - `blake2sPtrEnd` is a pointer to the address where to write the result
//
// There are 3 versions of Blake2sFinalize hint, this implementation corresponds to V1 and V2
func newBlake2sFinalizeHint(blake2sPtrEnd hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Blake2sFinalize",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> # Add dummy pairs of input and output.
			//> from starkware.cairo.common.cairo_blake2s.blake2s_utils import IV, blake2s_compress
			//> _n_packed_instances = int(ids.N_PACKED_INSTANCES)
			//> assert 0 <= _n_packed_instances < 20
			//> _blake2s_input_chunk_size_felts = int(ids.INPUT_BLOCK_FELTS) (V1) //> _blake2s_input_chunk_size_felts = int(ids.BLAKE2S_INPUT_CHUNK_SIZE_FELTS) (V2)
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

			//> assert 0 <= _n_packed_instances < 20
			// as N_PACKED_INSTANCES is a constant of 7, this can be skipped

			//> assert 0 <= _blake2s_input_chunk_size_felts < 100
			// as INPUT_BLOCK_FELTS (or BLAKE2S_INPUT_CHUNK_SIZE_FELTS) is a constant of 16, this can be skipped

			message := make([]uint32, utils.INPUT_BLOCK_FELTS)
			modifiedIv := utils.IV()
			modifiedIv[0] = modifiedIv[0] ^ 0x01010020
			output := utils.Blake2sCompress(message, modifiedIv, 0, 0, 0xffffffff, 0)
			padding := modifiedIv[:]
			padding = append(padding, message[:]...)
			padding = append(padding, 0, 0xffffffff)
			padding = append(padding, output[:]...)
			fullPadding := []uint32{}
			for i := uint64(0); i < utils.N_PACKED_INSTANCES-1; i++ {
				fullPadding = append(fullPadding, padding...)
			}

			for _, val := range fullPadding {
				mv := mem.MemoryValueFromInt(val)
				if err != nil {
					return err
				}
				err = vm.Memory.WriteToAddress(blake2sPtrEnd, &mv)
				if err != nil {
					return err
				}
				temp, err := blake2sPtrEnd.AddOffset(1)
				if err != nil {
					return err
				}
				*blake2sPtrEnd = temp

			}
			return nil
		},
	}
}

func createBlake2sFinalizeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	blake2sPtrEnd, err := resolver.GetResOperander("blake2s_ptr_end")
	if err != nil {
		return nil, err
	}

	return newBlake2sFinalizeHint(blake2sPtrEnd), nil
}

// Blake2sFinalizeV3 hint finalizes the Blake2s hash computation, ie it verifies
// that the results of blake2s() are valid
//
// `newBlake2sFinalizeV3Hint` takes 1 operander as argument
//   - `blake2sPtrEnd` is a pointer to the address where to write the result
//
// There are 3 versions of Blake2sFinalize hint, this is the V3 implementation, with a slightly
// modification in the way padding is done compared to V1 and V2
func newBlake2sFinalizeV3Hint(blake2sPtrEnd hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Blake2sFinalizeV3",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> # Add dummy pairs of input and output.
			//> from starkware.cairo.common.cairo_blake2s.blake2s_utils import IV, blake2s_compress
			//> _n_packed_instances = int(ids.N_PACKED_INSTANCES)
			//> assert 0 <= _n_packed_instances < 20
			//> _blake2s_input_chunk_size_felts = int(ids.BLAKE2S_INPUT_CHUNK_SIZE_FELTS)
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
			//> padding = (message + modified_iv + [0, 0xffffffff] + output) * (_n_packed_instances - 1)
			//> segments.write_arg(ids.blake2s_ptr_end, padding)

			blake2sPtrEnd, err := hinter.ResolveAsAddress(vm, blake2sPtrEnd)
			if err != nil {
				return err
			}

			//> assert 0 <= _n_packed_instances < 20
			// as N_PACKED_INSTANCES is a constant of 7, this can be skipped

			//> assert 0 <= _blake2s_input_chunk_size_felts < 100
			// as BLAKE2S_INPUT_CHUNK_SIZE_FELTS is a constant of 16, this can be skipped

			message := make([]uint32, utils.INPUT_BLOCK_FELTS)
			modifiedIv := utils.IV()
			modifiedIv[0] = modifiedIv[0] ^ 0x01010020
			output := utils.Blake2sCompress(message, modifiedIv, 0, 0, 0xffffffff, 0)
			padding := message[:]
			padding = append(padding, modifiedIv[:]...)
			padding = append(padding, 0, 0xffffffff)
			padding = append(padding, output[:]...)
			fullPadding := []uint32{}
			for i := uint64(0); i < utils.N_PACKED_INSTANCES-1; i++ {
				fullPadding = append(fullPadding, padding...)
			}

			for _, val := range fullPadding {
				mv := mem.MemoryValueFromInt(val)
				if err != nil {
					return err
				}
				err = vm.Memory.WriteToAddress(blake2sPtrEnd, &mv)
				if err != nil {
					return err
				}
				temp, err := blake2sPtrEnd.AddOffset(1)
				if err != nil {
					return err
				}
				*blake2sPtrEnd = temp

			}
			return nil
		},
	}
}

func createBlake2sFinalizeV3Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	blake2sPtrEnd, err := resolver.GetResOperander("blake2s_ptr_end")
	if err != nil {
		return nil, err
	}

	return newBlake2sFinalizeV3Hint(blake2sPtrEnd), nil
}

// Blake2sCompute hint computes the blake2s compress function and fills the value in the right position.
//
// `newBlake2sComputeHint` takes 1 operander as an argument
//   - `output` should point to the middle of an instance, right after initial_state, message, t, f,
//     which should all have a value at this point, and right before the output portion which will be
//     written by this function.
func newBlake2sComputeHint(output hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Blake2sCompute",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_blake2s.blake2s_utils import compute_blake2s_func
			//> compute_blake2s_func(segments=segments, output_ptr=ids.output)

			// Expanding the compute_blake2s_func function here:
			//> def compute_blake2s_func(segments: MemorySegmentManager, output_ptr: RelocatableValue):
			//>    h = segments.memory.get_range(output_ptr - 26, 8)
			//>    message = segments.memory.get_range(output_ptr - 18, 16)
			//>    t = segments.memory[output_ptr - 2]
			//>    f = segments.memory[output_ptr - 1]
			//>    new_state = blake2s_compress(
			//>        message=message,
			//>        h=h,
			//>        t0=t,
			//>        t1=0,
			//>        f0=f,
			//>        f1=0,
			//>    )
			//>    segments.write_arg(output_ptr, new_state)

			output, err := hinter.ResolveAsAddress(vm, output)
			if err != nil {
				return err
			}

			//> h = segments.memory.get_range(output_ptr - 26, 8)
			hSegmentInput := new(mem.MemoryAddress)
			err = hSegmentInput.Sub(output, new(fp.Element).SetUint64(26))
			if err != nil {
				return err
			}
			h, err := vm.Memory.GetConsecutiveMemoryValues(*hSegmentInput, 8)
			if err != nil {
				return err
			}
			var hUint32 [8]uint32
			for i := 0; i < 8; i++ {
				value, err := hintrunnerUtils.ToSafeUint32(&h[i])
				if err != nil {
					return err
				}
				hUint32[i] = value
			}

			//> message = segments.memory.get_range(output_ptr - 18, 16)
			messageSegmentInput := new(mem.MemoryAddress)
			err = messageSegmentInput.Sub(output, new(fp.Element).SetUint64(18))
			if err != nil {
				return err
			}
			message, err := vm.Memory.GetConsecutiveMemoryValues(*messageSegmentInput, 16)
			if err != nil {
				return err
			}
			var messageUint32 []uint32
			for i := 0; i < 16; i++ {
				value, err := hintrunnerUtils.ToSafeUint32(&message[i])
				if err != nil {
					return err
				}
				messageUint32 = append(messageUint32, value)
			}

			//> t = segments.memory[output_ptr - 2]
			tSegmentInput := new(mem.MemoryAddress)
			err = tSegmentInput.Sub(output, new(fp.Element).SetUint64(2))
			if err != nil {
				return err
			}
			t, err := vm.Memory.ReadFromAddress(tSegmentInput)
			if err != nil {
				return err
			}
			tUint32, err := hintrunnerUtils.ToSafeUint32(&t)
			if err != nil {
				return err
			}

			//> f = segments.memory[output_ptr - 1]
			fSegmentInput := new(mem.MemoryAddress)
			err = fSegmentInput.Sub(output, new(fp.Element).SetUint64(1))
			if err != nil {
				return err
			}
			f, err := vm.Memory.ReadFromAddress(fSegmentInput)
			if err != nil {
				return err
			}
			fUint32, err := hintrunnerUtils.ToSafeUint32(&f)
			if err != nil {
				return err
			}

			//> new_state = blake2s_compress(
			//>     message=message,
			//>     h=h,
			//>     t0=t,
			//>     t1=0,
			//>     f0=f,
			//>     f1=0,
			//> )
			newState := utils.Blake2sCompress(messageUint32, hUint32, tUint32, 0, fUint32, 0)

			//> segments.write_arg(output_ptr, new_state)
			for i := 0; i < len(newState); i++ {
				state := newState[i]
				stateMv := mem.MemoryValueFromUint(state)
				err := vm.Memory.WriteToAddress(output, &stateMv)
				if err != nil {
					return err
				}
				*output, err = output.AddOffset(1)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createBlake2sComputeHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	output, err := resolver.GetResOperander("output")
	if err != nil {
		return nil, err
	}

	return newBlake2sComputeHint(output), nil
}

// Blake2sCompute hint computes the blake2s compress function and fills the value in the right position.
//
// `newBlake2sComputeHint` takes 1 operander as an argument
//   - `output` should point to the middle of an instance, right after initial_state, message, t, f,
//     which should all have a value at this point, and right before the output portion which will be
//     written by this function.
func newBlake2sCompressHint(output, blake2s_start hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Blake2sCompute",
		Op: func(vm *VM.VirtualMachine, _ *hinter.HintRunnerContext) error {
			// > from starkware.cairo.common.cairo_blake2s.blake2s_utils import IV, blake2s_compress
			// >
			// > _blake2s_input_chunk_size_felts = int(ids.BLAKE2S_INPUT_CHUNK_SIZE_FELTS)
			// > assert 0 <= _blake2s_input_chunk_size_felts < 100
			// >
			// > new_state = blake2s_compress(
			// >       message=memory.get_range(ids.blake2s_start, _blake2s_input_chunk_size_felts),
			// >       h=[IV[0] ^ 0x01010020] + IV[1:],
			// >       t0=ids.n_bytes,
			// >       t1=0,
			// >       f0=0xffffffff,
			// >	      f1=0,
			// >    )
			//>
			//> segments.write_arg(ids.output, new_state)

			output, err := hinter.ResolveAsAddress(vm, output)
			if err != nil {
				return err
			}

			blake2s_start, err := hinter.ResolveAsAddress(vm, blake2s_start)
			if err != nil {
				return err
			}
			//> _blake2s_input_chunk_size_felts = int(ids.BLAKE2S_INPUT_CHUNK_SIZE_FELTS)
			//> assert 0 <= _blake2s_input_chunk_size_felts < 100
			// as BLAKE2S_INPUT_CHUNK_SIZE_FELTS is a constant of 16, this can be skipped

			//> new_state = blake2s_compress(
			//>       message=memory.get_range(ids.blake2s_start, _blake2s_input_chunk_size_felts),
			//>       h=[IV[0] ^ 0x01010020] + IV[1:],
			//>       t0=ids.n_bytes,
			//>       t1=0,
			//>       f0=0xffffffff,
			//>	      f1=0,
			//>    )
			message, err := vm.Memory.GetConsecutiveMemoryValues(*blake2s_start, 16)
			if err != nil {
				return err
			}
			var messageUint32 []uint32
			for i := 0; i < 16; i++ {
				value, err := hintrunnerUtils.ToSafeUint32(&message[i])
				if err != nil {
					return err
				}
				messageUint32 = append(messageUint32, value)
			}
			h := utils.IV()
			new_state := utils.Blake2sCompress(messageUint32, h, 0, 0, 0xffffffff, 0)

			//> segments.write_arg(ids.output, new_state)
			for _, val := range new_state {
				mv := mem.MemoryValueFromInt(val)
				if err != nil {
					return err
				}
				err = vm.Memory.WriteToAddress(output, &mv)
				if err != nil {
					return err
				}
				temp, err := output.AddOffset(1)
				if err != nil {
					return err
				}
				*output = temp
			}

			return nil
		},
	}
}

func createBlake2sCompressHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	output, err := resolver.GetResOperander("output")
	if err != nil {
		return nil, err
	}

	blake2s_start, err := resolver.GetResOperander("blake2s_start")
	if err != nil {
		return nil, err
	}

	return newBlake2sCompressHint(output, blake2s_start), nil
}
