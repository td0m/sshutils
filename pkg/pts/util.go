package pts

import (
	"bufio"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"errors"
)

// uses `ps -ef` and regex to extract the PID of /dev/pts/X process
func getPID(pts string) (int, error) {
	r, _ := regexp.Compile("^[^ ]+ +([0-9]+)")
	ttyStr := "pts/" + pts

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
