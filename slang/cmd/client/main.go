package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

const address = "http://localhost:8080"
const rps = 1.0
const maxTime = 10 * time.Second

func main() {
	client := NewClient(&http.Client{}, rps)
	ctx, cancel := context.WithTimeout(context.Background(), maxTime)
	defer cancel()
	//fmt.Println(client.callEndpoint(ctx, "ball"))
	//fmt.Println(client.callEndpoint(ctx, "apple", "bread"))
	log.Println(client.callEndpoint(ctx, "atlanta", "baltimore", "charlotte"))
}

type Client struct {
	client  *http.Client
	limiter *rate.Limiter
}

func NewClient(client *http.Client, rps float64) *Client {
	return &Client{
		client:  client,
		limiter: rate.NewLimiter(rate.Limit(rps), 1),
	}
}

func (c *Client) callEndpoint(ctx context.Context, args ...string) ([]string, error) {
	var responses []string
	out := make(chan string, len(args))

	group, groupCtx := errgroup.WithContext(ctx)
	for i := range args {
		input := args[i]
		group.Go(func() error {
			log.Println("making request for", input)
			request, err := http.NewRequestWithContext(groupCtx, http.MethodPost, address, strings.NewReader(input))
			if err != nil {
				return fmt.Errorf("failed to create HTTP request: %w", err)
			}
			if waitErr := c.limiter.Wait(groupCtx); waitErr != nil {
				return fmt.Errorf("exceeded wait time: %w", waitErr)
			}
			body, err := makeRequest(c.client, request)
			if err != nil {
				return fmt.Errorf("failed to perform HTTP request: %w", err)
			}
			log.Println("received response for", input, body)
			out <- body
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
	}
	close(out)
	for response := range out {
		responses = append(responses, response)
	}
	return responses, nil
}

func makeRequest(client *http.Client, request *http.Request) (string, error) {
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to perform HTTP request: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received unexpected status code: %d", response.StatusCode)
	}
	if response.Body == nil {
		return "", nil
	}
	defer response.Body.Close()
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}
	return buf.String(), nil
}
