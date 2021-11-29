package terraform

import (
	backendInit "github.com/hashicorp/terraform/not_internal/backend/init"
)

func init() {
	// Initialize the backends
	backendInit.Init(nil)
}
