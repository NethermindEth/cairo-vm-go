package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	disasm := &disasmProgram{}
	instFields := &instFieldsProgram{}

	app := &cli.App{
		Name:                 "casm-inspect",
		Usage:                "casm-inspect <subcmd> [args...]",
		Description:          "A cairo zero file inspector",
		EnableBashCompletion: true,
		Suggest:              true,
		DefaultCommand:       "help",
		Commands: []*cli.Command{
			{
				Name:        "inst-fields",
				Usage:       "inst-fields 0xa0680017fff8000",
				Description: "print CASM instruction fields",
				Action:      instFields.Action,
			},
			{
				Name:        "disasm",
				Usage:       "disasm compiled_cairo0.json",
				Description: "disassemble the casm from the compiled cairo program",
				Action:      disasm.Action,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "bytecode",
						Usage:       "a JSON key containing CASM bytecode (period-separated for multi-keys)",
						Required:    false,
						Value:       "data",
						Destination: &disasm.bytecodeKey,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
