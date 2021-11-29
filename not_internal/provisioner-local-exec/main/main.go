package main

import (
	localexec "github.com/hashicorp/terraform/not_internal/builtin/provisioners/local-exec"
	"github.com/hashicorp/terraform/not_internal/grpcwrap"
	"github.com/hashicorp/terraform/not_internal/plugin"
	"github.com/hashicorp/terraform/not_internal/tfplugin5"
)

func main() {
	// Provide a binary version of the not_internal terraform provider for testing
	plugin.Serve(&plugin.ServeOpts{
		GRPCProvisionerFunc: func() tfplugin5.ProvisionerServer {
			return grpcwrap.Provisioner(localexec.New())
		},
	})
}
