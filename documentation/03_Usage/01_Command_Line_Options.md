[TOC]

The general usage syntax is:

```text
bluetuith [<option>=<parameter>]
```

or

```text
bluetuith [<option> <parameter>]
```

- [adapter](#adapter): Specify an adapter to use.
- [list-adapters](#list-adapters): List available adapters.
- [adapter-states](#adapter-states): Specify adapter states to enable/disable.
- [connect-bdaddr](#connect-bdaddr): Specify device address to connect.
- [confirm-on-quit](#confirm-on-quit): Ask for confirmation before quitting the application.
- [no-warning](#no-warning): Do not display warnings when the application has initialized.
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

```bash
bluetuith --adapter=hci0
```

## list-adapters

This option can be used to list the available bluetooth adapters.

## adapter-states

This option can be used to set various adapter properties and states on initialization.

Valid properties are: `powered`, `scan`, `discoverable`, `pairable`.<br/>
Valid states are: `yes/y/on` to **enable** a property and `no/n/off` to **disable** a property.

The provided value must be in the **[\<property\>:\<state\>]** format.

!!! note "Property sequence"
    Each property will be parsed and its state set based on the order in which you provide the properties.<br/><br/>
    For example, if `discoverable:yes, powered:yes` is provided:<br/>
        - The 'discoverable' state will be set first and<br/>
        - The 'powered' state will be set after it.<br/>

For example:

```bash
bluetuith --adapter=hci0 --adapter-states="powered:yes, discoverable:yes, pairable:yes, scan:no"
```

## connect-bdaddr

This option can be used to connect to a device based on its address.

For example:

```bash
bluetuith --connect-bdaddr="AA:BB:CC:DD:EE:FF"
```

## confirm-on-quit

This option can be used to show a confirmation message before quitting the application.

## no-warning

This option can be used to hide warnings when the application has initialized.

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

```bash
bluetuith --theme='{ Adapter: "red" }'
```

To see the available element types and colors, use the `--help` option.

## generate

This option can be used to generate the configuration.

Note that if you are regenerating the config, the existing values will be re-applied to the generated output.

## version

This option can be used to print the current version of the application.
