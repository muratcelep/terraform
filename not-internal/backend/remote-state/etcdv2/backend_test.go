package etcdv2

import (
	"testing"

	"github.com/muratcelep/terraform/not-internal/backend"
)

func TestBackend_impl(t *testing.T) {
	var _ backend.Backend = new(Backend)
}
