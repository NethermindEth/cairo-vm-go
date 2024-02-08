package zero

import (
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
)

type hintReferenceResolver struct {
	refs []hintReference

	numCellRefer    int
	numResOperander int
}

type hintReference struct {
	name string

	cell      hinter.CellRefer
	operander hinter.ResOperander
}

func (m *hintReferenceResolver) NumCellRefers() int { return m.numCellRefer }

func (m *hintReferenceResolver) NumResOperanders() int { return m.numResOperander }

func (m *hintReferenceResolver) AddCellRefer(name string, v hinter.CellRefer) {
	m.refs = append(m.refs, hintReference{
		name: name,
		cell: v,
	})
	m.numCellRefer++
}

func (m *hintReferenceResolver) AddResOperander(name string, v hinter.ResOperander) {
	m.refs = append(m.refs, hintReference{
		name:      name,
		operander: v,
	})
	m.numResOperander++
}

func (m *hintReferenceResolver) GetResOperander(name string) hinter.ResOperander {
	ref := m.find(name)
	if ref != nil {
		return ref.operander
	}
	return nil
}

func (m *hintReferenceResolver) find(name string) *hintReference {
	for i, ref := range m.refs {
		if ref.name == name {
			return &m.refs[i]
		}
	}
	return nil
}
