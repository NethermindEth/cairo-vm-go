package zero

const (
	AllocSegmentCode string = "memory[ap] = segments.add()"

	// This is a very simple Cairo0 hint that allows us to test
	// the identifier resolution code.
	// Depending on the context, ids.a may be a complex reference.
	TestAssignCode string = "memory[ap] = ids.a"
)
