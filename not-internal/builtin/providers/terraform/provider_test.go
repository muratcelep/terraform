package terraform

import (
	backendInit "github.com/muratcelep/terraform/not-internal/backend/init"
)

func init() {
	// Initialize the backends
	backendInit.Init(nil)
}
