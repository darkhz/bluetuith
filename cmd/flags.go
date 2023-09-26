package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/darkhz/bluetuith/bluez"
	"github.com/darkhz/bluetuith/theme"
	"github.com/knadh/koanf/parsers/hjson"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/posflag"
	flag "github.com/spf13/pflag"
)

// Option describes a command-line option.
type Option struct {
	Name, Description, Value string
	IsBoolean                bool
}

var options = []Option{
	{
		Name:        "list-adapters",
		Description: "List available adapters.",
		IsBoolean:   true,
	},
	{
		Name:        "adapter",
		Description: "Specify an adapter to use. (For example, hci0)",
	},
	{
		Name:        "receive-dir",
		Description: "Specify a directory to store received files.",
	},
	{
		Name:        "gsm-apn",
		Description: "Specify GSM APN to connect to. (Required for DUN)",
	},
	{
		Name:        "gsm-number",
		Description: "Specify GSM number to dial. (Required for DUN)",
	},
	{
		Name:        "theme",
		Description: "Specify a theme in the HJSON format. (For example, '{ Adapter: \"red\" }')",
	},
	{
		Name:        "no-warning",
		Description: "Do not display warnings when the application has initialized.",
		IsBoolean:   true,
	},
	{
		Name:        "generate",
		Description: "Generate configuration.",
		IsBoolean:   true,
	},
	{
		Name:        "version",
		Description: "Print version information.",
		IsBoolean:   true,
	},
}

func parse() {
	configFile, err := ConfigPath("bluetuith.conf")
	if err != nil {
		PrintError("Cannot get config directory")
	}

	fs := flag.NewFlagSet("bluetuith", flag.ContinueOnError)
	fs.Usage = func() {
		var usage string

		usage += fmt.Sprintf(
			"bluetuith [<flags>]\n\nConfig file is %s\n\nFlags:\n",
			configFile,
		)

		fs.VisitAll(func(f *flag.Flag) {
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
			}

			if len(s) <= 4 {
				s += "\t"
			} else {
				s += "\n    \t"
			}

			s += strings.ReplaceAll(f.Usage, "\n", "\n    \t")

			usage += s + "\n"
		})

		usage += "\n" + theme.GetElementData()

		Print(usage, 0)
	}

	for _, option := range options {
		if option.IsBoolean {
			fs.Bool(option.Name, false, option.Description)
			continue
		}

		fs.String(option.Name, option.Value, option.Description)
	}

	if err = fs.Parse(os.Args[1:]); err != nil {
		PrintError(err.Error())
	}

	if err := config.Load(file.Provider(configFile), hjson.Parser()); err != nil {
		PrintError(err.Error())
	}

	if err := config.Load(posflag.Provider(fs, ".", config.Koanf), nil); err != nil {
		PrintError(err.Error())
	}
}

func cmdOptionAdapter(b *bluez.Bluez) {
	optionAdapter := GetProperty("adapter")
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

	PrintError(optionAdapter + ": The adapter does not exist.")
}

func cmdOptionListAdapters(b *bluez.Bluez) {
	var adapters string

	if !IsPropertyEnabled("list-adapters") {
		return
	}

	adapters += "List of adapters:\n"
	for _, adapter := range b.GetAdapters() {
		adapters += "- " + filepath.Base(adapter.Path) + "\n"
	}

	Print(strings.TrimRight(adapters, "\n"), 0)
}

func cmdOptionReceiveDir() {
	optionReceiveDir := GetProperty("receive-dir")
	if optionReceiveDir == "" {
		return
	}

	if statpath, err := os.Stat(optionReceiveDir); err == nil && statpath.IsDir() {
		AddProperty("receive-dir", optionReceiveDir)
		return
	}

	PrintError(optionReceiveDir + ": Directory is not accessible.")
}

func cmdOptionGsm() {
	optionGsmNumber := GetProperty("gsm-number")
	optionGsmApn := GetProperty("gsm-apn")

	if optionGsmNumber == "" && optionGsmApn != "" {
		PrintError("Specify GSM Number.")
	}

	number := "*99#"
	if optionGsmNumber != "" {
		number = optionGsmNumber
	}

	AddProperty("gsm-apn", optionGsmApn)
	AddProperty("gsm-number", number)
}

func cmdOptionTheme() {
	if !config.Exists("theme") {
		return
	}

	optionTheme := config.Get("theme")
	if t, ok := optionTheme.(string); ok {
		themeConfig, err := hjson.Parser().Unmarshal([]byte(t))
		if err != nil {
			PrintError("Provided theme format is invalid", err)
		}

		optionTheme = themeConfig
	}

	config.Set("theme", optionTheme)

	themeMap := config.StringMap("theme")
	if len(themeMap) == 0 {
		return
	}

	if err := theme.ParseThemeConfig(themeMap); err != nil {
		PrintError(err.Error())
	}
}

func cmdOptionGenerate() {
	optionGenerate := IsPropertyEnabled("generate")
	if !optionGenerate {
		return
	}

	generate()

	os.Exit(0)
}

func cmdOptionVersion() {
	optionVersion := IsPropertyEnabled("version")
	if !optionVersion {
		return
	}

	text := "Bluetuith v%s"

	versionInfo := strings.Split(Version, "@")
	if len(versionInfo) < 2 {
		Print(fmt.Sprintf(text, versionInfo), 0)
	}

	text += " (%s)"
	Print(fmt.Sprintf(text, versionInfo[0], versionInfo[1]), 0)
}
