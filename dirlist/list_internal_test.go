package dirlist

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newDirList(t *testing.T) {
	d := new(dirList)
	d.Reset()
	require.NotNil(t, d)
	require.NotNil(t, d)
	require.Equal(t, "", d.String())

	d.Append("/ciao")
	require.Equal(t, "/ciao", d.String())

	d1 := new(dirList)
	d1.Reset()
	d1 = d.clone(d1)
	d.Append("/var")
	require.NotEqual(t, &d, &d1)
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
