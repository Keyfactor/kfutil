package cmd

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_StoreTypesListCLI_ReturnsMoreThanOnePage(t *testing.T) {
	if testing.Short() || !hasIntegrationTestEnvironment() {
		t.Skip("requires Keyfactor Command integration environment")
	}
	defer resetRootCommandState()

	RootCmd.SetArgs([]string{"store-types", "list", "--no-prompt", "--format", "json"})
	output := captureOutput(func() {
		require.NoError(t, RootCmd.Execute())
	})

	var storeTypes []map[string]any
	require.NoError(t, json.Unmarshal([]byte(output), &storeTypes))
	require.Greater(t, len(storeTypes), 50, "store-types list should include results beyond the first default page")
}
