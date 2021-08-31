package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-digitalocean/builder/digitalocean"
	digitaloceanPP "github.com/hashicorp/packer-plugin-digitalocean/post-processor/digitalocean-import"
	"github.com/hashicorp/packer-plugin-digitalocean/version"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder(plugin.DEFAULT_NAME, new(digitalocean.Builder))
	pps.RegisterPostProcessor("import", new(digitaloceanPP.PostProcessor))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
