package builtins

type DilutedPoolInstance struct {
	UnitsPerStep uint32
	Spacing      uint32
	NBits        uint32
}

// DilutedPoolInstanceOption represents an option to a `DilutedPoolInstance`
type DilutedPoolInstanceOption struct {
	value *DilutedPoolInstance
}

// `SomeDilutedPoolInstance` creates a DilutedPoolInstanceOption with a value
func SomeDilutedPoolInstance(value DilutedPoolInstance) DilutedPoolInstanceOption {
	return DilutedPoolInstanceOption{value: &value}
}

// `NoneDilutedPoolInstance` creates a DilutedPoolInstanceOption without a value
func NoneDilutedPoolInstance() DilutedPoolInstanceOption {
	return DilutedPoolInstanceOption{value: nil}
}

// `IsNone` checks if the DilutedPoolInstanceOption has no value
func (o DilutedPoolInstanceOption) IsNone() bool {
	return o.value == nil
}

// `IsSome` checks if the DilutedPoolInstanceOption has a value
func (o DilutedPoolInstanceOption) IsSome() bool {
	return o.value != nil
}

// `Unwrap` returns the value if it exists, panics otherwise
func (o DilutedPoolInstanceOption) Unwrap() DilutedPoolInstance {
	if o.IsNone() {
		panic("Tried to unwrap None DilutedPoolInstanceOption")
	}
	return *o.value
}
