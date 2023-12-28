package pathlist

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_newDirList(t *testing.T) {
	d := new(dirList)
	d.Reset()
	require.NotNil(t, d)
	require.True(t, d.Nil())
	require.Equal(t, "", d.String())

	d.Append("/ciao")
	require.Equal(t, "/ciao", d.String())

	d1 := new(dirList)
	d1.Reset()
	d1 = d.clone(d1)
	d.Append("/var")
	require.NotEqual(t, &d, &d1)
}

func Test_DirList_Append(t *testing.T) {
	d := New()
	require.True(t, d.Nil())
	for _, p := range []string{"/var", "/var", "/bin", "/bin/", "/bin///"} {
		d.Append(p)
	}

	require.Equal(t, "/var:/bin", d.String())
	d.Prepend("/bin///")
	require.Equal(t, "/var:/bin", d.String())

	//require.Equal(t, 2, d.Append("/var"), ("/usr/local/bin", "/opt/local/bin"))
	//require.Equal(t, "/var:/bin:/usr/local/bin:/opt/local/bin", d.String())
}

func Test_DirList_Prepend(t *testing.T) {
	d := New()
	dirs := []string{
		"/var", "/var", "/bin", "/bin/",
	}

	for _, dir := range dirs {
		d.Prepend(dir)
	}

	d.Append("/bin/")

	require.Equal(t, "/bin:/var", d.String())
	d.Prepend("/sbin")
	require.Equal(t, d.Slice(), []string{"/sbin", "/bin", "/var"})
	require.Equal(t, d.String(), "/sbin:/bin:/var")
	d.Prepend("/var")
	d.Prepend("/usr/local/bin")
	d.Prepend("/opt/local/bin")
	require.Equal(t, d.String(), "/opt/local/bin:/usr/local/bin:/sbin:/bin:/var")
}

func Test_DirList_Drop(t *testing.T) {
	d := New()
	d.Load("/opt/local/bin:/usr/local/bin:/sbin:/bin:/var")
	d.Drop("/opt/local/bin")
	d.Drop("/opt/local/bin")
	d.Drop("/opt/local/bin")
	d.Drop("/usr/local/bin")
	d.Drop("/var")
	require.False(t, d.Nil())
	d.Drop("/sbin")
	d.Drop("/bin")
	require.Equal(t, "", d.String())
	require.True(t, d.Nil())
}
