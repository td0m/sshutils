package pts

type ReadCmd struct {
	Fd    int
	Buf   string
	Count int
	Out   int
}
