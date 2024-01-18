package dirlist

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppend(t *testing.T) {
	Reset()
	require.Equal(t, "", String())
	for _, p := range []string{"/var", "/var", "/bin", "/bin/", "/bin///"} {
		Append(p)
	}

	require.Equal(t, "/var:/bin", String())
	Prepend("/bin///")
	require.Equal(t, "/var:/bin", String())
}

func TestContains(t *testing.T) {
	Reset()
	Load("/opt/local/bin:/usr/local/bin:/sbin:/bin:/var:/bin")
	require.False(t, Contains("/ur/local/sbin"))
	require.False(t, Contains("/ur/local////sbin/"))
	require.True(t, Contains("/sbin"))
	require.True(t, Contains("///sbin//"))

}

func TestDrop(t *testing.T) {
	Reset()
	Load("/opt/local/bin:/usr/local/bin:/sbin:/bin:/var:/bin")
	require.Equal(t, Slice(), []string{"/opt/local/bin", "/usr/local/bin", "/sbin", "/bin", "/var"})
	Drop("/opt/local/bin")
	Drop("/opt/local/bin")
	Drop("/opt/local/bin")
	Drop("/usr/local/bin")
	Drop("/var")
	require.NotEqual(t, "", String())
	Drop("/sbin")
	Drop("/bin")
	require.False(t, Contains("/bin"))
	require.Equal(t, "", String())

	Reset()
	require.NotPanics(t, func() { Drop("") })
}

func TestLoadEnv(t *testing.T) {
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
			Reset()
			t.Setenv(envvar, tt.val)
			LoadEnv(envvar)
			require.Equal(t, tt2.want, String())
		})
	}
}

func TestPrepend(t *testing.T) {
	Reset()
	require.Equal(t, "", String())
	for _, p := range []string{"/var", "/var", "/bin", "/bin/", "/bin///"} {
		Append(p)
	}

	require.Equal(t, "/var:/bin", String())
	Prepend("/bin///")
	require.Equal(t, "/var:/bin", String())
}

func TestSlice(t *testing.T) {
	Reset()
	require.Equal(t, 0, len(Slice()))
	Prepend("/usr/bin")
	Append("/bin")
	require.Equal(t, []string{"/usr/bin", "/bin"}, Slice())
}

func TestString(t *testing.T) {
	Reset()
	require.Equal(t, "", String())
	Prepend("/usr/bin")
	Append("/bin")
	require.Equal(t, "/usr/bin:/bin", String())
}
