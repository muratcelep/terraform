package views

import (
	"strings"
	"testing"

	"github.com/muratcelep/terraform/not-internal/command/arguments"
	"github.com/muratcelep/terraform/not-internal/moduletest"
	"github.com/muratcelep/terraform/not-internal/terminal"
)

func TestTest(t *testing.T) {
	streams, close := terminal.StreamsForTesting(t)
	baseView := NewView(streams)
	view := NewTest(baseView, arguments.TestOutput{
		JUnitXMLFile: "",
	})

	results := map[string]*moduletest.Suite{}
	view.Results(results)

	output := close(t)
	gotOutput := strings.TrimSpace(output.All())
	wantOutput := `No tests defined. This module doesn't have any test suites to run.`
	if gotOutput != wantOutput {
		t.Errorf("wrong output\ngot:\n%s\nwant:\n%s", gotOutput, wantOutput)
	}

	// TODO: Test more at this layer. For now, the main UI output tests for
	// the "terraform test" command are in the command package as part of
	// the overall command tests.
}
