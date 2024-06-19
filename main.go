package main

import (
	"github.com/itzloop/plugin-azure/plugin"
	"github.com/kaytu-io/kaytu/pkg/plugin/sdk"
)

func main() {
    azplugin := plugin.NewAzurePlugin()
    sdk.New(azplugin, 4).Execute()
    
}
