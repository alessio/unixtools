package path_test

import (
	"github.com/alessio/unixtools/internal/path"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestList_Prepend(t *testing.T) {
	lst := path.NewDirList()
	lst.SetDirs("/var", "/root/config", "/Programs///")

	require.Equal(t, "/var:/root/config:/Programs", lst.String())

	require.True(t, lst.Prepend("/usr/local/go/bin"))
	require.False(t, lst.Prepend("/usr/local/go/bin"))
	require.False(t, lst.Prepend("/usr///local///go/bin/"))

	require.Equal(t, "/usr/local/go/bin:/var:/root/config:/Programs", lst.String())
	require.Equal(t, []string{"/usr/local/go/bin", "/var", "/root/config", "/Programs"}, lst.Slice())
}

func TestList_Append(t *testing.T) {
	lst := path.NewDirList()
	lst.SetDirs("/var", "/root/config", "/Programs///")

	require.Equal(t, "/var:/root/config:/Programs", lst.String())

	require.True(t, lst.Append("/usr/local/go/bin"))
	require.False(t, lst.Append("/usr/local/go/bin"))
	require.False(t, lst.Append("/usr///local///go/bin/"))

	require.Equal(t, "/var:/root/config:/Programs:/usr/local/go/bin", lst.String())
	require.Equal(t, []string{"/var", "/root/config", "/Programs", "/usr/local/go/bin"}, lst.Slice())
}

func TestList_Drop(t *testing.T) {
	lst := path.NewDirList()
	lst.SetDirs(strings.Split("/usr/local/bin:/home/user/.local/bin/:/usr/local/sbin:/var:/root", ":")...)

	require.Equal(t, "/usr/local/bin:/home/user/.local/bin:/usr/local/sbin:/var:/root", lst.String())
	require.False(t, lst.Drop("/etc")) // non existing
	require.True(t, lst.Drop("/home/user/.local/bin"))
	require.False(t, lst.Drop("/home/user/.local/bin"))
	require.True(t, lst.Drop("/root/./"))

	require.Equal(t, "/usr/local/bin:/usr/local/sbin:/var", lst.String())
	require.Equal(t, []string{"/usr/local/bin", "/usr/local/sbin", "/var"}, lst.Slice())
}

func TestList_String(t *testing.T) {
	tests := []struct {
		name string
		lst  string
		want string
	}{
		{"simple", "/usr/local/bin:/usr/sbin", "/usr/local/bin:/usr/sbin"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(ttt *testing.T) {
			p := path.NewDirList()
			p.SetDirs(strings.Split(tt.lst, ":")...)
			if got := p.String(); got != tt.want {
				ttt.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
