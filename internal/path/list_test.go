package path_test

import (
	"fmt"
	"github.com/alessio/unixtools/internal/path"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPathList_Prepend(t *testing.T) {
	envVarName := fmt.Sprintf("TEST_%s", t.Name())
	t.Setenv(envVarName, "/var/:/root/config:/Programs///")
	lst := path.NewPathList(envVarName)

	require.Equal(t, lst.String(), "/var:/root/config:/Programs")

	require.True(t, lst.Prepend("/usr/local/go/bin"))
	require.False(t, lst.Prepend("/usr/local/go/bin"))
	require.False(t, lst.Prepend("/usr///local///go/bin/"))

	require.Equal(t, "/usr/local/go/bin:/var:/root/config:/Programs", lst.String())
	require.Equal(t, []string{"/usr/local/go/bin", "/var", "/root/config", "/Programs"}, lst.StringSlice())
}

func TestPathList_Append(t *testing.T) {
	envVarName := fmt.Sprintf("TEST_%s", t.Name())
	t.Setenv(envVarName, "/var/:/root/config:/Programs///")
	lst := path.NewPathList(envVarName)

	require.Equal(t, "/var:/root/config:/Programs", lst.String())

	require.True(t, lst.Append("/usr/local/go/bin"))
	require.False(t, lst.Append("/usr/local/go/bin"))
	require.False(t, lst.Append("/usr///local///go/bin/"))

	require.Equal(t, "/var:/root/config:/Programs:/usr/local/go/bin", lst.String())
	require.Equal(t, []string{"/var", "/root/config", "/Programs", "/usr/local/go/bin"}, lst.StringSlice())
}

func TestPathList_Drop(t *testing.T) {
	envVarName := fmt.Sprintf("TEST_%s", t.Name())
	t.Setenv(envVarName,
		"/usr/local/bin:/home/user/.local/bin/:/usr/local/sbin:/var:/root")
	lst := path.NewPathList(envVarName)

	require.Equal(t, "/usr/local/bin:/home/user/.local/bin:/usr/local/sbin:/var:/root", lst.String())
	require.False(t, lst.Drop("/etc")) // non existing
	require.True(t, lst.Drop("/home/user/.local/bin"))
	require.False(t, lst.Drop("/home/user/.local/bin"))
	require.True(t, lst.Drop("/root/./"))

	require.Equal(t, "/usr/local/bin:/usr/local/sbin:/var", lst.String())
	require.Equal(t, []string{"/usr/local/bin", "/usr/local/sbin", "/var"}, lst.StringSlice())
}
