## Recursive Web Crawler


This command line application takes in a root link (Defaults to `go.dev`) crawls the page and recursively parses all the links that matches the given base link. A directory (Defaults to `data/saves`) is also created to store all the html files that will be downloaded and crawled as well.

## Startup:

* `make run` starts up the with the default path
* `go run ./cmd/ -url=https://go.dev -dir=data/saves` with `url` being the base URL and `dir` is the path where the values should be saved.