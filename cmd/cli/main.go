package main

import (
	"fmt"
	rn "github.com/NethermindEth/cairo-vm-go/pkg/runner"
	"github.com/urfave/cli/v2"
	"os"
)

func main() {
	var proofmode bool
	var entrypoint string
	var traceLocation string

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
						Value:       "./trace",
						Usage:       "path to file to store the trace",
						Required:    false,
						Destination: &traceLocation,
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
						return err
					}
					runner := new(rn.Runner)
					runner.LoadCairoZeroProgram(content)

					fmt.Printf("Running...")
					runner.Run(entrypoint)
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
