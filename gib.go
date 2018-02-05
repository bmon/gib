package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"

	"github.com/fatih/color"
	"github.com/google/go-github/github"

	"github.com/urfave/cli"
)

var RedString = color.New(color.FgRed).SprintFunc()
var RepoFlag = cli.StringFlag{
	Name:   "repo, r",
	Usage:  RedString("[REQUIRED]") + " Repository to operate on",
	EnvVar: "GIB_REPO",
}

func main() {
	app := cli.NewApp()

	// create a copy of a repo flag, we're going to reuse it a bit
	// list all available commands
	app.Commands = []cli.Command{
		ListCommand,
		MergeCommand,
	}
	app.Flags = []cli.Flag{
		RepoFlag,
	}
	app.Run(os.Args)
}

// utility function to ensure the --repo flag has been supplied
func ParseRepoFlag(c *cli.Context) (string, string) {
	repo := c.String("repo")
	if repo == "" {
		color.Red("Error: The --repo flag is required\n")
		cli.ShowAppHelpAndExit(c, 1)
	}
	// shave off "..github.com." if it has been supplied
	urlSplit := strings.Split(repo, "github.com/")
	if len(urlSplit) == 2 {
		repo = urlSplit[1]
	}
	repoSplit := strings.Split(repo, "/")
	if len(repoSplit) != 2 {
		color.Red("Error: Bad repository supplied. Should be of format `user/repo`\n")
		cli.OsExiter(1)
	}
	return repoSplit[0], repoSplit[1]
}

func CreateBasicAuthTransport() github.BasicAuthTransport {
	r := bufio.NewReader(os.Stdin)
	fmt.Print("GitHub Username: ")
	username, _ := r.ReadString('\n')

	fmt.Print("GitHub Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	// because ReadPassword does not blit the newline char
	fmt.Println()
	password := string(bytePassword)

	return github.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}
}
