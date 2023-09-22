package main

import (
	"fmt"
	"os"

	runnerzero "github.com/NethermindEth/cairo-vm-go/pkg/runners/zero"
	"github.com/urfave/cli/v2"
)

func main() {
	var proofmode bool
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
				Usage: "runs a cairo file",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:        "proofmode",
						Usage:       "runs the cairo vm in proof mode",
						Required:    false,
						Destination: &proofmode,
					},
					&cli.StringFlag{
						Name:        "tracelocation",
						Usage:       "location to store the relocated trace",
						Required:    false,
						Destination: &traceLocation,
					},
					&cli.StringFlag{
						Name:        "memorylocation",
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
					runner, err := runnerzero.NewRunner(program, proofmode)
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
						if err := os.WriteFile(traceLocation, trace, 0644); err != nil {
							return fmt.Errorf("cannot write relocated trace: %w", err)
						}
						if err := os.WriteFile(memoryLocation, memory, 0644); err != nil {
							return fmt.Errorf("cannot write relocated memory: %w", err)
						}
					}

					fmt.Println("Success!")
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
