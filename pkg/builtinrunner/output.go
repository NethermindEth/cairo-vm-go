package builtinrunner

import "fmt"

type Output struct {
	segmentIndex uint64
}

func NewOutput(segmentIndex uint64) Output {
	return Output{segmentIndex: segmentIndex}
}

func (o Output) Segment() uint64 {
	return o.segmentIndex
}

func (o Output) Run() {
	fmt.Println("Running output builtin")
}

func (o Output) Name() string {
	return "output"
}
