package main

import (
	"github.com/hashicorp/terraform/not_internal/grpcwrap"
	"github.com/hashicorp/terraform/not_internal/plugin"
	simple "github.com/hashicorp/terraform/not_internal/provider-simple"
	"github.com/hashicorp/terraform/not_internal/tfplugin5"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin5.ProviderServer {
			return grpcwrap.Provider(simple.Provider())
		},
	})
}
