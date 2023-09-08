package path

import (
	"testing"
)

func TestAddDir(t *testing.T) {
	type args struct {
		path   string
		s      string
		append bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", "", false}, "."},
		{"push ok", args{"hello:world:x:/y", "/Y", false}, "/Y:hello:world:x:/y"},
		{"push no op", args{"hello:world:x:/y", "/y/", false}, "hello:world:x:/y"},
		{"push to empty", args{"", "/y///", false}, "/y"},
		{"empty (append mode)", args{"", "", true}, "."},
		{"push ok (append mode)", args{"hello:world:x:/y", "/Y", true}, "hello:world:x:/y:/Y"},
		{"push no op (append mode)", args{"hello:world:x:/y", "/y/", true}, "hello:world:x:/y"},
		{"push to empty (append mode)", args{"", "/y///", true}, "/y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddDir(tt.args.path, tt.args.s, tt.args.append); got != tt.want {
				t.Errorf("AddDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveDir(t *testing.T) {
	type args struct {
		path string
		s    string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty", args{"", ""}, ""},
		{"not found", args{"hello:world:x:/y", "/Y///"}, "hello:world:x:/y"},
		{"sanitized and removed", args{"hello:world:x:/y", "/y/"}, "hello:world:x"},
		{"many occurrencies", args{"world:hello:world:x:/y", "world"}, "hello:x:/y"},
		{"remove all", args{"world:", "world"}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RemoveDir(tt.args.path, tt.args.s); got != tt.want {
				t.Errorf("RemoveDir() = %v, want %v", got, tt.want)
			}
		})
	}
}
