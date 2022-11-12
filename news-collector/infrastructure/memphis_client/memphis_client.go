package memphis_client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/memphisdev/memphis.go"
	"log"
	"news-collector/config"
)

func New(ctx context.Context) (*memphis.Conn, error) {
	cfg, err := config.GetConfig(ctx)
	if err != nil {
		return nil, err
	}

	conn, err := memphis.Connect(cfg.Memphis.Host,
		cfg.Memphis.Username,
		cfg.Memphis.Token, memphis.Port(cfg.Memphis.Port),
		memphis.Reconnect(true),
		memphis.MaxReconnect(3))
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func PublishToMemphis(ctx context.Context, message any) error {
	//TODO use connection pooling, rather than destroy after each call
	cfg, errCfg := config.GetConfig(ctx)
	if errCfg != nil {
		return errCfg
	}

	conn, errConn := New(ctx)
	if errConn != nil {
		return errConn
	}
	defer conn.Close()

	producer, errProducerCreate := conn.CreateProducer(cfg.Memphis.NewsStation,
		fmt.Sprintf("%s_%s", cfg.Application.Name, cfg.Application.Version))
	if errProducerCreate != nil {
		return errProducerCreate
	}

	result, errJson := json.Marshal(message)
	if errJson != nil {
		return errJson
	}

	log.Println("[INFO]-[memphis_client] Starting to publish data to memphis")

	err := producer.Produce(result)
	if err != nil {
		return err
	}

	log.Println("[INFO]-[memphis_client] Published successfully data into memphis")

	return nil
}
