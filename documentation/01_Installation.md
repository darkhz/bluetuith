# Requirements

Before installation, ensure that the required dependencies are installed:

- Bluez
- DBus
- NetworkManager (optional, required for PANU)
- ModemManager (optional, required for DUN)
- PulseAudio (optional, required to manage device audio profiles)

# Installation

After installing the dependencies, you can install bluetuith using any one method listed below.

## Package manager

If your distribution's repositories have bluetuith, you can install it directly with your package manager.
For **Arch Linux**, The package bluetuith-bin is in the AUR. Install it using an AUR helper of your choice like so:

```bash
<aur-helper> -S bluetuith-bin.
```

For more information on whether bluetuith is packaged for your distribution, check the [repology](https://repology.org/project/bluetuith/versions) page.

## Releases

You can retrieve the package's tagged release from the project's [Releases](https://github.com/darkhz/bluetuith/releases/) page.

Before downloading, note:

- The latest tag of the release
- Your operating system (for example, Linux)
- Your architecture (for example, x86_64)

The binary is packaged in a gzipped tar file (with the extension `.tar.gz`) in the format:
`bluetuith_<tag>_<Operating System>_<Architecture>.tar.gz`

To download a package for:

- with the release tag 'v0.1.7',
- a 'Linux' distribution,
- on the 'x86_64' architecture,

You would select: `bluetuith_0.1.7_Linux_x86_64.tar.gz`

You can follow these steps for other Operating Systems as well. Note that for Apple computers like Macs, the Operating System is **Darwin**.

## Go Toolchain

Ensure that the `go` binary is present in your system before following the listed steps. [Visit the official Go documentation to learn how to install it](https://go.dev/doc/install).

### Install

To install it directly from the repository without having to compile anything, use:

```bash
go install github.com/darkhz/bluetuith@latest
```

Note that the installed binary may be present in `~/go/bin` (may vary based on your Go environment configuration), so ensure that your `$PATH` points to that directory as well.

### Compile

- Clone the [source](https://github.com/darkhz/bluetuith/) into a folder using git, like so:<br/><br/>

  ```bash
  git clone https://github.com/darkhz/bluetuith.git
  ```

  The source should be cloned into a directory named "**bluetuith**".<br/><br/>

- Next, change the directory to the "**bluetuith**" folder:<br/><br/>

  ```bash
  cd bluetuith
  ```

- Finally, use the Go toolchain to build the project:<br/><br/>

  ```bash
  go get -v && go build -v
  ```

  After the build process finishes, a binary named **`bluetuith`** should be present in your current directory.

  *To produce a binary named `bluetuith-bin`, update the above `go build` command to `go build -v -o bluetuith-bin`.*
