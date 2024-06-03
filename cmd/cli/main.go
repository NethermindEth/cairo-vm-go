package main

import (
	"fmt"
	"math"
	"os"

	hintrunner "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/zero"
	zero "github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
	runnerzero "github.com/NethermindEth/cairo-vm-go/pkg/runners/zero"
	"github.com/urfave/cli/v2"
)

func main() {
	var proofmode bool
	var maxsteps uint64
	var entrypointOffset uint64
	var traceLocation string
	var memoryLocation string
	var layoutName string
	app := &cli.App{
		Name:                 "cairo-vm",
		Usage:                "A cairo virtual machine",
		EnableBashCompletion: true,
		Suggest:              true,
		DefaultCommand:       "help",
		Commands: []*cli.Command{
			{
				Name:  "run",
				Usage: "runs a cairo zero compiled file",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "proofmode",
						Usage:       "runs the cairo vm in proof mode",
						Required:    false,
						Destination: &proofmode,
					},
					&cli.Uint64Flag{
						Name:        "maxsteps",
						Usage:       "limits the execution steps to 'maxsteps'",
						DefaultText: "2**64 - 1",
						Value:       math.MaxUint64,
						Required:    false,
						Destination: &maxsteps,
					},
					&cli.Uint64Flag{
						Name:        "entrypoint",
						Usage:       "a PC offset that will be used as an entry point (by default it executes a main function)",
						Value:       0,
						Destination: &entrypointOffset,
					},
					&cli.StringFlag{
						Name:        "tracefile",
						Usage:       "location to store the relocated trace",
						Required:    false,
						Destination: &traceLocation,
					},
					&cli.StringFlag{
						Name:        "memoryfile",
						Usage:       "location to store the relocated memory",
						Required:    false,
						Destination: &memoryLocation,
					},
					&cli.StringFlag{
						Name:        "layout",
						Usage:       "specifies the set of builtins to be used",
						Required:    false,
						Destination: &layoutName,
					},
				},
				Action: func(ctx *cli.Context) error {
					// TODO: move this action's body to a separate function to decrease the
					// code nesting significantly.

					pathToFile := ctx.Args().Get(0)
					if pathToFile == "" {
						return fmt.Errorf("path to cairo file not set")
					}

					fmt.Printf("Loading program at %s\n", pathToFile)
					content, err := os.ReadFile(pathToFile)
					if err != nil {
						return fmt.Errorf("cannot load program: %w", err)
					}
					cairoZeroJson, err := zero.ZeroProgramFromJSON(content)
					if err != nil {
						return fmt.Errorf("cannot load program: %w", err)
					}
					program, err := runnerzero.LoadCairoZeroProgram(cairoZeroJson)
					if err != nil {
						return fmt.Errorf("cannot load program: %w", err)
					}

					hints, err := hintrunner.GetZeroHints(cairoZeroJson)
					if err != nil {
						return fmt.Errorf("cannot create hints: %w", err)
					}
					fmt.Println("Running....")
					runner, err := runnerzero.NewRunner(program, hints, proofmode, maxsteps, layoutName)
					if err != nil {
						return fmt.Errorf("cannot create runner: %w", err)
					}

					// Run executes main(), RunEntryPoint is used to test contract_class-style entry points.
					// In theory, calling RunEntryPoint with main's offset should behave identically,
					// but these functions are implemented differently in both this and cairo-rs VMs
					// and the difference is quite subtle.
					if entrypointOffset == 0 {
						if err := runner.Run(); err != nil {
							return fmt.Errorf("runtime error: %w", err)
						}
					} else {
						if err := runner.RunEntryPoint(entrypointOffset); err != nil {
							return fmt.Errorf("runtime error (entrypoint=%d): %w", entrypointOffset, err)
						}
					}

					if proofmode {
						runner.EndRun()
						runner.FinalizeSegments()
						trace, memory, err := runner.BuildProof()
						if err != nil {
							return fmt.Errorf("cannot build proof: %w", err)
						}
						if traceLocation != "" {
							if err := os.WriteFile(traceLocation, trace, 0644); err != nil {
								return fmt.Errorf("cannot write relocated trace: %w", err)
							}
						}
						if memoryLocation != "" {
							if err := os.WriteFile(memoryLocation, memory, 0644); err != nil {
								return fmt.Errorf("cannot write relocated memory: %w", err)
							}
						}
					}

					fmt.Println("Success!")
					output := runner.Output()
					if len(output) > 0 {
						fmt.Println("Program output:")
						for _, val := range output {
							// cairo-run v0.11-0.13 pad the output lines with two spaces.
							fmt.Printf("  %s\n", val)
						}
					}
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
