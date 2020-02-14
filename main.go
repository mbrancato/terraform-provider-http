package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/mbrancato/terraform-provider-http/http"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: http.Provider})
}
