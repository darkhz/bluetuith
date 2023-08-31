[TOC]

The general usage syntax is:

```
bluetuith [<option>=<parameter>]
```

or 

```
bluetuith [<option> <parameter>]
```

- [adapter](#adapter): Specify an adapter to use.
- [list-adapters](#list-adapters): List available adapters.
- [receive-dir](#receive-dir): Specify a directory to store received files.
- [gsm-apn](#gsm-apn): Specify GSM APN to connect to.
- [gsm-number](#gsm-number): Specify GSM number to dial.
- [theme](#theme): Specify a theme in the HJSON format.
- [generate](#generate): Generate configuration.
- [version](#version): Print version information.

!!! warning "set-theme and set-theme-config"
	As of v0.1.7, the set-theme and set-theme-config options are deprecated. Use the `--theme` command-line option or specify a theme directive within the configuration file.

## adapter
This option can be used to select the adapter when the application has initialized. 

For example:
```
bluetuith --adapter=hci0
```

## list-adapters
This option can be used to list the available bluetooth adapters.

## receive-dir
This option can be used to set the directory to receive transferred files.

If this option is not set, the path will be set to `$HOME/bluetuith`.

## gsm-apn
This option can be used to specify the GSM APN to connect to.
While performing DUN-base networking, this option is required and can be used along with the `gsm-number` option.

## gsm-number
This option can be used to specify the GSM number to connect to.
If this option is set, the `gsm-apn` option must also be provided.

## theme
This option can be used to set the theme for the application.
For example:
```
bluetuith --theme='{ Adapter: "red" }'
```

To see the available element types and colors, use the `--help` option.

## generate
This option can be used to generate the configuration.

Note that if you are regenerating the config, the existing values will be re-applied to the generated output.

## version
This option can be used to print the current version of the application.
