package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-github/v48/github"
	"golang.org/x/oauth2"
)

type Metadata struct {
	URL      string
	Title    string
	Type     string
	Image    string
	SiteName string
	Outline  string
}

type MetaFinder struct {
	doc *goquery.Document

	meta Metadata
}

func (mf *MetaFinder) findTitle() {
	ogTitle := mf.doc.Find(`meta[property="og:title"]`)

	if ogTitle != nil && ogTitle.Text() != "" {
		val, exists := ogTitle.Attr("content")

		if exists {
			mf.meta.Title = val

			return
		}
	}

	titleNode := mf.doc.Find("title")

	if titleNode == nil {
		return
	}

	mf.meta.Title = titleNode.Text()
}

func (mf *MetaFinder) findSiteName() {
	ogSiteName := mf.doc.Find(`meta[property="og:site_name"]`)

	if ogSiteName == nil {
		return
	}

	val, exists := ogSiteName.Attr("content")

	if !exists {
		return
	}

	mf.meta.SiteName = val
}

func (mf *MetaFinder) findImage() {
	ogImage := mf.doc.Find(`meta[property="og:image"]`)

	if ogImage == nil {
		return
	}

	val, exists := ogImage.Attr("content")

	if !exists {
		return
	}

	mf.meta.Image = val
}

func (mf *MetaFinder) findType() {
	ogType := mf.doc.Find(`meta[property="og:type"]`)

	if ogType == nil {
		return
	}

	val, exists := ogType.Attr("content")

	if !exists {
		return
	}

	mf.meta.Type = val
}

var spaces = regexp.MustCompile(`\s{2,}`)

func (mf *MetaFinder) findOutline() {
	deletedTags := []string{"script", "noscript", "style", "header", "nav", "footer"}
	for _, d := range deletedTags {
		mf.doc.Find(d).Each(func(i int, s *goquery.Selection) {
			s.Remove()
		})
	}

	if article := mf.doc.Find("article"); article != nil && article.Text() != "" {
		mf.meta.Outline = spaces.ReplaceAllString(article.Text(), "\n")

		return
	}

	if body := mf.doc.Find("body"); body != nil && body.Text() != "" {
		mf.meta.Outline = spaces.ReplaceAllString(body.Text(), "\n")

		return
	}

	mf.meta.Outline = spaces.ReplaceAllString(mf.doc.Text(), "\n")
}

func createIssue(meta Metadata) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	owner := os.Getenv("GITHUB_REPOSITORY_OWNER")
	repo := strings.TrimPrefix(os.Getenv("GITHUB_REPOSITORY"), owner+"/")

	outline := []rune(strings.ReplaceAll(meta.Outline, "```", ""))
	if len(outline) > 60000 {
		outline = outline[:60000]
	}
	outlineBlock := "```\n" + string(outline) + "\n```"

	body := fmt.Sprintf(`Title: %s
URL: %s

![OG Image](%s)

<details>

%s

</details>
`, meta.Title, meta.URL, meta.Image, outlineBlock)

	labels := []string{}
	if meta.Type != "" {
		labels = append(labels, fmt.Sprintf("type:%s", meta.Type))
	}
	if meta.SiteName != "" {
		labels = append(labels, fmt.Sprintf("site:%s", meta.SiteName))
	}

	_, _, err := client.Issues.Create(ctx, owner, repo, &github.IssueRequest{
		Title:  &meta.Title,
		Body:   &body,
		Labels: &labels,
	})

	if err != nil {
		panic(err)
	}
}

func main() {
	url := os.Args[1]

	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		panic(err)
	}

	mf := MetaFinder{
		doc: doc,
		meta: Metadata{
			URL: url,
		},
	}

	mf.findTitle()
	mf.findSiteName()
	mf.findImage()
	mf.findType()
	mf.findOutline()

	fmt.Println(mf.meta.Title)
	fmt.Println(mf.meta.Type)
	fmt.Println(mf.meta.SiteName)
	fmt.Println(mf.meta.Outline)

	createIssue(mf.meta)
}
