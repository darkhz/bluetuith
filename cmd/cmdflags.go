package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/theme"
	"github.com/jnovack/flag"
)

var (
	optionAdapter      string
	optionListAdapters bool

	optionReceiveDir string

	optionGsmApn    string
	optionGsmNumber string

	optionTheme       string
	optionColorConfig string
)

func ParseCmdFlags(bluezConn *bluez.Bluez) {
	configFile, err := ConfigPath("config")
	if err != nil {
		fmt.Println("Cannot get config directory")
		return
	}

	flag.BoolVar(
		&optionListAdapters,
		"list-adapters", false,
		"List available adapters.",
	)
	flag.StringVar(
		&optionAdapter,
		"adapter", "",
		"Specify an adapter to use. (For example, hci0)",
	)
	flag.StringVar(
		&optionReceiveDir,
		"receive-dir", "",
		"Specify a directory to store received files.",
	)
	flag.StringVar(
		&optionGsmApn,
		"gsm-apn", "",
		"Specify GSM APN to connect to. (Required for DUN)",
	)
	flag.StringVar(
		&optionGsmNumber,
		"gsm-number", "",
		"Specify GSM number to dial. (Required for DUN)",
	)
	flag.StringVar(
		&optionTheme,
		"set-theme", "",
		"Specify a theme."+theme.GetThemes(),
	)
	flag.StringVar(
		&optionColorConfig,
		"set-theme-config", "",
		"Specify a comma-separated list of custom colors for modifier elements."+
			"\n\nAvailable modifiers:\n"+theme.GetElementModifiers()+
			"\n\nAvailable colors:\n"+theme.GetElementColors()+
			"\n\nFor example: --set-theme-config='Adapter=red,Device=purple,MenuBar=blue'",
	)

	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"bluetuith [<flags>]\n\nConfig file is %s\n\nFlags:\n",
			configFile,
		)

		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			s := fmt.Sprintf("  --%s", f.Name)

			switch f.Name {
			case "adapter":
				s += " <adapter>"

			case "receive-dir":
				s += " <dir>"

			case "gsm-apn":
				s += " <apn>"

			case "gsm-number":
				s += " <number>"

			case "set-theme":
				s += " <theme>"

			case "set-theme-config":
				s += " [modifier1=color1,modifier2=color2,...]"
			}

			if len(s) <= 4 {
				s += "\t"
			} else {
				s += "\n    \t"
			}

			s += strings.ReplaceAll(f.Usage, "\n", "\n    \t")

			fmt.Fprint(flag.CommandLine.Output(), s, "\n\n")
		})
	}

	flag.CommandLine.ParseFile(configFile)
	flag.Parse()

	cmdOptionReceiveDir()

	cmdOptionGsm()

	cmdOptionTheme()
	cmdOptionThemeConfig()

	cmdOptionAdapter(bluezConn)
	cmdOptionListAdapters(bluezConn)
}

func cmdOptionAdapter(b *bluez.Bluez) {
	if optionAdapter == "" {
		b.SetCurrentAdapter()
		return
	}

	for _, adapter := range b.GetAdapters() {
		if optionAdapter == filepath.Base(adapter.Path) {
			b.SetCurrentAdapter(adapter)
			return
		}
	}

	fmt.Println(optionAdapter + ": The adapter does not exist.")

	os.Exit(0)
}

func cmdOptionListAdapters(b *bluez.Bluez) {
	if !optionListAdapters {
		return
	}

	fmt.Println("List of adapters:")
	for _, adapter := range b.GetAdapters() {
		fmt.Println("- " + filepath.Base(adapter.Path))
	}

	os.Exit(0)
}

func cmdOptionReceiveDir() {
	if optionReceiveDir == "" {
		return
	}

	if statpath, err := os.Stat(optionReceiveDir); err == nil && statpath.IsDir() {
		AddConfigProperty("receive-dir", optionReceiveDir)
		return
	}

	fmt.Println(optionReceiveDir + ": Directory is not accessible.")

	os.Exit(0)
}

func cmdOptionGsm() {
	if optionGsmNumber == "" && optionGsmApn != "" {
		fmt.Println("Specify GSM Number.")
		os.Exit(0)
	}

	number := "*99#"
	if optionGsmNumber != "" {
		number = optionGsmNumber
	}

	AddConfigProperty("gsm-apn", optionGsmApn)
	AddConfigProperty("gsm-number", number)
}

func cmdOptionTheme() {
	if optionTheme == "" {
		return
	}

	if err := theme.ParseThemeFile(optionTheme); err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}

func cmdOptionThemeConfig() {
	if optionColorConfig == "" {
		return
	}

	if err := theme.ParseThemeConfig(optionColorConfig); err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
}
