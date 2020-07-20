package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"

	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/olekukonko/tablewriter"
	"github.com/td0m/sshutils/pkg/pts"
)

var ptsSuggestions []prompt.Suggest

func main() {
	printUserTable()
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

func runCommand(cmd string, args []string) {
	switch cmd {
	case "help":
	case "kill":
		killPts(args)
	case "ls":
		printUserTable()
	case "attach":
		attach(args)
	}
}

func killPts(args []string) {
	if len(args) != 1 {
		fmt.Printf("1 argument required, given %d", len(args))
		return
	}
	pts, err := pts.New(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	pts.Kill()
}

func updateCompletions() {
	ptsSuggestions = []prompt.Suggest{}
	ptsList := pts.ListAll()
	for _, pts := range ptsList {
		ptsSuggestions = append(ptsSuggestions, prompt.Suggest{Text: pts.Terminal[4:]})
	}
}

// TODO: check how to complete 2nd arg for "attach"
func completer(d prompt.Document) []prompt.Suggest {
	commands := []prompt.Suggest{
		{Text: "help", Description: ""},
		{Text: "ls", Description: "List PTS sessions"},
		// TODO: implement, use kill -9 PID, make sure we have root access
		{Text: "kill", Description: "Kill a PTS"},
		{Text: "attach", Description: "Attach a PTS"},
		// TODO: Implement, notify on both CONNECT and DISCONNECT
		{Text: "watch", Description: "Start goroutine that looks for new ssh sessions to be attached and notifies you in real time."},
		{Text: "exit", Description: "Exit interactive prompt"},
	}
	words := strings.Split(d.TextBeforeCursor(), " ")
	command := words[0]
	if len(words) == 1 {
		return prompt.FilterHasPrefix(commands, d.GetWordBeforeCursor(), false)
	}
	switch command {
	case "kill":
		fallthrough
	case "attach":
		return prompt.FilterHasPrefix(ptsSuggestions, d.GetWordBeforeCursor(), false)
	}
	return []prompt.Suggest{}
}

// TODO: check if device exists
func getPTSN(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err == nil {
		return n, nil
	}
	return -1, errors.New("Failed to parse TTY")
}

func attach(args []string) {
	if len(args) != 1 {
		fmt.Printf("1 argument required, given %d", len(args))
		return
	}
	pts, err := pts.New(args[0])
	if err != nil {
		fmt.Println(err)
		return
	}

	done := make(chan bool, 1)
	go func() {
		log.Panic(pts.ReadBuffer(os.Stdout))
	}()

	go func() {
		exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
		// don't display entered characters on the screen
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

		var b []byte = make([]byte, 1)
		for {
			os.Stdin.Read(b)
			pts.Write(b)
		}
	}()

	<-done
}

func printUserTable() {
	updateCompletions()

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Terminal", "User", "Started"})
	table.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor}, tablewriter.Colors{}, tablewriter.Colors{})

	pts := pts.ListAll()
	for _, u := range pts {
		table.Append([]string{"/dev/" + u.Terminal, u.User, "todo time"})
	}

	if table.NumLines() > 0 {
		table.Render()
	} else {
		fmt.Println("no TTYs processes found")
	}
}
