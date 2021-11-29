package main

import (
	"github.com/hashicorp/terraform/not_internal/builtin/providers/terraform"
	"github.com/hashicorp/terraform/not_internal/grpcwrap"
	"github.com/hashicorp/terraform/not_internal/plugin"
	"github.com/hashicorp/terraform/not_internal/tfplugin5"
)

func main() {
	// Provide a binary version of the not_internal terraform provider for testing
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin5.ProviderServer {
			return grpcwrap.Provider(terraform.NewProvider())
		},
	})
}
