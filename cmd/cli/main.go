package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	// "runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/core"
	"github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/hinter"
	hintrunner "github.com/NethermindEth/cairo-vm-go/pkg/hintrunner/zero"
	"github.com/NethermindEth/cairo-vm-go/pkg/parsers/starknet"
	zero "github.com/NethermindEth/cairo-vm-go/pkg/parsers/zero"
	"github.com/NethermindEth/cairo-vm-go/pkg/runner"
	"github.com/urfave/cli/v2"
)

func main() {
	var proofmode bool
	var buildMemory bool
	var collectTrace bool
	var maxsteps uint64
	var entrypointOffset uint64
	var traceLocation string
	var memoryLocation string
	var layoutName string
	var airPublicInputLocation string
	var airPrivateInputLocation string
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
					&cli.BoolFlag{
						Name:        "collect_trace",
						Usage:       "collects the trace and builds the relocated trace after execution",
						Required:    false,
						Destination: &collectTrace,
					},
					&cli.StringFlag{
						Name:        "tracefile",
						Usage:       "location to store the relocated trace",
						Required:    false,
						Destination: &traceLocation,
					},
					&cli.BoolFlag{
						Name:        "build_memory",
						Usage:       "builds the relocated memory after execution",
						Required:    false,
						Destination: &buildMemory,
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
					&cli.StringFlag{
						Name:        "air_public_input",
						Usage:       "location to store the air_public_input",
						Required:    false,
						Destination: &airPublicInputLocation,
					},
					&cli.StringFlag{
						Name:        "air_private_input",
						Usage:       "location to store the air_private_input",
						Required:    false,
						Destination: &airPrivateInputLocation,
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
					zeroProgram, err := zero.ZeroProgramFromFile(pathToFile)
					if err != nil {
						return fmt.Errorf("cannot load program: %w", err)
					}

					var hints map[uint64][]hinter.Hinter
					if zeroProgram.CompilerVersion[0] == '1' {
						cairoProgram, err := starknet.StarknetProgramFromFile(pathToFile)
						if err != nil {
							return fmt.Errorf("cannot load program: %w", err)
						}
						hints, err = core.GetCairoHints(cairoProgram)
						if err != nil {
							return fmt.Errorf("cannot get hints: %w", err)
						}
					} else {
						hints, err = hintrunner.GetZeroHints(zeroProgram)
						if err != nil {
							return fmt.Errorf("cannot create hints: %w", err)
						}
					}
					program, err := runner.LoadCairoZeroProgram(zeroProgram)
					if err != nil {
						return fmt.Errorf("cannot load program: %w", err)
					}
					fmt.Println("Running....")
					runner, err := runner.NewRunner(program, hints, proofmode, collectTrace, maxsteps, layoutName)
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
						if err := runner.EndRun(); err != nil {
							return fmt.Errorf("cannot end run: %w", err)
						}
						if err := runner.FinalizeSegments(); err != nil {
							return fmt.Errorf("cannot finalize segments: %w", err)
						}
					}

					if proofmode || collectTrace {
						trace, err := runner.BuildTrace()
						if err != nil {
							return fmt.Errorf("cannot build trace: %w", err)
						}

						if traceLocation != "" {
							if err := os.WriteFile(traceLocation, trace, 0644); err != nil {
								return fmt.Errorf("cannot write relocated trace: %w", err)
							}
						}
					}

					if proofmode || buildMemory {
						memory, err := runner.BuildMemory()
						if err != nil {
							return fmt.Errorf("cannot build memory: %w", err)
						}

						if memoryLocation != "" {
							if err := os.WriteFile(memoryLocation, memory, 0644); err != nil {
								return fmt.Errorf("cannot write relocated memory: %w", err)
							}
						}
					}

					if proofmode {
						if airPublicInputLocation != "" {
							airPublicInput, err := runner.GetAirPublicInput()
							if err != nil {
								return err
							}
							airPublicInputJson, err := json.MarshalIndent(airPublicInput, "", "  ")
							if err != nil {
								return err
							}
							err = os.WriteFile(airPublicInputLocation, airPublicInputJson, 0644)
							if err != nil {
								return fmt.Errorf("cannot write air_public_input: %w", err)
							}
						}

						if airPrivateInputLocation != "" {
							tracePath, err := filepath.Abs(traceLocation)
							if err != nil {
								return err
							}
							memoryPath, err := filepath.Abs(memoryLocation)
							if err != nil {
								return err
							}
							airPrivateInput, err := runner.GetAirPrivateInput(tracePath, memoryPath)
							if err != nil {
								return err
							}
							airPrivateInputJson, err := json.MarshalIndent(airPrivateInput, "", "  ")
							if err != nil {
								return err
							}
							err = os.WriteFile(airPrivateInputLocation, airPrivateInputJson, 0644)
							if err != nil {
								return fmt.Errorf("cannot write air_private_input: %w", err)
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

	f, err := os.Create("cpu.prof")
	if err != nil {
		fmt.Println("could not create CPU profile:", err)
		return
	}

	memFile, err := os.Create("mem.prof")
	if err != nil {
		fmt.Println("could not create memory profile:", err)
		return
	}

	// Trace profiling setup
	traceFile, err := os.Create("trace.out")
	if err != nil {
		fmt.Println("could not create trace file:", err)
		return
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		fmt.Println("could not start CPU profile:", err)
		f.Close()
		return
	}

	// Start tracing
	if err := trace.Start(traceFile); err != nil {
		fmt.Println("could not start trace:", err)
		traceFile.Close()
		return
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	pprof.StopCPUProfile()
	f.Close()

	trace.Stop()
	traceFile.Close()

	if err := pprof.WriteHeapProfile(memFile); err != nil {
		fmt.Println("could not write memory profile:", err)
		return
	}
	memFile.Close()
}
