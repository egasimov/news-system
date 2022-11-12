package theguardian_client

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
	"news-collector/models/theguardian"
	"time"
)

func FetchGuardianNews(ctx context.Context) ([]theguardian.Result, error) {
	cfg := ctx.Value(constants.CONFIG_CTX_KEY).(*config.Config)

	uri := fmt.Sprintf("%s/%s&api-key=%s", cfg.TheGuardianConfig.BaseUrl,
		cfg.TheGuardianConfig.GetNewsPath,
		cfg.TheGuardianConfig.ApiKey,
	)
	//"https://content.guardianapis.com/search?order-by=newest&api-key=5bd4bdab-60d8-44ae-8ceb-9d0638c0637c"
	req, err := http.NewRequestWithContext(ctx, cfg.TheGuardianConfig.HttpMethod, uri,
		bytes.NewBuffer([]byte("")))

	if err != nil {
		return nil, fmt.Errorf("theguardian client error: %w", err)
	}

	log.Println("[INFO]-[theguardian_client] Starting to fetch data from guardian api")

	data, err := doFetch(req)
	if err != nil {
		return nil, fmt.Errorf("theguardian client error: %w", err)
	}

	log.Println("[INFO]-[theguardian_client] Finished fetching data from guardian api")

	return data, nil
}

func doFetch(req *http.Request) ([]theguardian.Result, error) {
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

	apiResponse := theguardian.ApiResponse{}

	jsonErr := json.Unmarshal(body, &apiResponse)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return apiResponse.Response.Results, nil
}
