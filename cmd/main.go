package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"web-crawler/crawler"
)

var (
	destDir, baseLink string
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	flag.StringVar(&baseLink, "url", "https://go.dev", "Base URL to start crawling from")
	flag.StringVar(&destDir, "dir", "data/saves", "Destination directory where the sites should be saved")

	flag.Parse()

	linkChan := make(chan string, 1)
	cs := crawler.New(baseLink, destDir)
	var wg sync.WaitGroup

	ctx, stop := context.WithCancel(context.Background())

	// start some worker pools to read as links are being populated into the channel.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go worker(ctx, &wg, fmt.Sprintf("Worker: %d", i), cs, linkChan)
	}

	linkChan <- baseLink

	<-sigs
	stop()

	fmt.Println("Waiting for all routines to return")
	wg.Wait()
	close(linkChan)
	fmt.Println("Application Terminated successfully")
}

func worker(ctx context.Context, wg *sync.WaitGroup, name string, cs crawler.Crawler, linkChan chan string) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return

		case v := <-linkChan:
			res, err := cs.Process(ctx, v)
			if err != nil {
				panic(err)
			}
			go func() {
				for _, v := range res {
					linkChan <- v
				}
			}()
		}
	}
}
