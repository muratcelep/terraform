package main

import (
	"github.com/hashicorp/terraform/not_internal/grpcwrap"
	plugin "github.com/hashicorp/terraform/not_internal/plugin6"
	simple "github.com/hashicorp/terraform/not_internal/provider-simple-v6"
	"github.com/hashicorp/terraform/not_internal/tfplugin6"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin6.ProviderServer {
			return grpcwrap.Provider6(simple.Provider())
		},
	})
}
