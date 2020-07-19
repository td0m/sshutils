package main

import (
	"fmt"
	"os"

	"github.com/c-bata/go-prompt"
	"github.com/olekukonko/tablewriter"
	"github.com/shirou/gopsutil/host"
	"github.com/td0m/sshspy/pkg/util"
)

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "help", Description: ""},
		{Text: "ls", Description: "List sessions"},
		{Text: "exit", Description: "Exit interactive prompt"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func runCommand(cmd string) {
	switch cmd {
	case "help":
	case "ls":
		printUserTable()
	}
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
			table.Append([]string{u.Terminal, u.User, "todo time"})
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
		runCommand(cmd)
	}
}
