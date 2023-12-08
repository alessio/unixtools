package path

// List handles a list of directories in a predictable way.
type List interface {
	String() interface{} //
	StringSlice() []string
	Prepend(path string) bool
	Append(path string) bool
	Drop(path string) bool
}
