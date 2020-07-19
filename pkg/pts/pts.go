package pts

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// create a new pts
func New(s string) (PTSInterface, error) {
	ptsList := ListAll()
	found := false

	for _, pts := range ptsList {
		if "/dev/"+pts.Terminal == "/dev/pts/"+s {
			found = true
			break
		}
	}
	if !found {
		return nil, errors.New(fmt.Sprintf(`device "/dev/pts/%s" not found`, s))
	}

	pid, err := getPID(s)
	if err != nil {
		return nil, err
	}

	return &PTS{s, pid}, nil
}

// returns the path of the device
func (pts *PTS) Path() string {
	return "/dev/pts/" + pts.ID
}

// kills the current pts
func (pts *PTS) Kill() error {
	return exec.Command("kill", "-9", strconv.Itoa(pts.PID)).Run()
}

// write bytes to pts
func (pts *PTS) Write(bs []byte) error {
	ttyFile, err := os.Open(pts.Path())
	if err != nil {
		return err
	}
	defer ttyFile.Close()

	var eno syscall.Errno
	for _, b := range bs {
		_, _, eno = syscall.Syscall(syscall.SYS_IOCTL,
			ttyFile.Fd(),
			syscall.TIOCSTI,
			uintptr(unsafe.Pointer(&b)),
		)
		if eno != 0 {
			return errors.New(eno.Error())
		}
	}
	return nil
}

// read pts stdout/stderr and write it to file f
func (pts *PTS) ReadBuffer(f *os.File) error {
	// -xx specifies hex escapes, these can be decoded with Go's "strconv.Unquote"
	cmd := exec.Command("strace", "-xx", "-s", "16384", "-p", strconv.Itoa(pts.PID), "-e", "read")
	stdoutPipe, _ := cmd.StdoutPipe()
	// redirect stderr to stdout (2>&1) as we need to read both
	cmd.Stderr = cmd.Stdout
	cmd.Start()
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		// only try to parse read and write syscalls
		if strings.HasPrefix(line, "read") || strings.HasPrefix(line, "write") {
			c, err := parseLine(line)
			if err != nil {
				return errors.New(line)
			}
			// not sure why, but after trial and error this worked
			if c.Count == 16384 && c.Fd != 4 {
				f.WriteString(c.Buf)
			}
		}
	}
	cmd.Wait()
	return nil
}
