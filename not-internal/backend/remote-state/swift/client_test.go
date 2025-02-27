package swift

import (
	"fmt"
	"testing"
	"time"

	"github.com/muratcelep/terraform/not-internal/backend"
	"github.com/muratcelep/terraform/not-internal/states/remote"
)

func TestRemoteClient_impl(t *testing.T) {
	var _ remote.Client = new(RemoteClient)
}

func TestRemoteClient(t *testing.T) {
	testACC(t)

	container := fmt.Sprintf("terraform-state-swift-testclient-%x", time.Now().Unix())

	b := backend.TestBackendConfig(t, New(), backend.TestWrapConfig(map[string]interface{}{
		"container": container,
	})).(*Backend)

	state, err := b.StateMgr(backend.DefaultStateName)
	if err != nil {
		t.Fatal(err)
	}

	client := &RemoteClient{
		client:    b.client,
		container: b.container,
	}

	defer client.deleteContainer()

	remote.TestClient(t, state.(*remote.State).Client)
}
