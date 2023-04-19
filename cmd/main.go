package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
	"web-crawler/crawler"

	"github.com/rs/zerolog/log"
)

var (
	ticker            = time.NewTicker(1 * time.Second)
	destDir, baseLink string
)

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	flag.StringVar(&baseLink, "url", "https://go.dev", "Base URL to start crawling from")
	flag.StringVar(&destDir, "dir", "data/saves", "Destination directory where the sites should be saved")

	flag.Parse()

	fmt.Printf("\nURL: %s\tDir: %s\n\n", baseLink, destDir)

	linkChan := make(chan string)
	cs := crawler.New(baseLink, destDir)

	ctx := context.Background()

	// start 3 worker pools to read as links are being populated into the channel
	// no need to for sync.Waitgroup and the application will be termiated with ^C
	// which kills all routines anyways
	for i := 0; i < 3; i++ {
		go worker(ctx, fmt.Sprintf("Worker: %d", i), cs, linkChan)
	}

	linkChan <- baseLink

	<-sigs
	fmt.Printf("Application Terminated\n")
}

func worker(ctx context.Context, name string, cs crawler.Crawler, linkChan chan string) {
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("Processing completed\n")
			return

		case t := <-ticker.C:
			fmt.Println(name, "New Ticker: ", t)

		case msg := <-linkChan:
			fmt.Println("processing starts")
			res, err := cs.Process(ctx, msg)
			if err != nil {
				log.Err(err)
			}
			// populate the linkChannel with the newly returned links
			for _, v := range res {
				linkChan <- v
			}
		}
	}
}
