package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"github.com/urfave/cli"
)

var ListCommand = cli.Command{
	Name:   "list",
	Usage:  "list pull requests",
	Action: listAction,
	Flags: []cli.Flag{
		RepoFlag,
		cli.StringFlag{
			Name:  "state",
			Value: "open",
			Usage: "filter by pull status. [open|closed|all]",
		},
		cli.StringFlag{
			Name:  "sort",
			Value: "created",
			Usage: "what to sort results by. [created|updated|popularity|long-running]",
		},
		cli.IntFlag{
			Name:  "per-page",
			Value: 30,
			Usage: "Amount of results to return per page. [0-100]",
		},
	},
}

func listAction(c *cli.Context) error {
	ctx := context.Background()
	boldString := color.New(color.Bold).SprintFunc()

	userStr, repoStr := ParseRepoFlag(c)
	color.Green("Listing pull requests for https://github.com/%s/%s/", userStr, repoStr)

	// create the github client.
	client := github.NewClient(nil)

	// for loop enables pagination
	for page := 1; ; page++ {
		// request the pull requests for this page
		pulls, resp, err := client.PullRequests.List(
			ctx, userStr, repoStr,
			&github.PullRequestListOptions{
				State:       c.String("state"),
				Sort:        c.String("sort"),
				ListOptions: github.ListOptions{Page: page, PerPage: c.Int("per-page")},
			},
		)

		if err != nil {
			// if we hit the rate limit, display how long till it resets and
			// provide info on authenticating.
			// TODO: support authentication for the list command
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

		// range over the retrieved pull requests and render them
		for _, pull := range pulls {
			user := pull.GetUser()
			fmt.Printf("#%d %s by %s Last Updated %s\n", pull.GetNumber(), boldString(pull.GetTitle()), user.GetLogin(), humanize.Time(pull.GetUpdatedAt()))
		}

		// if we're out of pull requests to retrieve, quit.
		if len(pulls) == 0 || len(pulls) != c.Int("per-page") {
			// and if we never retrieved any, supply some feedback
			if len(pulls) == 0 && page == 1 {
				fmt.Println("No results.")
			}
			break
		}

		// otherwise we're paginating. Give the user a chance to ctrl-c first though
		fmt.Printf("Page %d of %d...", page, resp.LastPage)
		bufio.NewReader(os.Stdin).ReadBytes('\n')
	}
	return nil
}
