package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	hintrunnerUtils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)


func newFinalizeSha256Hint(sha256PtrEnd hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "FinalizeSha256",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> # Add dummy pairs of input and output.
			//> from starkware.cairo.common.cairo_sha256.sha256_utils import (
			//>     IV, compute_message_schedule, sha2_compress_function)
			//>
			//> _block_size = int(ids.BLOCK_SIZE)
			//> assert 0 <= _block_size < 20
			//> _sha256_input_chunk_size_felts = int(ids.SHA256_INPUT_CHUNK_SIZE_FELTS)
			//> assert 0 <= _sha256_input_chunk_size_felts < 100
			//>
			//> message = [0] * _sha256_input_chunk_size_felts
			//> w = compute_message_schedule(message)
			//> output = sha2_compress_function(IV, w)
			//> padding = (message + IV + output) * (_block_size - 1)
			//> segments.write_arg(ids.sha256_ptr_end, padding)

			//> _block_size = int(ids.BLOCK_SIZE)
			//> assert 0 <= _block_size < 20
			blockSize := 7

			//> _sha256_input_chunk_size_felts = int(ids.SHA256_INPUT_CHUNK_SIZE_FELTS)
			//> assert 0 <= _sha256_input_chunk_size_felts < 100
			sha256InputChunkSize := 16

			//> message = [0] * _sha256_input_chunk_size_felts
			message := make([]uint32, sha256InputChunkSize)

			//> w = compute_message_schedule(message)
			w, err := utils.ComputeMessageSchedule(message)
			if err != nil {
				return err
			}

			//> output = sha2_compress_function(IV, w)
			iv := utils.IV()
			output := utils.Sha256Compress(iv, w)

			//> padding = (message + IV + output) * (_block_size - 1)
			paddingSize := (len(message) + len(iv) + len(output)) * (blockSize - 1)
			padding := make([]uint32, 0, paddingSize)
			for i := 0; i < blockSize-1; i++ {
				padding = append(padding, message...)
				padding = append(padding, iv[:]...)
				padding = append(padding, output...)
			}

			//> segments.write_arg(ids.sha256_ptr_end, padding)
			sha256PtrEnd, err := hinter.ResolveAsAddress(vm, sha256PtrEnd)
			if err != nil {
				return err
			}
			for i := 0; i < paddingSize; i++ {
				paddingValue := mem.MemoryValueFromInt(padding[i])
				paddingOffset, err := sha256PtrEnd.AddOffset(int16(i))
				if err != nil {
					return err
				}

				err = vm.Memory.WriteToAddress(&paddingOffset, &paddingValue)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createFinalizeSha256Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	sha256PtrEnd, err := resolver.GetResOperander("sha256_ptr_end")
	if err != nil {
		return nil, err
	}

	return newFinalizeSha256Hint(sha256PtrEnd), nil
}


func newPackedSha256Hint(sha256Start, output hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "PackedSha256",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_sha256.sha256_utils import (
			//> 	IV, compute_message_schedule, sha2_compress_function)
			//>
			//> _sha256_input_chunk_size_felts = int(ids.SHA256_INPUT_CHUNK_SIZE_FELTS)
			//> assert 0 <= _sha256_input_chunk_size_felts < 100
			//>
			//> w = compute_message_schedule(memory.get_range(
			//> 	ids.sha256_start, _sha256_input_chunk_size_felts))
			//> new_state = sha2_compress_function(IV, w)
			//> segments.write_arg(ids.output, new_state)

			Sha256InputChunkSize := uint64(16)

			sha256Start, err := hinter.ResolveAsAddress(vm, sha256Start)
			if err != nil {
				return err
			}

			w, err := vm.Memory.GetConsecutiveMemoryValues(*sha256Start, Sha256InputChunkSize)
			if err != nil {
				return err
			}

			wUint32 := make([]uint32, len(w))
			for i := 0; i < len(w); i++ {
				value, err := hintrunnerUtils.ToSafeUint32(&w[i])
				if err != nil {
					return err
				}
				wUint32[i] = value
			}

			messageSchedule, err := utils.ComputeMessageSchedule(wUint32)
			if err != nil {
				return err
			}
			newState := utils.Sha256Compress(utils.IV(), messageSchedule)

			output, err := hinter.ResolveAsAddress(vm, output)
			if err != nil {
				return err
			}

			for i := 0; i < len(newState); i++ {
				newStateValue := mem.MemoryValueFromInt(newState[i])
				outputOffset, err := output.AddOffset(int16(i))
				if err != nil {
					return err
				}

				err = vm.Memory.WriteToAddress(&outputOffset, &newStateValue)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createPackedSha256Hinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	sha256Start, err := resolver.GetResOperander("sha256_start")
	if err != nil {
		return nil, err
	}

	output, err := resolver.GetResOperander("output")
	if err != nil {
		return nil, err
	}

	return newPackedSha256Hint(sha256Start, output), nil
}

func newSha256ChunkHint(sha256Start, state, output hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "Sha256Chunk",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> from starkware.cairo.common.cairo_sha256.sha256_utils import (
			//> 	compute_message_schedule, sha2_compress_function)
			//>
			//> _sha256_input_chunk_size_felts = int(ids.SHA256_INPUT_CHUNK_SIZE_FELTS)
			//> assert 0 <= _sha256_input_chunk_size_felts < 100
			//> _sha256_state_size_felts = int(ids.SHA256_STATE_SIZE_FELTS)
			//> assert 0 <= _sha256_state_size_felts < 100
			//> w = compute_message_schedule(memory.get_range(
			//> 	ids.sha256_start, _sha256_input_chunk_size_felts))
			//> new_state = sha2_compress_function(memory.get_range(ids.state, _sha256_state_size_felts), w)
			//> segments.write_arg(ids.output, new_state)

			Sha256InputChunkSize := uint64(16)

			sha256Start, err := hinter.ResolveAsAddress(vm, sha256Start)
			if err != nil {
				return err
			}

			w, err := vm.Memory.GetConsecutiveMemoryValues(*sha256Start, Sha256InputChunkSize)
			if err != nil {
				return err
			}

			wUint32 := make([]uint32, len(w))
			for i := 0; i < len(w); i++ {
				value, err := hintrunnerUtils.ToSafeUint32(&w[i])
				if err != nil {
					return err
				}
				wUint32[i] = value
			}

			messageSchedule, err := utils.ComputeMessageSchedule(wUint32)
			if err != nil {
				return err
			}

			Sha256StateSize := uint64(8)

			stateAddr, err := hinter.ResolveAsAddress(vm, state)
			if err != nil {
				return err
			}

			stateValuesMv, err := vm.Memory.GetConsecutiveMemoryValues(*stateAddr, Sha256StateSize)
			if err != nil {
				return err
			}

			var stateValues [8]uint32
			for i := 0; i < len(stateValues); i++ {
				value, err := hintrunnerUtils.ToSafeUint32(&stateValuesMv[i])
				if err != nil {
					return err
				}
				stateValues[i] = value
			}

			newState := utils.Sha256Compress(stateValues, messageSchedule)

			output, err := hinter.ResolveAsAddress(vm, output)
			if err != nil {
				return err
			}

			for i := 0; i < len(newState); i++ {
				newStateValue := mem.MemoryValueFromInt(newState[i])
				outputOffset, err := output.AddOffset(int16(i))
				if err != nil {
					return err
				}

				err = vm.Memory.WriteToAddress(&outputOffset, &newStateValue)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}
}

func createSha256ChunkHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	sha256Start, err := resolver.GetResOperander("sha256_start")
	if err != nil {
		return nil, err
	}

	state, err := resolver.GetResOperander("state")
	if err != nil {
		return nil, err
	}

	output, err := resolver.GetResOperander("output")
	if err != nil {
		return nil, err
	}

	return newSha256ChunkHint(sha256Start, state, output), nil
}