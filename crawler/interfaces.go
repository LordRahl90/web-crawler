package crawler

import (
	"context"
	"net/http"
)

// Crawler
type Crawler interface {
	BaseURL() string
	Crawl(ctx context.Context, path string) (*http.Response, error)
	Save(ctx context.Context, name string, content []byte) error
	ExtractLinks(ctx context.Context, r []byte) ([]string, error)
	ValidLink(path string) bool
	Visited(path string) bool
	Process(ctx context.Context, link string) ([]string, error)
}
