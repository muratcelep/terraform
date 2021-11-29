package main

import (
	"github.com/muratcelep/terraform/not-internal/grpcwrap"
	plugin "github.com/muratcelep/terraform/not-internal/plugin6"
	simple "github.com/muratcelep/terraform/not-internal/provider-simple-v6"
	"github.com/muratcelep/terraform/not-internal/tfplugin6"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin6.ProviderServer {
			return grpcwrap.Provider6(simple.Provider())
		},
	})
}
