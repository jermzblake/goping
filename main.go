package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Result struct {
	URL string
	Status int
	Latency time.Duration
	Error error
}

func ping(ctx context.Context, url string) Result {
	start := time.Now()

	// Create a request with the provided context
	req, _ := http.NewRequestWithContext(ctx, "HEAD", url, nil)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: url, Error: err}
	}
	defer resp.Body.Close()

	latency := time.Since(start)
	return Result{URL: url, Status: resp.StatusCode, Latency: latency}
}

func worker(ctx context.Context, jobs <-chan string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range jobs {
		result := ping(ctx, url)
		results <- result
	}
}

func main() {
	urls := []string{
		"https://www.google.com",
		"https://www.facebook.com",
		"https://www.twitter.com",
		"https://www.linkedin.com",
		"https://www.github.com",
		"https://www.reddit.com",
		"https://www.stackoverflow.com",
		"https://www.medium.com",
		"https://www.netflix.com",
		"https://www.amazon.com",
		"https://www.apple.com",
		"https://www.microsoft.com",
		"https://www.ibm.com",
		"https://www.oracle.com",
		"https://www.salesforce.com",
		"https://www.adobe.com",
		"https://www.spotify.com",
		"https://www.airbnb.com",
		"https://www.uber.com",
		"https://invalid-url.test",
	}

	numWorkers := 3
	jobs := make(chan string, len(urls))
	results := make(chan Result, len(urls))

	var wg sync.WaitGroup

	// Create a global timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Start Workers (Goroutines)
	for w := 1; w <= numWorkers; w++ {
		wg.Add(1)
		go worker(ctx, jobs, results, &wg)
	}

	// 2. Send Jobs (URLs) to the Jobs Channel
	for _, url := range urls {
		jobs <- url
	}
	close(jobs) // Signal that no more jobs will be sent

	// 3. Wait and Close Results in a separate goroutine
	go func() {
		wg.Wait() // Wait for all workers to finish
		close(results) // Close results channel after all workers are done
	}()

	// 4. Collect and Print Results
	for result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %-25s | Error: %v\n", result.URL, result.Error)
		} else {
			fmt.Printf("✅ %-25s | Status: %d | Latency: %v\n", result.URL, result.Status, result.Latency)
		}
	}

}