package builtinrunner

import "fmt"

type RangeCheck struct {
	segmentIndex uint64
}

func NewRangeCheck(segmentIndex uint64) RangeCheck {
	return RangeCheck{segmentIndex: segmentIndex}
}

func (r RangeCheck) Segment() uint64 {
	return r.segmentIndex
}

func (r RangeCheck) Run() {
	fmt.Println("Running range check builtin")
}

func (r RangeCheck) Name() string {
	return "range check"
}
