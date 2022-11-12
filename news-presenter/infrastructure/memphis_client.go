package memphis_client

import (
	"encoding/json"
	"errors"
	"github.com/caarlos0/env/v6"
	"github.com/memphisdev/memphis.go"
	"github.com/nats-io/nats.go"
	"log"
	"news-presenter/models/newsapi"
	"time"
)

type config struct {
	Host     string `env:"MEMPHIS_HOST,required"`
	Username string `env:"MEMPHIS_USERNAME,required"`
	Token    string `env:"MEMPHIS_TOKEN,required"`
	Port     int    `env:"MEMPHIS_PORT" envDefault:"6666"`
}

func New() (*memphis.Conn, error) {
	cfg := &config{}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	c, err := memphis.Connect(
		cfg.Host, cfg.Username, cfg.Token,
		memphis.Port(cfg.Port),
		memphis.Reconnect(true),
		memphis.MaxReconnect(3))

	if err != nil {
		return nil, err
	}

	return c, nil
}

func ConsumeFromMemphis(fnHandler func(data any)) error {
	//TODO use connection pooling, rather than destroy after each call
	conn, errConn := New()
	//defer conn.Close()

	if errConn != nil {
		return errConn
	}

	consumer, errConsumerCreate := conn.CreateConsumer(
		"station-news",
		"news-presenter-01",
		memphis.ConsumerGenUniqueSuffix())

	if errConsumerCreate != nil {
		return errConsumerCreate
	}

	handler := func(msgs []*memphis.Msg, err error) {
		if err != nil {
			if errors.Is(err, nats.ErrTimeout) {
				log.Printf("nats.ErrTimeout: %v\n", err)
				return
			}

			log.Fatalf("Fetch failed: %v\n", err)
		}

		log.Println("[INFO] Message polled from broker, sending to handler")
		for _, msg := range msgs {
			pArticles := make([]any, 0)

			pArticleAsStr := string(msg.Data())
			errUnmarshall := json.Unmarshal([]byte(pArticleAsStr), &pArticles)
			if errUnmarshall != nil {
				log.Fatalf("Unmarshall failed: %v\n", errUnmarshall)
			}
			fnHandler(pArticles)
			msg.Ack()
		}
	}

	errConsumption := consumer.Consume(handler)
	if errConsumption != nil {
		return errConsumption
	}

	return nil
}

func FetchFromMemphis() ([]any, error) {
	//TODO use connection pooling, rather than destroy after each call
	conn, errConn := New()
	defer conn.Close()

	if errConn != nil {
		return nil, errConn
	}

	consumer, errConsumerCreate := conn.CreateConsumer(
		"station-news",
		"news-presenter-01",
		memphis.BatchSize(1),
		memphis.BatchMaxWaitTime(60*time.Second),
		memphis.ConsumerGenUniqueSuffix())

	if errConsumerCreate != nil {
		return nil, errConsumerCreate
	}

	msgs, errFetch := consumer.Fetch()
	if errFetch != nil {
		return nil, errFetch
	}

	pArticles := make([]any, 0)

	for _, msg := range msgs {
		pArticles := make([]newsapi.Article, 0)

		pArticleAsStr := string(msg.Data())

		errUnmarshall := json.Unmarshal([]byte(pArticleAsStr), &pArticles)
		if errUnmarshall != nil {
			log.Fatalf("Unmarshall failed: %v\n", errUnmarshall)
		}

		msg.Ack()
	}

	return pArticles, nil
}
