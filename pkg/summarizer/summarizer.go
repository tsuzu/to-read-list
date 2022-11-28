package summarizer

import (
	"net/http"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

type Metadata struct {
	URL      string
	Title    string
	Type     string
	Image    string
	SiteName string
	Outline  string
}

func GetMetadata(url string) (*Metadata, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)

	if err != nil {
		return nil, err
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

	return &mf.meta, nil
}

type MetaFinder struct {
	doc *goquery.Document

	meta Metadata
}

func (mf *MetaFinder) findTitle() {
	ogTitle := mf.doc.Find(`meta[property="og:title"]`)

	val, exists := ogTitle.Attr("content")

	if exists {
		mf.meta.Title = spaces.ReplaceAllString(val, " ")

		return
	}

	titleNode := mf.doc.Find("head > title").First()

	if titleNode == nil {
		return
	}

	mf.meta.Title = spaces.ReplaceAllString(titleNode.Text(), " ")
}

func (mf *MetaFinder) findSiteName() {
	ogSiteName := mf.doc.Find(`meta[property="og:site_name"]`)

	val, exists := ogSiteName.Attr("content")

	if !exists {
		return
	}

	mf.meta.SiteName = val
}

func (mf *MetaFinder) findImage() {
	ogImage := mf.doc.Find(`meta[property="og:image"]`)

	val, exists := ogImage.Attr("content")

	if !exists {
		return
	}

	mf.meta.Image = val
}

func (mf *MetaFinder) findType() {
	ogType := mf.doc.Find(`meta[property="og:type"]`)

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
