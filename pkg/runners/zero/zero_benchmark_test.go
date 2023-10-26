package zero

import (
	"math"
	"testing"
)

func BenchmarkRunnerWithFibonacci(b *testing.B) {
	// compiled fibonacci to the millionth
	compiledJson := []byte(`
        {
            "compiler_version": "0.11.0.2",
            "data": [
                "0x40780017fff7fff",
                "0x0",
                "0x1104800180018000",
                "0x4",
                "0x10780017fff7fff",
                "0x0",
                "0x480680017fff8000",
                "0x1",
                "0x480680017fff8000",
                "0x1",
                "0x480680017fff8000",
                "0xf4240",
                "0x1104800180018000",
                "0x3",
                "0x208b7fff7fff7ffe",
                "0x20780017fff7ffd",
                "0x4",
                "0x480a7ffc7fff8000",
                "0x208b7fff7fff7ffe",
                "0x482a7ffc7ffb8000",
                "0x480a7ffc7fff8000",
                "0x48127ffe7fff8000",
                "0x482680017ffd8000",
                "0x800000000000011000000000000000000000000000000000000000000000000",
                "0x1104800180018000",
                "0x800000000000010fffffffffffffffffffffffffffffffffffffffffffffff8",
                "0x208b7fff7fff7ffe"
            ],
            "identifiers": {
                "__main__.__end__": {
                    "pc": 4,
                    "type": "label"
                },
                "__main__.__start__": {
                    "pc": 0,
                    "type": "label"
                },
                "__main__.fib": {
                    "decorators": [],
                    "pc": 15,
                    "type": "function"
                },
                "__main__.fib.Args": {
                    "full_name": "__main__.fib.Args",
                    "members": {
                        "first_element": {
                            "cairo_type": "felt",
                            "offset": 0
                        },
                        "n": {
                            "cairo_type": "felt",
                            "offset": 2
                        },
                        "second_element": {
                            "cairo_type": "felt",
                            "offset": 1
                        }
                    },
                    "size": 3,
                    "type": "struct"
                },
                "__main__.fib.ImplicitArgs": {
                    "full_name": "__main__.fib.ImplicitArgs",
                    "members": {},
                    "size": 0,
                    "type": "struct"
                },
                "__main__.fib.Return": {
                    "cairo_type": "(res: felt)",
                    "type": "type_definition"
                },
                "__main__.fib.SIZEOF_LOCALS": {
                    "type": "const",
                    "value": 0
                },
                "__main__.fib.first_element": {
                    "cairo_type": "felt",
                    "full_name": "__main__.fib.first_element",
                    "references": [
                        {
                            "ap_tracking_data": {
                                "group": 4,
                                "offset": 0
                            },
                            "pc": 15,
                            "value": "[cast(fp + (-5), felt*)]"
                        }
                    ],
                    "type": "reference"
                },
                "__main__.fib.n": {
                    "cairo_type": "felt",
                    "full_name": "__main__.fib.n",
                    "references": [
                        {
                            "ap_tracking_data": {
                                "group": 4,
                                "offset": 0
                            },
                            "pc": 15,
                            "value": "[cast(fp + (-3), felt*)]"
                        }
                    ],
                    "type": "reference"
                },
                "__main__.fib.second_element": {
                    "cairo_type": "felt",
                    "full_name": "__main__.fib.second_element",
                    "references": [
                        {
                            "ap_tracking_data": {
                                "group": 4,
                                "offset": 0
                            },
                            "pc": 15,
                            "value": "[cast(fp + (-4), felt*)]"
                        }
                    ],
                    "type": "reference"
                },
                "__main__.fib.y": {
                    "cairo_type": "felt",
                    "full_name": "__main__.fib.y",
                    "references": [
                        {
                            "ap_tracking_data": {
                                "group": 4,
                                "offset": 1
                            },
                            "pc": 20,
                            "value": "[cast(ap + (-1), felt*)]"
                        }
                    ],
                    "type": "reference"
                },
                "__main__.main": {
                    "decorators": [],
                    "pc": 6,
                    "type": "function"
                },
                "__main__.main.Args": {
                    "full_name": "__main__.main.Args",
                    "members": {},
                    "size": 0,
                    "type": "struct"
                },
                "__main__.main.ImplicitArgs": {
                    "full_name": "__main__.main.ImplicitArgs",
                    "members": {},
                    "size": 0,
                    "type": "struct"
                },
                "__main__.main.Return": {
                    "cairo_type": "()",
                    "type": "type_definition"
                },
                "__main__.main.SIZEOF_LOCALS": {
                    "type": "const",
                    "value": 0
                }
            },
            "main_scope": "__main__",
            "prime": "0x800000000000011000000000000000000000000000000000000000000000001",
            "reference_manager": {
                "references": [
                    {
                        "ap_tracking_data": {
                            "group": 4,
                            "offset": 0
                        },
                        "pc": 15,
                        "value": "[cast(fp + (-5), felt*)]"
                    },
                    {
                        "ap_tracking_data": {
                            "group": 4,
                            "offset": 0
                        },
                        "pc": 15,
                        "value": "[cast(fp + (-4), felt*)]"
                    },
                    {
                        "ap_tracking_data": {
                            "group": 4,
                            "offset": 0
                        },
                        "pc": 15,
                        "value": "[cast(fp + (-3), felt*)]"
                    },
                    {
                        "ap_tracking_data": {
                            "group": 4,
                            "offset": 1
                        },
                        "pc": 20,
                        "value": "[cast(ap + (-1), felt*)]"
                    }
                ]
            }
        }
    `)

	for i := 0; i < b.N; i++ {
		program, err := LoadCairoZeroProgram(compiledJson)
		if err != nil {
			panic(err)
		}

		runner, err := NewRunner(program, true, math.MaxUint64)
		if err != nil {
			panic(err)
		}

		err = runner.Run()
		if err != nil {
			panic(err)
		}
	}
}
