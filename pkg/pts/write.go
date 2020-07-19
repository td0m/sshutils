package pts

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"
)

func Write(ptsN int, b []byte) {
	tty := fmt.Sprintf("/dev/pts/%d", ptsN)
	ttyFile, err := os.Open(tty)
	if err != nil {
		log.Fatalln(err)
	}
	defer ttyFile.Close()

	var eno syscall.Errno
	for _, c := range b {
		_, _, eno = syscall.Syscall(syscall.SYS_IOCTL,
			ttyFile.Fd(),
			syscall.TIOCSTI,
			uintptr(unsafe.Pointer(&c)),
		)
		if eno != 0 {
			log.Fatalln(eno)
		}
	}
}
