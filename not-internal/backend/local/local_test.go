package local

import (
	"flag"
	"os"
	"testing"

	_ "github.com/muratcelep/terraform/not-internal/logging"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
