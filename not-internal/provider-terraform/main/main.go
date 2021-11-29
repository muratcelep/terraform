package main

import (
	"github.com/muratcelep/terraform/not-internal/builtin/providers/terraform"
	"github.com/muratcelep/terraform/not-internal/grpcwrap"
	"github.com/muratcelep/terraform/not-internal/plugin"
	"github.com/muratcelep/terraform/not-internal/tfplugin5"
)

func main() {
	// Provide a binary version of the not-internal terraform provider for testing
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin5.ProviderServer {
			return grpcwrap.Provider(terraform.NewProvider())
		},
	})
}
