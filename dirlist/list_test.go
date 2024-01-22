package dirlist_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"al.essio.dev/pkg/tools/dirlist"
)

func TestList_Append(t *testing.T) {
	d := dirlist.New()
	require.Equal(t, "", d.String())
	for _, p := range []string{"/var", "/var", "/bin", "/bin/", "/bin///"} {
		d.Append(p)
	}

	require.Equal(t, "/var:/bin", d.String())
	d.Prepend("/bin///")
	require.Equal(t, "/var:/bin", d.String())
}

func TestList_Prepend(t *testing.T) {
	d := dirlist.New()
	dirs := []string{
		"/var", "/var", "/bin", "/bin/",
	}

	for _, dir := range dirs {
		d.Prepend(dir)
	}

	d.Prepend("/sbin")
	d.Prepend("/var")
	d.Prepend("/usr/local/bin")
	d.Prepend("/opt/local/bin")
	require.Equal(t, "/opt/local/bin:/usr/local/bin:/sbin:/bin:/var", d.String())
}

func TestList_Drop(t *testing.T) {
	d := dirlist.New()
	d.Load("/opt/local/bin:/usr/local/bin:/sbin:/bin:/var:/bin")
	d.Drop("/opt/local/bin")
	d.Drop("/opt/local/bin")
	d.Drop("/opt/local/bin")
	d.Drop("/usr/local/bin")
	d.Drop("/var")
	require.NotEqual(t, "", d.String())
	d.Drop("/sbin")
	d.Drop("/bin")
	require.Equal(t, "", d.String())

	require.NotPanics(t, func() { dirlist.New().Drop("") })

	d1 := dirlist.New()
	d1.Load(`/Library/Application Support:/Library/Application Support/`)
	d1.Drop("/Library/Application Support")
	require.False(t, d1.Contains("/Library/Application Support"))
}

func TestList_Reset(t *testing.T) {
	d1 := dirlist.New()
	d1.Reset()

	d2 := dirlist.New()
	d2.Load("/opt/local/bin:/usr/local/bin:/sbin:/bin:/var:/bin")
	d2.Reset()
	require.Equal(t, 0, len(d2.Slice()))
	require.Equal(t, "", d2.String())
}

func TestList_Contains(t *testing.T) {
	d := dirlist.New()
	d.Load("/opt/local/bin:/usr/local/bin:/sbin:/bin:/var:/bin")
	require.False(t, d.Contains("/ur/local/sbin"))
	require.False(t, d.Contains("/ur/local////sbin/"))
	require.True(t, d.Contains("/sbin"))
	require.True(t, d.Contains("///sbin//"))
}

func TestList_LoadEnv(t *testing.T) {
	tests := []struct {
		name string
		val  string
		want string
	}{
		{"empty", "", ""},
	}
	for i, tt := range tests {
		tt2 := tt
		t.Run(tt2.name, func(t *testing.T) {
			envvar := fmt.Sprintf("%s_%d_VAR", t.Name(), i)
			d := dirlist.New()
			t.Setenv(envvar, tt.val)
			d.LoadEnv(envvar)
			require.Equal(t, tt2.want, d.String())
		})
	}
}

func TestList_Slice(t *testing.T) {
	d := dirlist.New()
	require.Equal(t, 0, len(d.Slice()))
	d.Load("/usr/bin:/bin")
	require.Equal(t, []string{"/usr/bin", "/bin"}, d.Slice())
}

func TestList_String(t *testing.T) {
	d := dirlist.New()
	d.Load("/usr/bin:/bin")
	require.Equal(t, "/usr/bin:/bin", d.String())
}
