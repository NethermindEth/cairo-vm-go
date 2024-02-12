package zero

import (
	"fmt"
	"strings"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

type hintReferenceResolver struct {
	refs map[string]hinter.Reference
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

// shortSymbolName turns a full symbol name like "a.b.c" into just "c".
func shortSymbolName(name string) string {
	i := strings.LastIndexByte(name, '.')
	if i != -1 {
		return name[i+1:]
	}
	return name
}
