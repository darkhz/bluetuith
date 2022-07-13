package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/darkhz/bluetuith/bluez"
)

var (
	optionAdapter      string
	optionListAdapters bool
)

func ParseCmdFlags(bluezConn *bluez.Bluez) {
	flag.BoolVar(&optionListAdapters, "list-adapters", false, "List available adapters.")
	flag.StringVar(&optionAdapter, "adapter", "", "Specify an adapter to use. (For example, hci0)")

	flag.Usage = func() {
		fmt.Fprintf(
			flag.CommandLine.Output(),
			"bluetuith [<flags>]\n\nFlags:\n",
		)

		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			s := fmt.Sprintf("  --%s", f.Name)

			if f.Name == "adapter" {
				s += " <adapter>"
			}

			if len(s) <= 4 {
				s += "\t"
			} else {
				s += "\n    \t"
			}

			s += strings.ReplaceAll(f.Usage, "\n", "\n    \t")

			fmt.Fprint(flag.CommandLine.Output(), s, "\n")
		})
	}

	flag.Parse()

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
