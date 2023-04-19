package crawler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const ext = ".html"

var (
	_ Crawler = (*CrawlerService)(nil)

	httpClient = &http.Client{
		Timeout: 5 * time.Second,
	}
)

// CrawlerService service implementation of the crawler interface
type CrawlerService struct {
	basePath, destPath string
	visited            map[string]struct{}
	linksChan          chan string
}

// New creates  anew instance/implementation of the crawler service
func New(path, destPath string) Crawler {
	return &CrawlerService{
		basePath:  path,
		destPath:  destPath,
		visited:   make(map[string]struct{}),
		linksChan: make(chan string),
	}
}

// Crawl reads data from the external path
func (cs *CrawlerService) Crawl(ctx context.Context, link string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, link, nil)
	if err != nil {
		return nil, err
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	cs.visited[link] = struct{}{}
	return res, nil
}

// BaseURL returns the original basepath
func (cs *CrawlerService) BaseURL() string {
	return cs.basePath
}

// ExtractLinks extracts links from given web body
func (cs *CrawlerService) ExtractLinks(ctx context.Context, r io.Reader) ([]string, error) {
	links := []string{}
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		link, ok := s.Attr("href")
		if ok && cs.ValidLink(link) {
			links = append(links, link)
		}

	})
	return links, nil
}

// Save saves the page content into the designated path
func (cs *CrawlerService) Save(ctx context.Context, name string, content []byte) error {
	fullPath := fmt.Sprintf("%s/%s%s", cs.destPath, name, ext)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(fullPath), 0700); err != nil {
			return err
		}
	}
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return err
	}
	return nil
}

// ValidLink checks if the path is a valid link by checking the prefix
func (cs *CrawlerService) ValidLink(link string) bool {
	ok := strings.HasPrefix(link, cs.basePath)
	if !ok {
		return false
	}
	newBasePath := cs.basePath
	if !strings.HasSuffix(newBasePath, "/") {
		newBasePath += "/"
	}
	return strings.HasPrefix(link, newBasePath)
}

// Visited checks if this path has been visited
func (cs *CrawlerService) Visited(link string) bool {
	_, ok := cs.visited[link]
	if ok {
		return true
	}
	savePath := savePathFromLink(link, cs.basePath)
	if savePath == "" {
		return true
	}

	_, err := os.Stat(fmt.Sprintf("%s/%s%s", cs.destPath, savePath, ext))
	return err == nil // return if the file exists or not
}

func savePathFromLink(link, basePath string) string {
	res, ok := strings.CutPrefix(link, basePath)
	if !ok || res == "" {
		return "home"
	}
	return strings.ReplaceAll(strings.TrimPrefix(res, "/"), "/", "_")
}

// Process takes a link, processes it and returns unvisited valid links
func (cs *CrawlerService) Process(ctx context.Context, link string) ([]string, error) {
	if cs.Visited(link) {
		// no need to return error for visited links as it could result in noise
		return nil, nil
	}
	var result []string
	res, err := cs.Crawl(ctx, link)
	if err != nil {
		return nil, err
	}

	content, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	fileName := savePathFromLink(link, cs.basePath)

	if err := cs.Save(ctx, fileName, content); err != nil {
		return nil, err
	}

	links, err := cs.ExtractLinks(ctx, bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}

	for _, v := range links {
		if ok := cs.ValidLink(v); ok && !cs.Visited(v) {
			result = append(result, v)
		}
	}

	return result, nil
}
