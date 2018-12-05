package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	"github.com/labstack/echo"
)

type Search struct {
	//XMLName xml.Name `xml:"search"`
	Entry []Entry `xml:"entry" json:"entry"`
}

type Entry struct {
	Title      string `xml:"title" json:"title"`
	Url        string `xml:"url" json:"url"`
	Content    string `xml:"content"`
	Categories []struct {
		Category string `xml:"category"`
	} `xml:"categories"`
	Tag []struct {
		Tag string `xml:"tag"`
	} `xml:"tags"`
}

type Article struct {
	Title      string   `json:"title"`
	URL        string   `json:"url"`
	Content    string   `json:"content"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags"`
}

var Articles []Article

func main() {
	initArticles()

	e := echo.New()

	e.Static("/static", "static")
	e.GET("/search", search)

	e.Start(":10001")
}

func initArticles() {
	searchXML := "./assets/search.xml"
	entries := parseSearchXML(searchXML)

	for _, e := range entries {
		path, _ := url.QueryUnescape(e.Url)

		article := Article{
			URL:     fmt.Sprintf("https://www.uegeek.com%s", path),
			Title:   e.Title,
			Content: e.Content,
		}

		for _, c := range e.Categories {
			article.Categories = append(article.Categories, c.Category)
		}

		for _, c := range e.Tag {
			article.Tags = append(article.Tags, c.Tag)
		}

		Articles = append(Articles, article)
	}
}

func search(ctx echo.Context) error {
	keywords := ctx.QueryParam("keywords")
	pageSizeQuery := ctx.QueryParam("pageSize")
	pageIndexQuery := ctx.QueryParam("pageIndex")

	pageSize := 10
	pageIndex := 1
	if pageSizeQuery != "" {
		pageSize, _ = strconv.Atoi(pageSizeQuery)
	}
	if pageIndexQuery != "" {
		pageIndex, _ = strconv.Atoi(pageIndexQuery)
	}

	var articles []Article
	for _, a := range Articles {
		if keywords == "" || strings.Contains(a.Title, keywords) {
			articles = append(articles, a)
		}
	}

	total := len(articles)
	start := (pageIndex - 1) * pageSize
	end := pageIndex * pageSize

	if start > total {
		articles = nil
	} else if end > total {
		articles = articles[start:total]
	} else {
		articles = articles[start:end]
	}

	return ctx.JSON(200, struct {
		Articles []Article `json:"articles"`
		Total    int       `json:"total"`
	}{Articles: articles, Total: total})
}

func parseSearchXML(path string) []Entry {
	//buf, err := ioutil.ReadFile(path)
	//panicIfError(err)
	//

	// 处理 xml 内容中的不可见字符，防止无法解析 xml
	printOnly := func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}

	URL := "https://www.uegeek.com/search.xml"
	resp, err := http.Get(URL)
	panicIfError(err)
	buf, err := ioutil.ReadAll(resp.Body)

	xmlData := []byte(strings.Map(printOnly, string(buf)))
	//content := strings.Replace(string(buf), "\b", "", -1)
	var search Search
	err = xml.Unmarshal(xmlData, &search)
	panicIfError(err)

	return search.Entry
}

func panicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
