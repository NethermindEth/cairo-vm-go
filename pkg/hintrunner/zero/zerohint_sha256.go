package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	hintrunnerUtils "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/utils"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	mem "github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"

	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
)

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
				memoryOffset := uint64(i)

				err = vm.Memory.Write(output.SegmentIndex, output.Offset+memoryOffset, &newStateValue)
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
