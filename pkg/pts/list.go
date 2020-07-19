package pts

import (
	"os"

	"github.com/shirou/gopsutil/host"
)

func ListAll() []host.UserStat {
	users, _ := host.Users()
	thisTerminal, _ := currentTerminalDevice()

	filtered := []host.UserStat{}
	for _, u := range users {
		// make sure it's a remote ssh terminal session AND isn't the current pts client
		if u.Terminal[:3] == "pts" && "/dev/"+u.Terminal != thisTerminal {
			filtered = append(filtered, u)
		}
	}

	return filtered
}

const fd0 = "/proc/self/fd/0"

// TTY prints the file name of the terminal connected to standard input
func currentTerminalDevice() (string, error) {
	dest, err := os.Readlink(fd0)
	if err != nil {
		return "", err
	}
	return dest, nil
}
