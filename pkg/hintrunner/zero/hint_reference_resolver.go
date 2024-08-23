package zero

import (
	"fmt"
	"strings"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

type hintReferenceResolver struct {
	refs map[string]hinter.Reference
}

func NewReferenceResolver() hintReferenceResolver {
	refs := make(map[string]hinter.Reference)
	return hintReferenceResolver{refs}
}

func (m *hintReferenceResolver) AddReference(name string, v hinter.Reference) error {
	shortName := shortSymbolName(name)
	_, ok := m.refs[shortName]
	if ok {
		return fmt.Errorf("cannot overwrite reference %s (%s)", shortName, name)
	}
	m.refs[shortName] = v
	return nil
}

func (m *hintReferenceResolver) GetReference(name string) (hinter.Reference, error) {
	if v, ok := m.refs[name]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("missing reference %s", name)
}

// GetResOperander returns the result of GetReference type-asserted to ResOperander.
// If reference is not found or is not of ResOperander type, a non-nil error is returned.
func (m *hintReferenceResolver) GetResOperander(name string) (hinter.Reference, error) {
	ref, err := m.GetReference(name)
	if err != nil {
		return nil, err
	}
	switch op := ref.(type) {
	case hinter.ApCellRef, hinter.FpCellRef:
		return nil, fmt.Errorf("expected %s to be ResOperander (got %T)", name, ref)
	default:
		return op, nil
	}
}

func (m *hintReferenceResolver) GetCellRefer(name string) (hinter.Reference, error) {
	ref, err := m.GetReference(name)
	if err != nil {
		return nil, err
	}
	switch op := ref.(type) {
	case hinter.Deref, hinter.DoubleDeref, hinter.BinaryOp, hinter.Immediate:
		return nil, fmt.Errorf("expected %s to be CellRefer (got %T)", name, ref)
	default:
		return op, nil
	}
}

// shortSymbolName turns a full symbol name like "a.b.c" into just "c".
func shortSymbolName(name string) string {
	i := strings.LastIndexByte(name, '.')
	if i != -1 {
		return name[i+1:]
	}
	return name
}
