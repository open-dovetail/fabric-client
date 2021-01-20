/*
SPDX-License-Identifier: BSD-3-Clause-Open-MPI
*/

package plugin

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/open-dovetail/fabric-client/activity/request"
	"github.com/project-flogo/cli/common" // Flogo CLI support code
	"github.com/spf13/cobra"
)

var configFile string
var matcherFile string

func init() {
	configfabric.Flags().StringVarP(&configFile, "config", "c", "config.yaml", "specify the yaml file for Fabric network configuration")
	configfabric.Flags().StringVarP(&matcherFile, "matchers", "m", "", "specify the yaml file for entity matchers override")
	common.RegisterPlugin(configfabric)
}

var configfabric = &cobra.Command{
	Use:   "configfabric",
	Short: "embed fabric network config",
	Long:  "This plugin generates embedded network config for Fabric client apps",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Embed network config file", configFile)
		networkConfig, err := request.ReadFile(configFile)
		if err != nil {
			fmt.Printf("Failed to read network config %s: %+v\n", configFile, err)
			os.Exit(1)
		}
		var matchersConfig []byte
		if len(matcherFile) > 0 {
			if matchersConfig, err = request.ReadFile(matcherFile); err != nil {
				fmt.Printf("Failed to read matchers config %s: %+v\n", matcherFile, err)
			}
		}
		if err = createFabricGoFile(networkConfig, matchersConfig); err != nil {
			os.Exit(1)
		}
	},
}

func createFabricGoFile(networkConfig, matcherConfig []byte) error {
	matcher := ""
	if len(matcherConfig) > 0 {
		matcher = string(matcherConfig)
	}
	data := struct {
		Config  string
		Matcher string
	}{
		string(networkConfig),
		matcher,
	}

	embedSrcPath := filepath.Join("src", "fabric_network.go")
	f, err := os.Create(embedSrcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	return renderTemplate(f, tplFabricGoFile, &data)
}

func renderTemplate(w io.Writer, text string, data interface{}) error {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(text))
	return t.Execute(w, data)
}

var tplFabricGoFile = `// Do not change this file, it has been generated using flogo-cli
// If you change it and rebuild the application your changes might get lost
package main

import "github.com/open-dovetail/fabric-client/activity/request"

// embedded flogo app descriptor file
const fabricConfig string = ` + "`{{.Config}}`" + `
const fabricMatcher string = ` + "`{{.Matcher}}`" + `

func init () {
	request.InitializeNetwork([]byte(fabricConfig), []byte(fabricMatcher))
}
`
