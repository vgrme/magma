/*
 Copyright (c) Facebook, Inc. and its affiliates.
 All rights reserved.

 This source code is licensed under the BSD-style license found in the
 LICENSE file in the root directory of this source tree.
*/

// NOTE: to run these tests outside the testing environment, e.g. from IntelliJ, ensure
// Postgres container is running, and use the DATABASE_SOURCE environment variable
// to target localhost and non-standard port.
// Example: `host=localhost port=5433 dbname=magma_test user=magma_test password=magma_test sslmode=disable`.

package reindex

import (
	"context"
	"fmt"
	"testing"

	"magma/orc8r/cloud/go/clock"
	"magma/orc8r/cloud/go/orc8r"
	"magma/orc8r/cloud/go/plugin"
	"magma/orc8r/cloud/go/pluginimpl"
	"magma/orc8r/cloud/go/pluginimpl/models"
	"magma/orc8r/cloud/go/serde"
	configurator_test_init "magma/orc8r/cloud/go/services/configurator/test_init"
	configurator_test "magma/orc8r/cloud/go/services/configurator/test_utils"
	device_test_init "magma/orc8r/cloud/go/services/device/test_init"
	"magma/orc8r/cloud/go/services/directoryd"
	"magma/orc8r/cloud/go/services/state"
	"magma/orc8r/cloud/go/services/state/indexer"
	"magma/orc8r/cloud/go/services/state/indexer/mocks"
	"magma/orc8r/cloud/go/services/state/servicers"
	state_test_init "magma/orc8r/cloud/go/services/state/test_init"
	state_test "magma/orc8r/cloud/go/services/state/test_utils"
	"magma/orc8r/cloud/go/sqorc"
	"magma/orc8r/lib/go/definitions"
	"magma/orc8r/lib/go/protos"

	"github.com/stretchr/testify/mock"
	assert "github.com/stretchr/testify/require"
)

const (
	singleAttempt = 1

	// Cause 3 batches per network
	numBatches       = numNetworks * 3
	numNetworks      = 3
	statesPerNetwork = 2*numStatesToReindexPerCall + 1
)

var (
	matchAll = []indexer.Subscription{{Type: orc8r.DirectoryRecordType, KeyMatcher: indexer.MatchAll}}
	matchOne = []indexer.Subscription{{Type: orc8r.DirectoryRecordType, KeyMatcher: indexer.NewMatchExact("imsi0")}}
)

func TestRun(t *testing.T) {
	ch := make(chan interface{})
	// Writes to channel after completing a job
	testHookReindexComplete = func() {
		ch <- nil
	}
	clock.SkipSleeps(t)
	defer clock.ResumeSleeps(t)

	q, store := initReindexTest(t)
	go Run(q, store)

	// Single indexer
	// Populate
	idx0 := getIndexer(id0, zero, version0, true)
	idx0.On("GetSubscriptions").Return(matchAll).Once()
	registerAndPopulate(t, q, idx0)
	// Check
	<-ch
	idx0.AssertExpectations(t)
	assertComplete(t, q, id0)

	// Bump existing indexer version
	// Populate
	idx0a := getIndexerNoIndex(id0, version0, version0a, false)
	idx0a.On("GetSubscriptions").Return(matchOne).Once()
	idx0a.On("Index", mock.Anything, mock.Anything).Return(nil, nil).Times(numNetworks)
	registerAndPopulate(t, q, idx0a)
	// Check
	<-ch
	idx0a.AssertExpectations(t)
	assertComplete(t, q, id0)

	// Indexer returns err => reindex jobs fail
	// Populate
	// Fail1 at PrepareReindex
	fail1 := getBasicIndexer(id1, version1)
	fail1.On("GetSubscriptions").Return(matchAll).Once()
	fail1.On("PrepareReindex", zero, version1, true).Return(someErr1).Once()
	// Fail2 at first Reindex
	fail2 := getBasicIndexer(id2, version2)
	fail2.On("GetSubscriptions").Return(matchAll).Once()
	fail2.On("PrepareReindex", zero, version2, true).Return(nil).Once()
	fail2.On("Index", mock.Anything, mock.Anything).Return(nil, someErr2).Once()
	// Fail3 at CompleteReindex
	fail3 := getBasicIndexer(id3, version3)
	fail3.On("GetSubscriptions").Return(matchAll).Once()
	fail3.On("PrepareReindex", zero, version3, true).Return(nil).Once()
	fail3.On("Index", mock.Anything, mock.Anything).Return(nil, nil).Times(numBatches)
	fail3.On("CompleteReindex", zero, version3).Return(someErr3).Once()
	registerAndPopulate(t, q, fail1, fail2, fail3)
	// Check
	<-ch
	<-ch
	<-ch
	fail1.AssertExpectations(t)
	fail2.AssertExpectations(t)
	fail3.AssertExpectations(t)
	assertErrored(t, q, id1, ErrPrepare, someErr1)
	assertErrored(t, q, id2, ErrReindex, someErr2)
	assertErrored(t, q, id3, ErrComplete, someErr3)
}

func initReindexTest(t *testing.T) (JobQueue, servicers.StateServiceInternal) {
	// Uncomment below to view reindex queue logs during test
	//_ = flag.Set("alsologtostderr", "true")

	// Start configurator service, add networks and gateways
	assert.NoError(t, plugin.RegisterPluginForTests(t, &pluginimpl.BaseOrchestratorPlugin{}))
	configurator_test_init.StartTestService(t)
	device_test_init.StartTestService(t)
	configurator_test.RegisterNetwork(t, nid0, "Network 0 for reindex test")
	configurator_test.RegisterNetwork(t, nid1, "Network 1 for reindex test")
	configurator_test.RegisterNetwork(t, nid2, "Network 2 for reindex test")
	configurator_test.RegisterGateway(t, nid0, hwid0, &models.GatewayDevice{HardwareID: hwid0})
	configurator_test.RegisterGateway(t, nid1, hwid1, &models.GatewayDevice{HardwareID: hwid1})
	configurator_test.RegisterGateway(t, nid2, hwid2, &models.GatewayDevice{HardwareID: hwid2})

	// Start state service, add states
	store := state_test_init.StartTestServiceInternal(t)
	ctxByNetwork := map[string]context.Context{
		nid0: state_test.GetContextWithCertificate(t, hwid0),
		nid1: state_test.GetContextWithCertificate(t, hwid1),
		nid2: state_test.GetContextWithCertificate(t, hwid2),
	}
	for _, nid := range []string{nid0, nid1, nid2} {
		var records []*directoryd.DirectoryRecord
		var deviceIDs []string
		for i := 0; i < statesPerNetwork; i++ {
			hwid := fmt.Sprintf("hwid%d", i)
			imsi := fmt.Sprintf("imsi%d", i)
			records = append(records, &directoryd.DirectoryRecord{LocationHistory: []string{hwid}})
			deviceIDs = append(deviceIDs, imsi)
		}
		reportStates(t, ctxByNetwork[nid], deviceIDs, records)
	}

	sqlDriver := definitions.GetEnvWithDefault("SQL_DRIVER", "postgres")
	databaseSource := definitions.GetEnvWithDefault("DATABASE_SOURCE", connectionStringPostgres)
	db, err := sqorc.Open(sqlDriver, databaseSource)
	assert.NoError(t, err)

	_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", queueTableName))
	assert.NoError(t, err)
	_, err = db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", versionTableName))
	assert.NoError(t, err)

	queue := NewSQLJobQueue(singleAttempt, db, sqorc.GetSqlBuilder())
	err = queue.Initialize()
	assert.NoError(t, err)
	return queue, store
}

func reportStates(t *testing.T, ctx context.Context, deviceIDs []string, records []*directoryd.DirectoryRecord) {
	client, err := state.GetStateClient()
	assert.NoError(t, err)

	var states []*protos.State
	for i, st := range records {
		serialized, err := serde.Serialize(state.SerdeDomain, orc8r.DirectoryRecordType, st)
		assert.NoError(t, err)
		pState := &protos.State{
			Type:     orc8r.DirectoryRecordType,
			DeviceID: deviceIDs[i],
			Value:    serialized,
		}
		states = append(states, pState)
	}
	_, err = client.ReportStates(
		ctx,
		&protos.ReportStatesRequest{States: states},
	)
	assert.NoError(t, err)
}

func getBasicIndexer(id string, v indexer.Version) *mocks.Indexer {
	idx := &mocks.Indexer{}
	idx.On("GetID").Return(id)
	idx.On("GetVersion").Return(v)
	return idx
}

func getIndexerNoIndex(id string, from, to indexer.Version, isFirstReindex bool) *mocks.Indexer {
	idx := &mocks.Indexer{}
	idx.On("GetID").Return(id)
	idx.On("GetVersion").Return(to)
	idx.On("PrepareReindex", from, to, isFirstReindex).Return(nil).Once()
	idx.On("CompleteReindex", from, to).Return(nil).Once()
	return idx
}

func getIndexer(id string, from, to indexer.Version, isFirstReindex bool) *mocks.Indexer {
	idx := &mocks.Indexer{}
	idx.On("GetID").Return(id)
	idx.On("GetVersion").Return(to)
	idx.On("PrepareReindex", from, to, isFirstReindex).Return(nil).Once()
	idx.On("Index", mock.Anything, mock.Anything).Return(nil, nil).Times(numBatches)
	idx.On("CompleteReindex", from, to).Return(nil).Once()
	return idx
}

func registerAndPopulate(t *testing.T, q JobQueue, idx ...indexer.Indexer) {
	indexer.DeregisterAllForTest(t)
	err := indexer.RegisterAll(idx...)
	assert.NoError(t, err)
	populated, err := q.PopulateJobs()
	assert.True(t, populated)
	assert.NoError(t, err)
}

func assertComplete(t *testing.T, q JobQueue, id string) {
	st, err := GetStatus(q, id)
	assert.NoError(t, err)
	assert.Equal(t, StatusComplete, st)
	e, err := GetError(q, id)
	assert.NoError(t, err)
	assert.Empty(t, e)
}

func assertErrored(t *testing.T, q JobQueue, indexerID string, sentinel Error, rootErr error) {
	st, err := GetStatus(q, indexerID)
	assert.NoError(t, err)
	assert.Equal(t, StatusAvailable, st)
	e, err := GetError(q, indexerID)
	assert.NoError(t, err)
	// Job err contains relevant info
	assert.Contains(t, e, indexerID)
	assert.Contains(t, e, sentinel)
	assert.Contains(t, e, rootErr.Error())
}
