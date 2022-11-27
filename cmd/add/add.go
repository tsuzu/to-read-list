package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/go-github/v48/github"
	"github.com/tsuzu/to-read-list/pkg/issue"
	"github.com/tsuzu/to-read-list/pkg/summarizer"
	"golang.org/x/oauth2"
)

func createIssue(meta *summarizer.Metadata) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	repo := strings.TrimPrefix(os.Getenv("GITHUB_REPOSITORY"), owner+"/")

	link, err := issue.Create(ctx, client, owner, repo, meta)

	if err != nil {
		panic(err)
	}

	fmt.Println(link)
}

func main() {
	meta, err := summarizer.GetMetadata(os.Args[1])

	if err != nil {
		panic(err)
	}

	fmt.Println(meta.Title)
	fmt.Println(meta.Type)
	fmt.Println(meta.SiteName)
	fmt.Println(meta.Outline)

	createIssue(meta)
}
