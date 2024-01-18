package dirlist

var (
	dList List
)

func init() {
	dList = New()
}

func Reset() { dList.Reset() }

func Contains(p string) bool {
	return dList.Contains(quoteAndClean(p))
}

func Load(s string) {
	dList.Load(s)
}

func LoadEnv(s string) {
	dList.LoadEnv(s)
}

func Prepend(p string) {
	dList.Prepend(p)
}

func Append(p string) {
	dList.Append(p)
}

func Drop(p string) {
	dList.Drop(p)
}

func Slice() []string {
	return dList.Slice()
}

func String() string {
	return dList.String()
}
