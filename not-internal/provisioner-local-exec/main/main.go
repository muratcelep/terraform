package main

import (
	localexec "github.com/muratcelep/terraform/not-internal/builtin/provisioners/local-exec"
	"github.com/muratcelep/terraform/not-internal/grpcwrap"
	"github.com/muratcelep/terraform/not-internal/plugin"
	"github.com/muratcelep/terraform/not-internal/tfplugin5"
)

func main() {
	// Provide a binary version of the not-internal terraform provider for testing
	plugin.Serve(&plugin.ServeOpts{
		GRPCProvisionerFunc: func() tfplugin5.ProvisionerServer {
			return grpcwrap.Provisioner(localexec.New())
		},
	})
}
