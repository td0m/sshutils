package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/olekukonko/tablewriter"
	"github.com/shirou/gopsutil/host"
	"github.com/td0m/sshspy/pkg/util"
)

// TODO: check how to complete 2nd arg for "attach"
func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "help", Description: ""},
		{Text: "ls", Description: "List sessions"},
		{Text: "attach", Description: "Attach a session"},
		// TODO: Implement, notify on both CONNECT and DISCONNECT
		{Text: "watch", Description: "Start goroutine that looks for new ssh sessions to be attached and notifies you in real time."},
		{Text: "exit", Description: "Exit interactive prompt"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func runCommand(cmd string, args []string) {
	switch cmd {
	case "help":
	case "ls":
		printUserTable()
	case "attach":
		attach(args)
	}
}

// TODO: implement for numerical input
// TODO: check if device exists
func getTTYN(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err == nil {
		return n, nil
	}
	return -1, errors.New("Failed to parse TTY")
}

// uses `ps -ef` and regex to extract the PID of pts/X process
func getTTYPID(ttyN int) (int, error) {
	r, _ := regexp.Compile("^[^ ]+ +([0-9]+)")
	ttyStr := fmt.Sprintf("pts/%d", ttyN)

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

func attach(args []string) {
	if !util.IsAdmin() {
		fmt.Println("root permissions required! Please run this script as sudo")
		return
	}
	ttyN, err := getTTYN(args[0])
	if err != nil {
		log.Panic(err)
		return
	}
	fmt.Printf("Attempting to connect to: /dev/pts/%d\n", ttyN)
	pid, err := getTTYPID(ttyN)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("PID: %d\n", pid)
	spyOnTTY(pid)
}

func spyOnTTY(pid int) {
	cmd := exec.Command("strace", "-s", "16384", "-p", strconv.Itoa(pid), "-e", "read,write")
	stdoutPipe, _ := cmd.StderrPipe()
	//cmd.Stderr = os.Stderr // print errors directly into stderr
	cmd.Start()
	scanner := bufio.NewScanner(stdoutPipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "read") || strings.HasPrefix(line, "write") {
			parsed, err := parseLine(line)
			if err != nil {
				panic(line)
			}
			fmt.Println(parsed)
		}
	}
	cmd.Wait()
	fmt.Println("closed")
}

type ReadWriteCmd struct {
	Type  string
	Fd    int
	Buf   string
	Count int
}

func parseLine(line string) (ReadWriteCmd, error) {
	r, _ := regexp.Compile(`(read|write)\(([0-9]+), "((?:[^"\\]|\\.)*)", ([0-9]+)\) += +([0-9]+)`)
	matches := r.FindStringSubmatch(line)
	if len(matches) != 6 {
		return ReadWriteCmd{}, errors.New("failed parsing read/write command")
	}

	fd, _ := strconv.Atoi(matches[2])
	count, _ := strconv.Atoi(matches[4])

	return ReadWriteCmd{
		Type:  matches[1],
		Fd:    fd,
		Buf:   matches[3],
		Count: count,
	}, nil
}

func printUserTable() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Terminal", "User", "Started"})
	table.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor}, tablewriter.Colors{}, tablewriter.Colors{})

	users, _ := host.Users()
	selfTTY, _ := util.TTY()
	for _, u := range users {
		// make sure it's a remote ssh terminal session AND isn't the current tty client
		if u.Terminal[:3] == "pts" && "/dev/"+u.Terminal != selfTTY {
			table.Append([]string{"/dev/" + u.Terminal, u.User, "todo time"})
		}
	}

	if table.NumLines() > 0 {
		table.Render()
	} else {
		fmt.Println("no TTYs processes found")
	}
}

func main() {
	for {

		cmd := prompt.Input("> ", completer)
		if cmd == "exit" || cmd == "" {
			os.Exit(0)
			break
		}
		words := strings.Split(cmd, " ")
		rest := []string{}
		if len(words) > 1 {
			rest = words[1:]
		}
		runCommand(words[0], rest)
	}
}
