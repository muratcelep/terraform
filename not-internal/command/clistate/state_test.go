package clistate

import (
	"testing"

	"github.com/muratcelep/terraform/not-internal/command/arguments"
	"github.com/muratcelep/terraform/not-internal/command/views"
	"github.com/muratcelep/terraform/not-internal/states/statemgr"
	"github.com/muratcelep/terraform/not-internal/terminal"
)

func TestUnlock(t *testing.T) {
	streams, _ := terminal.StreamsForTesting(t)
	view := views.NewView(streams)

	l := NewLocker(0, views.NewStateLocker(arguments.ViewHuman, view))
	l.Lock(statemgr.NewUnlockErrorFull(nil, nil), "test-lock")

	diags := l.Unlock()
	if diags.HasErrors() {
		t.Log(diags.Err().Error())
	} else {
		t.Error("expected error")
	}
}
