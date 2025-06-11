package hostlist

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/grafviktor/goto/internal/constant"
	"github.com/grafviktor/goto/internal/model/host"
)

type dummyItem struct{}

func (d dummyItem) FilterValue() string {
	return ""
}

func Test_keyMap_UpdateKeyVisibility(t *testing.T) {
	km := newDelegateKeyMap()

	// Case 1: item is nil (not ListItemHost)
	state := km.UpdateKeyVisibility(nil)
	require.Equal(t, string(keyMapState.EditkeysHidden), state)
	require.Equal(t, keyMapState.EditkeysHidden, km.keyMapState)

	// Case 2: item is ListItemHost and IsReadOnly() == true
	readonlyHost := ListItemHost{Host: host.Host{StorageType: constant.HostStorageType.SSHConfig}}
	state = km.UpdateKeyVisibility(readonlyHost)
	require.Equal(t, string(keyMapState.EditkeysPartiallyShown), state)
	require.Equal(t, keyMapState.EditkeysPartiallyShown, km.keyMapState)

	// Case 3: item is ListItemHost and IsReadOnly() == false
	writableHost := ListItemHost{Host: host.Host{StorageType: constant.HostStorageType.YAMLFile}}
	state = km.UpdateKeyVisibility(writableHost)
	require.Equal(t, string(keyMapState.EditkeysShown), state)
	require.Equal(t, keyMapState.EditkeysShown, km.keyMapState)

	// Case 4: item is not ListItemHost but implements list.Item
	state = km.UpdateKeyVisibility(dummyItem{})
	require.Equal(t, string(keyMapState.EditkeysHidden), state)
	require.Equal(t, keyMapState.EditkeysHidden, km.keyMapState)
}
