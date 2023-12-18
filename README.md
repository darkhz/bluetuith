# Documentation for bluetuith

This branch provides the [daux.io](https://daux.io/index.html)-based documentation for bluetuith.

The information contained in this README is only relevant for those interested in making changes to the documentation.

## Requirements

The only hard requirement is currently:

- [composer](https://getcomposer.org/) (install instructions provided below)

You can optionally install the following to improve the development process:

- [`GNU Make`](https://www.gnu.org/software/make/) for running `make build` and other commands
- [`Docker`](https://www.docker.com/get-started/) or [`Podman`](https://podman.io/get-started) for serving docs locally in a container

## Setup

On Fedora and other dnf-based systems:

```bash
sudo dnf install composer
```

On Debian and other apt-based systems:

```bash
sudo apt install composer
```

On Arch Linux and other pacman-based systems:

```bash
sudo pacman -S composer
```

On Alpine and any other apk-based systems:

```bash
sudo apk add composer
```

Then, [install daux according to its instructions](https://daux.io/Getting_Started.html#install):

```bash
composer global require daux/daux.io
```

The default install location for the `daux` PHP application is `~/.composer/vendor/bin/daux`.

## Making documentation changes

First, make your desired changes to the markdown files in the `documentation` directory.

Once you've completed your changes, the next step is to build the site. To help with this, a [`Makefile`](https://www.gnu.org/software/make/) has been provided. It provides a basic set of commands to build and review your documentation changes. Some Linux distributions include the `make` command, but some may not (for example, Alpine may not ship with it; install via `apk add make`).

```bash
make build
```

To review your documentation changes, you can run the following command to preview, visit `http://localhost:8099` in your browser:

```bash
make serve
```

If it all looks good, commit your changes and create a pull request.
