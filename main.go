package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	clilog "github.com/b4b4r07/go-cli-log"
	"github.com/google/go-github/github"
	"github.com/jessevdk/go-flags"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
)

// These variables are set in build step
var (
	Version  = "unset"
	Revision = "unset"
)

// CLI represents this application itself
type CLI struct {
	Option Option
	Stdout io.Writer
	Stderr io.Writer
}

// Option represents application options
type Option struct {
	Owner  string `long:"owner" description:"owner" required:"true"`
	Repo   string `long:"repo" description:"repo" required:"true"`
	Number int    `long:"number" description:"number" required:"true"`
	Event  string `long:"event" description:"review event" choice:"APPROVE" choice:"REQUEST_CHANGES" choice:"COMMENT" default:"APPROVE"`
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	clilog.Env = "LOG"
	clilog.SetOutput()
	defer log.Printf("[INFO] finish main function")

	log.Printf("[INFO] Version: %s (%s)", Version, Revision)
	log.Printf("[INFO] Args: %#v", args)

	var opt Option
	args, err := flags.ParseArgs(&opt, args)
	if err != nil {
		return 2
	}

	cli := CLI{
		Option: opt,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	if err := cli.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return 1
	}

	return 0
}

func (c *CLI) Run(args []string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if len(token) == 0 {
		return errors.New("GITHUB_TOKEN is missing")
	}

	// Construct github HTTP client
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: token,
	})
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	tc.Transport = clilog.NewTransport("github", tc.Transport)
	client := github.NewClient(tc)

	review, _, err := client.PullRequests.CreateReview(
		context.Background(),
		c.Option.Owner,
		c.Option.Repo,
		c.Option.Number,
		// https://docs.github.com/en/rest/pulls/reviews#create-a-review-for-a-pull-request
		&github.PullRequestReviewRequest{
			Event: github.String(c.Option.Event),
		},
	)
	if err != nil {
		log.Printf("[ERROR] Failed to create a review")
		return err
	}

	log.Printf("[INFO] Successfully created a review!")
	log.Printf("[TRACE] response: %v", review)

	return nil
}
