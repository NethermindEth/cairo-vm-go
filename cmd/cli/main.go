package main

import (
	"fmt"
	"math"
	"os"

	runnerzero "github.com/NethermindEth/cairo-vm-go/pkg/runners/zero"
	"github.com/urfave/cli/v2"
)

func main() {
	var proofmode bool
	var maxsteps uint64
	var traceLocation string
	var memoryLocation string

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
				},
				Action: func(ctx *cli.Context) error {
					pathToFile := ctx.Args().Get(0)
					if pathToFile == "" {
						return fmt.Errorf("path to cairo file not set")
					}

					fmt.Printf("Loading program at %s\n", pathToFile)
					content, err := os.ReadFile(pathToFile)
					if err != nil {
						return fmt.Errorf("cannot load program: %w", err)
					}
					program, err := runnerzero.LoadCairoZeroProgram(content)
					if err != nil {
						return fmt.Errorf("cannot load program: %w", err)
					}

					fmt.Println("Running....")
					runner, err := runnerzero.NewRunner(program, proofmode, maxsteps)
					if err != nil {
						return fmt.Errorf("cannot create runner: %w", err)
					}

					if err := runner.Run(); err != nil {
						return fmt.Errorf("runtime error: %w", err)
					}

					if proofmode {
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
							fmt.Printf("\t%s\n", val)
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
