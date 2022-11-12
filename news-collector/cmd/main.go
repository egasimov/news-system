package main

import (
	"context"
	"log"
	"news-collector/config"
	"news-collector/infrastructure/memphis_client"
	"news-collector/infrastructure/newsapi_client"
	"news-collector/infrastructure/theguardian_client"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	cfg, errCfg := config.NewConfig()
	if errCfg != nil {
		log.Fatalln(errCfg)
	}

	appContext := config.PutConfig(context.Background(), cfg)

	// Termination signal should be handled by child go routines via passed context
	ctx, stop := signal.NotifyContext(appContext, os.Interrupt, syscall.SIGTERM)
	defer stop()

	logAppStartUp(ctx)

	// channel used for data exchange between producer-goroutine and consume-goroutine
	msgChannel := make(chan any, 1)
	var wg sync.WaitGroup

	// Add producer like worker goroutines
	//TODO in the future, add multiple based on source of site it consuming
	wg.Add(1)
	// httpclient worker goroutine should consume from external service, and pass it into channel
	go startFetchNews(ctx, &wg, msgChannel)

	//Add consumer like worker goroutine
	//TODO in the future, add multiple based on source of site it consuming
	wg.Add(1)
	// msg-publisher-worker should consume message and publish into memphis station
	go startPublishMessage(ctx, &wg, msgChannel)

	log.Println("[INFO]-[MAIN] Application started...")

	wg.Wait()

	log.Println("[INFO]-[MAIN] Application stopped...")
}

// Intended to log informative message about running application
func logAppStartUp(ctx context.Context) {
	cfg, errCfg := config.GetConfig(ctx)
	if errCfg != nil {
		log.Fatalln(errCfg)
	}
	log.Printf(
		"App name: %s, "+
			"App Version: %s, "+
			"Environment: %s",
		cfg.Application.Name,
		cfg.Application.Version,
		cfg.Application.Env)
}

// Intended to read data from in-memory channel(where data filled by #startFetchNews)
// and publish it into external queue system(namely: memphis)
func startPublishMessage(ctx context.Context, wg *sync.WaitGroup, ch chan any) {
	defer wg.Done()

	log.Println("[INFO]-[worker:startPublishMessage] started ...")

	for {
		select {
		case <-ctx.Done():
			log.Printf("[ERROR]-[worker:startPublishMessage] stopped due to receiving signal to via context")

			//TODO consider consume all messages before close channel, to avoid data loss
			close(ch)
		case val, ok := <-ch:
			if !ok || val == nil {
				log.Fatalln("[ERROR]-[worker:startPublishMessage] unable to fetch value from channel, because closed")
			}
			errPublish := memphis_client.PublishToMemphis(ctx, val)
			if errPublish != nil {
				log.Fatalln(errPublish)
			}

		}
	}

	log.Println("[INFO]-[worker:startPublishMessage] finished ...")
}

// Intended to fetch the news from predefined sources, and write it into internal in-memory channel
func startFetchNews(ctx context.Context, wg *sync.WaitGroup, channel chan any) {
	defer wg.Done()

	cfg, errCfg := config.GetConfig(ctx)
	if errCfg != nil {
		log.Fatalf("[ERROR]-[worker:startFetchNews] failed because of: %s", errCfg)
	}

	durationFrequency, err := time.ParseDuration(cfg.CollectorSettings.ScrapeInterval)
	if err != nil {
		log.Fatalf("[ERROR]-[worker:startFetchNews] failed because of: %s", err)
	}

	log.Printf("[INFO]-[worker:startFetchNews] started ...")
	for {
		select {
		case <-ctx.Done():
			log.Printf("[ERROR]-[worker:startFetchNews] stopped due to receiving signal to via context")

			//TODO close the channel, when there is no item. to prevent data loss
			close(channel)
		case <-time.After(durationFrequency):
			var data any
			var err error
			switch cfg.CollectorSettings.SourceOfNews {
			case "THEGUARDIAN":
				data, err = theguardian_client.FetchGuardianNews(ctx)
			case "NEWSAPI":
				data, err = newsapi_client.FetchNews(ctx)
			default:
				log.Printf("[WARNING] no source of news found, so going with default one: NEWSAPI")
				data, err = newsapi_client.FetchNews(ctx)
			}

			if err != nil {
				panic(err)
			}

			channel <- data
		}
	}
	log.Println("[INFO]-[worker:startFetchNews] finished ...")
}
