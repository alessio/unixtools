package dirlist

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newDirList(t *testing.T) {
	d := new(dirList)
	d.Reset()
	require.NotNil(t, d)
	require.Equal(t, "", d.String())

	d.Append("/sbin")
	require.Equal(t, "/sbin", d.String())

	d1 := new(dirList)
	d1.Reset()
	d1 = d.clone(d1)
	d.Append("/var")
	require.NotEqual(t, &d, &d1)
	d1.Prepend("/usr/bin")
	d1.Append("/usr/local/bin")
	require.Equal(t, "/usr/bin:/sbin:/usr/local/bin", d1.String())
}

func Test_removeDups(t *testing.T) {
	require.Equal(t,
		[]string{"alpha", "bravo", "charlie", "."},
		removeDups(
			[]string{"alpha", "bravo", "charlie", "bravo", "   ", "."},
			filterEmptyStrings,
		),
	)
}
