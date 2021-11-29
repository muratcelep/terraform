package etcdv2

import (
	"testing"

	"github.com/hashicorp/terraform/not_internal/backend"
)

func TestBackend_impl(t *testing.T) {
	var _ backend.Backend = new(Backend)
}
