package newsapi_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"news-collector/config"
	"news-collector/constants"
	"news-collector/models/newsapi"
	"time"
)

func FetchNews(ctx context.Context) ([]newsapi.Article, error) {
	cfg := ctx.Value(constants.CONFIG_CTX_KEY).(*config.Config)

	req, err := http.NewRequestWithContext(ctx, cfg.TheNewsApiConfig.HttpMethod,
		fmt.Sprintf("%s/%s", cfg.TheNewsApiConfig.BaseUrl, cfg.TheNewsApiConfig.GetNewsPath),
		bytes.NewBuffer([]byte("")))
	if err != nil {
		return nil, fmt.Errorf("newsapi client error: %w", err)
	}

	log.Println("[INFO]-[newsapi_client] Starting to fetch data from news api")

	data, err := doFetch(req)
	if err != nil {
		return nil, fmt.Errorf("newsapi client error: %w", err)
	}

	log.Println("[INFO]-[newsapi_client] Finished fetching data from news api")

	return data, nil
}

func doFetch(req *http.Request) ([]newsapi.Article, error) {
	client := http.Client{
		Timeout: time.Duration(60) * time.Second,
	}
	resp, getErr := client.Do(req)

	if getErr != nil {
		return nil, getErr
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	apiResponse := newsapi.ApiResponse{}

	jsonErr := json.Unmarshal(body, &apiResponse)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return apiResponse.Articles[0:10], nil
}
