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
	fmt.Printf("hola1")
	m.refs[shortName] = v
	fmt.Printf("hola2")
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
