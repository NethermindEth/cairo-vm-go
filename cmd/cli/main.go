package main

import (
	"fmt"
	runnerzero "github.com/NethermindEth/cairo-vm-go/pkg/runners/zero"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	var proofmode bool
	var entrypoint string
	var traceLocation string
	var arguments cli.StringSlice

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
						Aliases:     []string{"p"},
						Value:       false,
						Usage:       "outputs the proof of execution",
						Required:    false,
						Destination: &proofmode,
					},
					&cli.StringFlag{
						Name:        "entrypoint",
						Aliases:     []string{"e"},
						Value:       "main",
						Usage:       "name of the function to run",
						Required:    true,
						Destination: &entrypoint,
					},
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						Usage:       "path to file to store the trace",
						Required:    false,
						Destination: &traceLocation,
					},
					&cli.StringSliceFlag{
						Name:        "args",
						Usage:       "path to file to store the trace",
						Required:    false,
						Destination: &arguments,
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

					fmt.Printf("Running....")
					runner, err := runnerzero.NewRunner(program, proofmode)
					if err != nil {
						return fmt.Errorf("cannot create runner: %w", err)
					}

					end, err := runner.InitializeMainEntrypoint()
					if err != nil {
						return fmt.Errorf("cannot create runner: %w", err)
					}

					err = runner.RunUntilPc(end)
					if err != nil {
						return fmt.Errorf("cannot create runner: %w", err)
					}

					if proofmode {
						err = runner.BuildProof()
						if err != nil {
							return err
						}
					}

					fmt.Printf("Success!")
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
