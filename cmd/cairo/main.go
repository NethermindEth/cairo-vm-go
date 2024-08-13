package main

import (
	"fmt"
	"os"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	"github.com/urfave/cli/v2"
)

func main() {
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
				Flags: []cli.Flag{},
				Action: func(ctx *cli.Context) error {
					pathToFile := ctx.Args().Get(0)
					fmt.Println(pathToFile)
					if pathToFile == "" {
						return fmt.Errorf("path to cairo file not set")
					}

					fmt.Printf("Loading program at %s\n", pathToFile)

					cairoZeroJson, err := starknet.StarknetProgramFromFile(pathToFile)
					if err != nil {
						return fmt.Errorf("cannot load program: %w", err)
					}
					cairoHints, err := core.GetCairoHints(cairoZeroJson)
					if err != nil {
						return fmt.Errorf("cannot get hints: %w", err)
					}
					fmt.Println(cairoHints)
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
