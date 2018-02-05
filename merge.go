package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
)

var MergeCommand = cli.Command{
	Name:   "merge",
	Usage:  "merge pull requests eg: `gib merge 21`",
	Action: mergeAction,
	Flags: []cli.Flag{
		RepoFlag,
	},
}

func mergeAction(c *cli.Context) error {
	ctx := context.Background()
	reader := bufio.NewReader(os.Stdin)
	boldString := color.New(color.Bold).SprintFunc()
	userStr, repoStr := ParseRepoFlag(c)
	color.Green("Merging pull request into https://github.com/%s/%s/", userStr, repoStr)

	tp := CreateBasicAuthTransport()
	client := github.NewClient(tp.Client())

	user, _, err := client.Users.Get(ctx, "")

	// Is this a two-factor auth error? If so, prompt for OTP and try again.
	if _, ok := err.(*github.TwoFactorAuthError); ok {
		fmt.Print("\nGitHub OTP: ")
		otp, _ := reader.ReadString('\n')
		tp.OTP = strings.TrimSpace(otp)
		user, _, err = client.Users.Get(ctx, "")
	}

	if err != nil {
		color.Red("Error:", err.Error())
		return err
	}

	pullNumberStr := c.Args().Get(0)
	if pullNumberStr == "" {
		fmt.Print("Enter the pull request number to merge: ")
		var err error
		pullNumberStr, err = reader.ReadString('\n')
		if err != nil {
			color.Red("Error:", err.Error())
			return err
		}
	}

	pullNumber, err := strconv.Atoi(pullNumberStr)
	if err != nil {
		color.Red("Error:", err.Error())
		return err
	}

	pull, resp, err := client.PullRequests.Get(ctx, userStr, repoStr, pullNumber)
	fmt.Printf("rate-limit:%d, remaining:%d, rate-limit resets %s\n", resp.Limit, resp.Remaining, humanize.Time(resp.Reset.Time))
	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			color.Red("Error: Rate limit exceeded\n")
			fmt.Printf("rate-limit:%d, remaining:%d, rate-limit resets %s\n", resp.Limit, resp.Remaining, humanize.Time(resp.Reset.Time))
			fmt.Println("Unauthenticated users are limited to 60 requests per hour")
			fmt.Println("Authenticated users get 5000 requests an hour")
			cli.OsExiter(1)
		}
		color.Red("Error: unhandled api failure")
		color.Red(err.Error())
		return err
	}

	user = pull.GetUser()
	for confirmed := false; confirmed != true; {
		fmt.Printf("Merge #%d %s by %s? [y/n] ", pull.GetNumber(), boldString(pull.GetTitle()), user.GetLogin())
		input, _ := reader.ReadString('\n')

		switch input[0] {
		case 'n':
			fallthrough
		case 'N':
			fmt.Println("Aborting merge.")
			cli.OsExiter(0)
		case 'y':
			fallthrough
		case 'Y':
			confirmed = true
		default:
		}
	}

	fmt.Println("Please enter a commit message for the merge (or leave blank):")
	var msg string
	prevEmpty := true
	for {
		fmt.Print(">")
		in, _ := reader.ReadString('\n')
		msg += in
		if in == "\n" {
			if prevEmpty {
				break
			}
			prevEmpty = true
		} else {
			prevEmpty = false
		}
	}

	// trim extra whitespace and newline, then add a single newline.
	msg = strings.Trim(msg, " \n")
	if len(msg) > 0 {
		msg += "\n"
	}

	res, resp, err := client.PullRequests.Merge(
		ctx, userStr, repoStr, pullNumber, msg,
		&github.PullRequestOptions{MergeMethod: c.String("method")},
	)

	if err != nil {
		if _, ok := err.(*github.RateLimitError); ok {
			color.Red("Error: Rate limit exceeded\n")
			fmt.Printf("rate-limit:%d, remaining:%d, rate-limit resets %s\n", resp.Limit, resp.Remaining, humanize.Time(resp.Reset.Time))
			fmt.Println("Unauthenticated users are limited to 60 requests per hour")
			fmt.Println("Authenticated users get 5000 requests an hour")
			cli.OsExiter(1)
		}
		color.Red(err.Error())
		return err
	}

	fmt.Println(res.GetMessage())

	return nil
}
