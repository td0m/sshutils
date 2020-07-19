package pts

import "os"

type ReadCmd struct {
	Fd    int
	Buf   string
	Count int
	Out   int
}

type PTS struct {
	ID  string
	PID int
}

type PTSInterface interface {
	Path() string
	Kill() error
	Write([]byte) error
	ReadBuffer(*os.File) error
}
