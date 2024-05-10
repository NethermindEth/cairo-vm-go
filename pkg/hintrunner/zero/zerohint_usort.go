package zero

import (
	"fmt"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	"github.com/NethermindEth/cairo-vm-go/pkg/utils"
	VM "github.com/NethermindEth/cairo-vm-go/pkg/vm"
	"github.com/NethermindEth/cairo-vm-go/pkg/vm/memory"
	"github.com/consensys/gnark-crypto/ecc/stark-curve/fp"
)

// UsortEnterScope hint enters a new scope with `__usort_max_size` value
//
// `newUsortEnterScopeHint` doesn't take any operander as argument
//
// `newUsortEnterScopeHint` gets `__usort_max_size` value from the current
// scope and enters a new scope with this same value
func newUsortEnterScopeHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortEnterScope",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> vm_enter_scope(dict(__usort_max_size = globals().get('__usort_max_size')))

			usortMaxSize, err := ctx.ScopeManager.GetVariableValue("__usort_max_size")
			if err != nil {
				return err
			}

			ctx.ScopeManager.EnterScope(map[string]any{
				"__usort_max_size": usortMaxSize,
			})

			return nil
		},
	}
}

func createUsortEnterScopeHinter() (hinter.Hinter, error) {
	return newUsortEnterScopeHint(), nil
}

// UsortVerifyMultiplicityAssert hint checks that the `positions` variable in scope
// doesn't contain any value
//
// `newUsortVerifyMultiplicityAssertHint` doesn't take any operander as argument
//
// This hint is used when sorting an array of field elements while removing duplicates
// in `usort` Cairo function
func newUsortVerifyMultiplicityAssertHint() hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortVerifyMultiplicityAssert",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> assert len(positions) == 0
			positionsInterface, err := ctx.ScopeManager.GetVariableValue("positions")

			if err != nil {
				return err
			}

			positions, ok := positionsInterface.([]uint64)
			if !ok {
				return fmt.Errorf("casting positions into an array failed")
			}

			if len(positions) != 0 {
				return fmt.Errorf("assertion `len(positions) == 0` failed")
			}

			return nil
		},
	}
}

func createUsortVerifyMultiplicityAssertHinter() (hinter.Hinter, error) {
	return newUsortEnterScopeHint(), nil
}

// UsortVerify hint prepares for verifying the presence of duplicates of
// a specific value in the sorted output (array of fields)
//
// `newUsortVerifyHint` takes one operander as argument
//   - `value` is the value at the given position in the output
//
// `last_pos` is set to zero
// `positions` is set to the reversed order list associated with `ids.value`
// key in `positions_dict`
// `newUsortVerifyHint` assigns `last_pos` and `positions` in the current scope
func newUsortVerifyHint(value hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortVerify",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> last_pos = 0
			//> positions = positions_dict[ids.value][::-1]

			positionsDictInterface, err := ctx.ScopeManager.GetVariableValue("positions_dict")

			if err != nil {
				return err
			}

			positionsDict, ok := positionsDictInterface.(map[fp.Element][]uint64)

			if !ok {
				return fmt.Errorf("casting positions_dict into an dictionary failed")
			}

			value, err := hinter.ResolveAsFelt(vm, value)

			if err != nil {
				return err
			}

			positions := positionsDict[*value]
			utils.Reverse(positions)

			return ctx.ScopeManager.AssignVariables(map[string]any{
				"last_pos":  0,
				"positions": positions,
			})
		},
	}
}

func createUsortVerifyHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	value, err := resolver.GetResOperander("value")

	if err != nil {
		return nil, err
	}

	return newUsortVerifyHint(value), nil
}

// UsortVerifyMultiplicityBody hint extracts a specific value
// of the sorted output with `pop`, updating indices for the verification
// of the next value
//
// `newUsortVerifyMultiplicityBodyHint` takes one operander as argument
//   - `nextItemIndex` is the index of the next item
//
// `next_item_index` is set to `current_pos - last_pos` for the next iteration
// `newUsortVerifyMultiplicityBodyHint` assigns `last_pos` in the current scope
func newUsortVerifyMultiplicityBodyHint(nextItemIndex hinter.ResOperander) hinter.Hinter {
	return &GenericZeroHinter{
		Name: "UsortVerifyMultiplicityBody",
		Op: func(vm *VM.VirtualMachine, ctx *hinter.HintRunnerContext) error {
			//> current_pos = positions.pop()
			//> ids.next_item_index = current_pos - last_pos
			//> last_pos = current_pos + 1

			positionsInterface, err := ctx.ScopeManager.GetVariableValue("positions")
			if err != nil {
				return err
			}

			positions, ok := positionsInterface.([]fp.Element)
			if !ok {
				return fmt.Errorf("cannot cast positionsInterface to []fp.Element")
			}

			currentPos, err := utils.Pop(&positions)
			if err != nil {
				return err
			}

			err = ctx.ScopeManager.AssignVariable("positions", positions)
			if err != nil {
				return err
			}

			lastPos, err := ctx.ScopeManager.GetVariableValue("last_pos")
			if err != nil {
				return err
			}

			lastPosFelt, ok := lastPos.(fp.Element)
			if !ok {
				return fmt.Errorf("cannot cast last_pos to felt")
			}

			// Calculate `next_item_index` memory value
			var newNextItemIndexValue fp.Element
			newNextItemIndexValue.Sub(&currentPos, &lastPosFelt)
			newNextItemIndexMemoryValue := memory.MemoryValueFromFieldElement(&newNextItemIndexValue)

			// Save `next_item_index` value in address
			addrNextItemIndex, err := nextItemIndex.GetAddress(vm)
			if err != nil {
				return err
			}

			err = vm.Memory.WriteToAddress(&addrNextItemIndex, &newNextItemIndexMemoryValue)
			if err != nil {
				return err
			}

			return ctx.ScopeManager.AssignVariable("last_pos", currentPos.Add(&currentPos, &utils.FeltOne))
		},
	}
}

func createUsortVerifyMultiplicityBodyHinter(resolver hintReferenceResolver) (hinter.Hinter, error) {
	nextItemIndex, err := resolver.GetResOperander("next_item_index")
	if err != nil {
		return nil, err
	}

	return newUsortVerifyMultiplicityBodyHint(nextItemIndex), nil
}
