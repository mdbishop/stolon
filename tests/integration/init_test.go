// Copyright 2015 Sorint.lab
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied
// See the License for the specific language governing permissions and
// limitations under the License.

package integration

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sorintlab/stolon/common"
	"github.com/sorintlab/stolon/pkg/store"

	"github.com/sorintlab/stolon/Godeps/_workspace/src/github.com/satori/go.uuid"
)

func TestInit(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer os.RemoveAll(dir)

	tstore, err := NewTestStore(t, dir)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tstore.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tstore.WaitUp(10 * time.Second); err != nil {
		t.Fatalf("error waiting on store up: %v", err)
	}
	storeEndpoints := fmt.Sprintf("%s:%s", tstore.listenAddress, tstore.port)
	defer tstore.Stop()

	clusterName := uuid.NewV4().String()

	ts, err := NewTestSentinel(t, dir, clusterName, tstore.storeBackend, storeEndpoints)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := ts.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer ts.Stop()
	tk, err := NewTestKeeper(t, dir, clusterName, tstore.storeBackend, storeEndpoints)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}

	if err := tk.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer tk.Stop()

	if err := tk.WaitDBUp(60 * time.Second); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	t.Logf("database is up")

}

func TestInitialClusterConfig(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer os.RemoveAll(dir)

	tstore, err := NewTestStore(t, dir)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tstore.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tstore.WaitUp(10 * time.Second); err != nil {
		t.Fatalf("error waiting on store up: %v", err)
	}
	defer tstore.Stop()

	clusterName := uuid.NewV4().String()

	storeEndpoints := fmt.Sprintf("%s:%s", tstore.listenAddress, tstore.port)
	storePath := filepath.Join(common.StoreBasePath, clusterName)

	kvstore, err := store.NewStore(tstore.storeBackend, storeEndpoints)
	if err != nil {
		t.Fatalf("cannot create store: %v", err)
	}

	e := store.NewStoreManager(kvstore, storePath)

	tmpFile, err := ioutil.TempFile(dir, "initial-cluster-config.json")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer tmpFile.Close()
	tmpFile.WriteString(`{ "synchronous_replication": true }`)

	ts, err := NewTestSentinel(t, dir, clusterName, tstore.storeBackend, storeEndpoints, fmt.Sprintf("--initial-cluster-config=%s", tmpFile.Name()))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := ts.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer ts.Stop()

	if err := WaitClusterInitialized(e, 30*time.Second); err != nil {
		t.Fatal("expected cluster initialized")
	}

	cv, _, err := e.GetClusterView()
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !*cv.Config.SynchronousReplication {
		t.Fatal("expected cluster config with InitWithMultipleKeepers enabled")
	}
}

func TestExclusiveLock(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer os.RemoveAll(dir)

	tstore, err := NewTestStore(t, dir)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tstore.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tstore.WaitUp(10 * time.Second); err != nil {
		t.Fatalf("error waiting on store up: %v", err)
	}
	storeEndpoints := fmt.Sprintf("%s:%s", tstore.listenAddress, tstore.port)
	defer tstore.Stop()

	clusterName := uuid.NewV4().String()

	u := uuid.NewV4()
	id := fmt.Sprintf("%x", u[:4])

	tk1, err := NewTestKeeperWithID(t, dir, id, clusterName, tstore.storeBackend, storeEndpoints)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tk1.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer tk1.Stop()

	// Wait for tk1 up before starting tk2
	if err := tk1.WaitUp(10 * time.Second); err != nil {
		t.Fatalf("expecting tk1 up but it's down")
	}

	tk2, err := NewTestKeeperWithID(t, dir, id, clusterName, tstore.storeBackend, storeEndpoints)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tk2.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer tk2.Stop()

	// tk2 should exit because it cannot take an exclusive lock on dataDir
	if err := tk2.Wait(10 * time.Second); err != nil {
		t.Fatalf("expecting tk2 exiting due to failed exclusive lock, but it's active.")
	}

}

func TestKeeperPGConfDirBad(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer os.RemoveAll(dir)

	clusterName := uuid.NewV4().String()

	tstore, err := NewTestStore(t, dir)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	storeEndpoints := fmt.Sprintf("%s:%s", tstore.listenAddress, tstore.port)

	// Test pgConfDir not absolute path
	tk, err := NewTestKeeper(t, dir, clusterName, tstore.storeBackend, storeEndpoints, "--pg-conf-dir=not/absolute/path")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tk.StartExpect(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer tk.Stop()
	if err := tk.cmd.Expect("keeper: pg-conf-dir must be an absolute path"); err != nil {
		t.Fatalf("expecting keeper reporting error due to pg-conf-dir provided as a non absolute path")
	}

	// Test unexistent pgConfDir
	tk2, err := NewTestKeeper(t, dir, clusterName, tstore.storeBackend, storeEndpoints, "--pg-conf-dir=/unexistent-configuration-directory")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tk2.StartExpect(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer tk2.Stop()
	if err := tk2.cmd.Expect("keeper: cannot stat pg-conf-dir:"); err != nil {
		t.Fatalf("expecting keeper reporting error due to unexistent pg-conf-dir")
	}

	// Test pgConfDir is a file
	tmpFile, err := ioutil.TempFile(dir, "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()
	tk3, err := NewTestKeeper(t, dir, clusterName, tstore.storeBackend, storeEndpoints, fmt.Sprintf("--pg-conf-dir=%s", tmpFile.Name()))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tk3.StartExpect(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer tk3.Stop()
	if err := tk3.cmd.Expect("keeper: pg-conf-dir is not a directory"); err != nil {
		t.Fatalf("expecting keeper reporting error due to pg-conf-dir being a file")
	}
}

func TestKeeperPGConfDirGood(t *testing.T) {
	t.Parallel()

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer os.RemoveAll(dir)

	clusterName := uuid.NewV4().String()

	tstore, err := NewTestStore(t, dir)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tstore.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tstore.WaitUp(10 * time.Second); err != nil {
		t.Fatalf("error waiting on store up: %v", err)
	}
	storeEndpoints := fmt.Sprintf("%s:%s", tstore.listenAddress, tstore.port)
	defer tstore.Stop()

	ts, err := NewTestSentinel(t, dir, clusterName, tstore.storeBackend, storeEndpoints)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := ts.Start(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer ts.Stop()

	// Test good pgConfDir.
	tmpDir, err := ioutil.TempDir(dir, "")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer os.Remove(tmpDir)

	tk, err := NewTestKeeper(t, dir, clusterName, tstore.storeBackend, storeEndpoints, fmt.Sprintf("--pg-conf-dir=%s", tmpDir))
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if err := tk.StartExpect(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	defer tk.Stop()
	if err := tk.cmd.ExpectTimeout("keeper: masterID: \""+tk.id+"\"", 60*time.Second); err != nil {
		t.Fatalf("expecting keeper active and being elected as master")
	}
}
