package etcd

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/muratcelep/terraform/not-internal/backend"
	"github.com/muratcelep/terraform/not-internal/states/remote"
	"github.com/muratcelep/terraform/not-internal/states/statemgr"
)

func TestRemoteClient_impl(t *testing.T) {
	var _ remote.Client = new(RemoteClient)
}

func TestRemoteClient(t *testing.T) {
	prepareEtcdv3(t)
	defer cleanupEtcdv3(t)

	prefix := fmt.Sprintf("%s/%s/", keyPrefix, time.Now().Format(time.RFC3339))

	// Get the backend
	b := backend.TestBackendConfig(t, New(), backend.TestWrapConfig(map[string]interface{}{
		"endpoints": stringsToInterfaces(etcdv3Endpoints),
		"prefix":    prefix,
	}))

	// Grab the client
	state, err := b.StateMgr(backend.DefaultStateName)
	if err != nil {
		t.Fatalf("Error: %s.", err)
	}

	// Test
	remote.TestClient(t, state.(*remote.State).Client)
}

func TestEtcdv3_stateLock(t *testing.T) {
	prepareEtcdv3(t)
	defer cleanupEtcdv3(t)

	prefix := fmt.Sprintf("%s/%s/", keyPrefix, time.Now().Format(time.RFC3339))

	// Get the backend
	s1, err := backend.TestBackendConfig(t, New(), backend.TestWrapConfig(map[string]interface{}{
		"endpoints": stringsToInterfaces(etcdv3Endpoints),
		"prefix":    prefix,
	})).StateMgr(backend.DefaultStateName)
	if err != nil {
		t.Fatal(err)
	}

	s2, err := backend.TestBackendConfig(t, New(), backend.TestWrapConfig(map[string]interface{}{
		"endpoints": stringsToInterfaces(etcdv3Endpoints),
		"prefix":    prefix,
	})).StateMgr(backend.DefaultStateName)
	if err != nil {
		t.Fatal(err)
	}

	remote.TestRemoteLocks(t, s1.(*remote.State).Client, s2.(*remote.State).Client)
}

func TestEtcdv3_destroyLock(t *testing.T) {
	prepareEtcdv3(t)
	defer cleanupEtcdv3(t)

	prefix := fmt.Sprintf("%s/%s/", keyPrefix, time.Now().Format(time.RFC3339))

	// Get the backend
	b := backend.TestBackendConfig(t, New(), backend.TestWrapConfig(map[string]interface{}{
		"endpoints": stringsToInterfaces(etcdv3Endpoints),
		"prefix":    prefix,
	}))

	// Grab the client
	s, err := b.StateMgr(backend.DefaultStateName)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	c := s.(*remote.State).Client.(*RemoteClient)

	info := statemgr.NewLockInfo()
	id, err := c.Lock(info)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Unlock(id); err != nil {
		t.Fatal(err)
	}

	res, err := c.Client.KV.Get(context.TODO(), c.info.Path)
	if err != nil {
		t.Fatal(err)
	}
	if res.Count != 0 {
		t.Fatalf("lock key not cleaned up at: %s", string(res.Kvs[0].Key))
	}
}
