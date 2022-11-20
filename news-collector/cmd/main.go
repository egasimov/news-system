package main

import (
	"context"
	"fmt"
	"github.com/google/uuid"
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
	ctx, stop := signal.NotifyContext(appContext, os.Interrupt, syscall.SIGTERM /*, syscall.SIGKILL*/)
	defer stop()

	logAppStartUp(ctx)

	// channel used for data exchange between producer-goroutine and consume-goroutine
	msgChannel := make(chan any, 20)
	var wg sync.WaitGroup

	// Add producer like worker goroutines
	// httpclient worker goroutines should consume from external service, and pass it into channel
	StartFetchNewsService(ctx, &wg, msgChannel, 3)

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
func startPublishMessage(ctx context.Context, wg *sync.WaitGroup, dataChannel chan any) {
	grName := fmt.Sprintf("%s-%s", "startPublishMessage", uuid.New().String())
	defer func() {
		log.Printf("[INFO]-[worker:%s] gracefully made shutdown\n",
			grName)
		wg.Done()
	}()

	log.Printf("[INFO]-[worker:%s] started ...\n",
		grName)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[ERROR]-[worker:%s] recevied termination signal to via context",
				grName)

			// Consume all messages before close channel, to avoid data loss
			if len(dataChannel) != 0 {
				log.Printf("[INFO]-[worker:%s] started draining data channel(left item count: %d).. \n",
					grName, len(dataChannel))

				for leftValueItem := range dataChannel {
					errPublish := memphis_client.PublishToMemphis(ctx, leftValueItem)
					if errPublish != nil {
						log.Fatalln(errPublish)
					}
					log.Printf("[INFO]-[worker:%s] draining data channel(left item count: %d).. \n",
						grName, len(dataChannel))
				}

				log.Printf("[INFO]-[worker:%s] finished draining data channel(left item count: %d).. \n",
					grName, len(dataChannel))
			}
			val, ok := <-dataChannel
			if !ok || val == nil {
				log.Printf("[INFO]-[worker:%s] channel already closed, terminating goroutine\n",
					grName)
			} else {
				log.Printf("[INFO]-[worker:%s] channel is open, terminating goroutine\n",
					grName)
			}
			return
		case val, ok := <-dataChannel:
			if !ok || val == nil {
				log.Fatalf("[ERROR]-[worker:%s] unable to fetch value from channel, because closed\n",
					grName)
			}
			errPublish := memphis_client.PublishToMemphis(ctx, val)

			if errPublish != nil {
				log.Fatalln(errPublish)
			}

		}
	}
}

// Intended to fetch the news from predefined sources, and write it into internal in-memory channel
func startFetchNews(ctx context.Context, wg *sync.WaitGroup, dataChannel chan any, onceChCloser *sync.Once) {
	grName := fmt.Sprintf("%s-%s", "startFetchNews", uuid.New().String())
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[ERROR]-[worker-%s] Recovered Error:\n", grName)
			fmt.Println("Recovered. Error:\n", r)
		}
		log.Printf("[INFO]-[worker:%s] gracefully made shutdown\n",
			grName)
		wg.Done()
	}()

	cfg, errCfg := config.GetConfig(ctx)
	if errCfg != nil {
		log.Fatalf("[ERROR]-[worker:%s] failed because of: %s \n",
			grName, errCfg)
	}

	durationFrequency, err := time.ParseDuration(cfg.CollectorSettings.ScrapeInterval)
	if err != nil {
		log.Fatalf("[ERROR]-[worker:%s] failed because of: %s\n", grName, err)
	}

	log.Printf("[INFO]-[worker:%s] started ...\n", grName)

	for {
		select {
		case <-ctx.Done():
			log.Printf("[ERROR]-[worker:%s] recevied termination signal via context", grName)

			// Close the channel, when there is no item. to prevent data loss
			onceChCloser.Do(func() { close(dataChannel) })

			log.Printf("[ERROR]-[worker:%s] closed the channel, channel left item count: %d\n",
				grName, len(dataChannel))
			return
		case <-time.After(durationFrequency):
			var data any
			var err error
			switch cfg.CollectorSettings.SourceOfNews {
			case "THEGUARDIAN":
				data, err = theguardian_client.FetchGuardianNews(ctx)
			case "NEWSAPI":
				data, err = newsapi_client.FetchNews(ctx)
			default:
				log.Printf("[WARNING] - [worker:%s] no source of news found, so going with default one: NEWSAPI \n",
					grName)
				data, err = newsapi_client.FetchNews(ctx)
			}

			if err != nil {
				log.Fatalln(err)
			}

			dataChannel <- data
		}
	}
}

// TODO in the future, add multiple based on source of site it consuming
func StartFetchNewsService(ctx context.Context, wg *sync.WaitGroup, dataChannel chan any, workerGrCount int) {
	onceDataChannelCloser := sync.Once{}
	for i := 0; i < workerGrCount; i++ {
		wg.Add(1)
		go startFetchNews(ctx, wg, dataChannel, &onceDataChannelCloser)
	}
}
