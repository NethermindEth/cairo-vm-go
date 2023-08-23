# Cairo VM in Go

> ⚠️  This project is undergoing heavy development and is still on its early stages. There will be constant breaking changes.

This project aims to implement a Cairo Virtual Machine using Go. This is one of [many other](#related-projects) implementations that are being developed and its main goals are:

* making the Starknet ecosystem secure by reducing the risk of a single critical vulnerability,
* decentralizing development and maintenance of the different VMs,
* cross-checking and validation with other implementations and
* to foster innovation through competition.

## Intro

The Cairo Virtual Machine is a crucial component of the Starknet ecosystem. It serves as the runtime environment for all smart contracts on the platform. 
When users write contracts in high-level Cairo, it gets compiled to Sierra, and then to CASM bytecode. The VM receives this bytecode, executes it and generates a proof of execution. This proof is then sent from a sequencer to the verifier to include the transaction in a new block.

## Install

This Virtual Machine is still in development and there is no public release available yet.
Currently, it is only possible to use it by building it from source by following these instructions:

1. Clone the repo to your machine: `git clone https://github.com/NethermindEth/cairo-vm-go`.
2. Install `golang` on your PC, instructions [here](https://go.dev/.doc/install).
3. Execute on the root folder of the repo: `make build`.
4. Make sure everything is running smoothly by executing: `make test`.

After completing these steps, you can find the compiled VM in `bin/cairo-vm`. It is worth noting that this binary, in the current state of the project, is still non-functional.

### Useful Commands

For convenience, we have created a `makefile` that includes the most used commands such as `make build`. To see all of them please run:
```bash
make help
```

## Documentation

### Cairo

* Cairo Zero Docs: [How Cairo Works](https://docs.cairo-lang.org/0.12.0/how_cairo_works/index.html)
* Whitepaper: [Cairo a Turing-complete STARK-friendly CPU architecture](https://eprint.iacr.org/2021/1063.pdf)
* A formalization of the whitepaper: [A Verified Algebraic Representation of Cairo Program Execution](https://arxiv.org/pdf/2109.14534.pdf)

### Other

The previous list includes the most helpful documentation for the current state of the project but it does not represent all available documentation. If you are interested to go beyond, there is this [list](https://github.com/lambdaclass/cairo-vm#-documentation) made by [LambdaClass](https://github.com/lambdaclass) which has a much broader scope.

## Related Projects

* [Cairo Zero VM](https://github.com/starkware-libs/cairo-lang) by [Starkware](https://github.com/starkware-libs).
* [VM in Rust](https://github.com/lambdaclass/cairo-vm), [Go](https://github.com/lambdaclass/cairo-vm.go) and [C](https://github.com/lambdaclass/cairo-vm.c) by [LambdaClass](https://github.com/lambdaclass).
* [oriac](https://github.com/xJonathanLEI/oriac/) a toy VM by [xJonathanLEI](https://github.com/xJonathanLEI)

## Contributing

If you wish to contribute please visit our [CONTRIBUTING.md](./CONTRIBUTING.md) for general guidelines.
