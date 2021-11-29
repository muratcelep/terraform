package main

import (
	"github.com/muratcelep/terraform/not-internal/grpcwrap"
	"github.com/muratcelep/terraform/not-internal/plugin"
	simple "github.com/muratcelep/terraform/not-internal/provider-simple"
	"github.com/muratcelep/terraform/not-internal/tfplugin5"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		GRPCProviderFunc: func() tfplugin5.ProviderServer {
			return grpcwrap.Provider(simple.Provider())
		},
	})
}
