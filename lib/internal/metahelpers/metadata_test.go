package metahelpers

import (
	"context"
	"github.com/hanfei1991/microcosm/lib/common"
	"testing"

	"github.com/hanfei1991/microcosm/pkg/metadata"
	"github.com/stretchr/testify/require"
)


func TestMasterMetadata(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	metaKVClient := metadata.NewMetaMock()
	meta := []*common.MasterMetaKVData{
		{
			MasterMetaExt: &common.MasterMetaExt{
				ID: JobManagerUUID,
				Tp: common.JobManager,
			},
		},
		{
			MasterMetaExt: &common.MasterMetaExt{
				ID: "master-1",
				Tp: common.FakeJobMaster,
			},
		},
		{
			MasterMetaExt: &common.MasterMetaExt{
				ID: "master-2",
				Tp: common.FakeJobMaster,
			},
		},
	}
	for _, data := range meta {
		cli := NewMasterMetadataClient(data.MasterMetaExt.ID, metaKVClient)
		err := cli.Store(ctx, data)
		require.Nil(t, err)
	}
	cli := NewMasterMetadataClient("job-manager", metaKVClient)
	masters, err := cli.LoadAllMasters(ctx)
	require.Nil(t, err)
	require.Len(t, masters, 2)
	for _, master := range masters {
		require.Equal(t, common.FakeJobMaster, master.MasterMetaExt.Tp)
	}
}
