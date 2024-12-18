# zsv-playground

[![ci](https://github.com/iamazeem/zsv-playground/actions/workflows/ci.yml/badge.svg)](https://github.com/iamazeem/zsv-playground/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/license-MIT-darkgreen.svg?style=flat-square)](https://github.com/iamAzeem/zsv-playground/blob/main/LICENSE)
[![GitHub release (latest by date)](https://img.shields.io/github/v/release/iamAzeem/zsv-playground?style=flat-square)](https://github.com/iamazeem/zsv-playground/releases)

[`zsv-playground`](https://github.com/iamazeem/zsv-playground) is a playground
for [`zsv`](https://github.com/liquidaty/zsv) CLI executable for AMD64 Linux,
MacOS, and FreeBSD.

> [!IMPORTANT]
>
> This is just a small [proof of
> concept](https://en.wikipedia.org/wiki/Proof_of_concept) (PoC) and is only
> meant to explore what is possible.
>
> Visit the official `zsv` playground instead:
> <https://liquidaty.github.io/zsv>.

## How it works

- On startup, `zsv-playground` downloads and extracts the latest three `zsv`
  [releases](https://github.com/liquidaty/zsv/releases) to `zsv` subdirectory on
  the current path.
- These downloaded `zsv` executables are used to serve the generated webpage in
  the browser.
- The user inputs the `CSV`, selects a command, chooses the required flags, and
  hits `Run`.
- The output is then shown under `Result`.

### Demo

![demo](./demo/demo.gif)

## Limitations

- Only global flags and main commands are supported now.
  - Other commands are not supported.
- Some commands require multiple files or non-flag CLI arguments as input.
  - This is currently not supported but may be added later.

## Download

Download the latest binaries from the
[releases](https://github.com/iamazeem/zsv-playground/releases) page.

## Run

```shell
./zsv-playground
```

Go to http://localhost:8080/ in your browser.

Run `zsv-playground --help` to check the available CLI options.

## Development

### Tech stack

- [Go 1.21.4](https://go.dev/doc/install)
- [Bootstrap 5.3.3](https://getbootstrap.com/)
- [HTMX 1.9.10](https://htmx.org/)

### Build

```shell
git clone https://github.com/iamazeem/zsv-playground.git
cd zsv-playground
go build
```

### Docker

Build:

```shell
docker build -t zsv-playground .
```

Run:

```shell
docker run -p 8080:8080 zsv-playground
```

## Contribute

Feedback is always welcome!

[Open an issue](https://github.com/iamazeem/zsv-playground/issues/new/choose) to
report bugs or propose new features and enhancements.

- [Fork](https://github.com/iamazeem/zsv-playground/fork) the project.
- Check out the latest `main` branch.
- Create a `feature` or `bugfix` branch.
- Commit and push your changes.
- Submit the PR.

## License

[MIT](./LICENSE)
