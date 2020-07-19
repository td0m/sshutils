package main

import (
	"fmt"
	"os"

	"github.com/c-bata/go-prompt"
	"github.com/olekukonko/tablewriter"
	"github.com/shirou/gopsutil/host"
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

func isAdmin() bool {
	return os.Geteuid() == 0
}

func printUserTable() {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Terminal", "User", "Started"})
	table.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgCyanColor}, tablewriter.Colors{}, tablewriter.Colors{})

	users, _ := host.Users()
	for _, u := range users {
		if u.Terminal[:3] == "pts" {
			table.Append([]string{u.Terminal, u.User, "todo time"})
		}
	}

	table.Render()
}

func main() {
	u, _ := host.Users()
	fmt.Printf("%+v\n", u)
	for {
		cmd := prompt.Input("> ", completer)
		if cmd == "exit" {
			break
		}
		runCommand(cmd)
	}
}
