package issue

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v48/github"
	"github.com/tsuzu/to-read-list/pkg/summarizer"
)

func Create(ctx context.Context, gh *github.Client, owner, repo string, meta *summarizer.Metadata) (string, error) {
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

	issue, _, err := gh.Issues.Create(ctx, owner, repo, &github.IssueRequest{
		Title:  &meta.Title,
		Body:   &body,
		Labels: &labels,
	})

	if err != nil {
		return "", err
	}

	return issue.GetHTMLURL(), nil
}
