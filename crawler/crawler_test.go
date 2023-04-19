package crawler

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var destPath = "testdata/saves"

func TestMain(m *testing.M) {
	code := 1
	defer func() {
		os.RemoveAll("./testdata/saves")
		os.Exit(code)
	}()

	code = m.Run()
}

func TestCrawler(t *testing.T) {
	ctx := context.Background()
	path := "https://go.dev/doc/tutorial/web-service-gin"

	cs := New(path, destPath)

	res, err := cs.Crawl(ctx, path)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.True(t, cs.Visited(path))
}

func TestSavePathFromLink(t *testing.T) {
	path := "https://go.dev/"

	cs := New(path, destPath)
	require.NotNil(t, cs)

	link := "https://go.dev/doc/tutorial/web-service-gin"
	exp := "doc_tutorial_web-service-gin"

	assert.Equal(t, exp, savePathFromLink(link, path))
}

func TestSavePathFromLink_Home(t *testing.T) {
	path := "https://go.dev/"

	cs := New(path, destPath)
	require.NotNil(t, cs)

	exp := "home"

	assert.Equal(t, exp, savePathFromLink(path, path))
}

func TestResume(t *testing.T) {
	path := "https://go.dev"
	link := "https://go.dev/doc"
	content := `
	<html>
    <head>
        <title>Demo</title>
    </head>
    <body>
        <h1>Hello, World!</h1>
    </body>
</html>`

	cs := New(path, destPath)
	require.NotNil(t, cs)

	err := cs.Save(context.TODO(), "doc", []byte(content))
	require.NoError(t, err)

	assert.True(t, cs.Visited(link)) // visited as there is a file already
	assert.False(t, cs.Visited("https://go.dev/docs"))
}

func TestExtractLinks(t *testing.T) {
	ctx := context.Background()
	link := "https://start.url/abc"

	defer mockHttp(t, link, "landing.html")()

	cs := New(link, destPath)
	require.NotNil(t, cs)

	res, err := cs.Crawl(ctx, link)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.NoError(t, err)
	assert.NotEmpty(t, res)

	links, err := cs.ExtractLinks(ctx, res.Body)
	require.NoError(t, err)
	assert.Len(t, links, 4)
}

func TestValidLink(t *testing.T) {
	path := "https://go.dev/doc"
	cs := New(path, destPath)
	require.NotNil(t, cs)
	table := []struct {
		name, args string
		exp        bool
	}{
		{
			name: "Valid",
			args: "https://go.dev/doc/tutorial/web-service-gin",
			exp:  true,
		},
		{
			name: "Valid",
			args: "https://go.devs/doc/tutorial/web-service-gin",
			exp:  false,
		},
		{
			name: "Valid",
			args: "https://go.dev/docs/tutorial/web-service-gin",
			exp:  false,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			got := cs.ValidLink(tt.args)
			assert.Equal(t, tt.exp, got)
		})
	}

	assert.Equal(t, path, cs.BaseURL())
}

func TestSave(t *testing.T) {
	ctx := context.Background()
	path := "https://go.dev/doc"
	cs := New(path, destPath)
	require.NotNil(t, cs)

	content := `
	<html>
    <head>
        <title>Demo</title>
    </head>
    <body>
        <h1>Hello, World!</h1>
    </body>
</html>`
	name := "demo"

	require.NoError(t, cs.Save(ctx, name, []byte(content)))

	src := fmt.Sprintf("%s/%s%s", destPath, name, ext)
	_, err := os.Stat(src)
	require.NoError(t, err) // file should exist

	link := path + "/" + name
	// link is visited as file exists
	assert.True(t, cs.Visited(link))
	require.NoError(t, os.Remove(src))
}

func TestProcess(t *testing.T) {
	ctx := context.Background()
	link := "https://start.url/abc/foo"
	cs := New("https://start.url/abc", destPath)

	defer mockHttp(t, link, "foo.html")()

	links, err := cs.Process(ctx, link)
	require.NoError(t, err)
	require.Len(t, links, 2)
}

func mockHttp(t *testing.T, link, content string) func() {
	t.Helper()
	httpmock.Activate()

	b, err := os.ReadFile("./testdata/demo/" + content)
	require.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, link, httpmock.NewStringResponder(http.StatusOK, string(b)))

	return httpmock.DeactivateAndReset
}
