# zsv-playground

[`zsv-playground`](https://github.com/iamazeem/zsv-playground) is a PoC for
[`zsv`](https://github.com/liquidaty/zsv).

Currently, it only works on Linux.

## Download

Download the latest build from the
[releases](https://github.com/iamazeem/zsv-playground/releases) page.

## Run

```shell
./zsv-playground
```

Go to http://localhost:8080/ in your browser.

## How it works

- `zsv-playground` aims to be a single static binary with everything bundled in
  it.
- By default, it uses the latest three `zsv`
  [releases](https://github.com/liquidaty/zsv/releases). It does not build `zsv`
  itself.
  - These downloaded releases are extracted to a subdirectory named `zsv` on the
    current path.
  - Remove this directory manually if it's no longer needed.
- These downloaded `zsv` versions are used to generate the HTML with the main
  commands and their respective flags.
- The user inputs the CSV, selects a command, chooses the required flags, and
  hits `Run`.
- The generated user CLI is run by the backend and the output is shown under
  `Results`.

### Demo

TODO: Add a demo GIF here!

## Build

To build it locally:

- Set up the [Go](https://go.dev/doc/install) environment.
  - Check [`go.mod`](./go.mod) file to check the required Go version.
- Clone this project and `cd` into its directory.
- Run `go build`.

> [!NOTE]
>
> `zsv-playground` itself statically serves [Bootstrap
> v5.3.3](https://getbootstrap.com/docs/5.3/getting-started/introduction/) (CSS
> only) for UI with some JavaScript for input validation and for a better UX.
> The only thing that it downloads to work with is `zsv`.

## Limitations

- This is just a small [PoC](https://en.wikipedia.org/wiki/Proof_of_concept)!
- Not every command is supposed to work.
  - Only global flags and the main commands are being supported now.
  - The CLI differs from command to command and may not be fully supported.
  - However, this PoC tries to converge most of them.
- The basic input validation is supported but there is still room for
  improvement.
- Some commands require multiple files as input.
  - This is currently not supported but may be added later.
- To avoid duplicate runs, there is no caching at the moment.

## License

[MIT](./LICENSE)
