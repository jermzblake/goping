package main

import (
	"context"
	"flag"
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

// NewClient configures a tuned HTTP client for high-concurrency reuse.
// This is your "Infrastructure Configuration".
func NewHTTPClient() *http.Client {
    t := http.DefaultTransport.(*http.Transport).Clone()

    // Tune pooling while preserving useful defaults from DefaultTransport.
    t.MaxIdleConns = 100
    t.MaxIdleConnsPerHost = 10
    t.MaxConnsPerHost = 20
    t.IdleConnTimeout = 90 * time.Second

    // Optional guards for slow endpoints.
    t.TLSHandshakeTimeout = 5 * time.Second
    t.ResponseHeaderTimeout = 5 * time.Second

    return &http.Client{
        Transport: t,
        // Keep this unset (zero) so request context timeout is the single source of truth.
    }
}

func ping(ctx context.Context, client *http.Client, method string, url string) Result {
	start := time.Now()

	// Create a request with the provided context
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return Result{URL: url, Error: err}
	}

	resp, err := client.Do(req)
	if err != nil {
		return Result{URL: url, Error: err}
	}
	defer resp.Body.Close()

	latency := time.Since(start)
	return Result{URL: url, Status: resp.StatusCode, Latency: latency}
}

func worker(ctx context.Context, client *http.Client, method string, jobs <-chan string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range jobs {
		// Individual timeout per request
		reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		result := ping(reqCtx, client, method, url)
		cancel()
		results <- result
	}
}

func reporter(results <-chan Result) {
	for result := range results {
		if result.Error != nil {
			fmt.Printf("❌ %-30s | ERROR: %v\n", result.URL, result.Error)
			continue
		}
		
		statusEmoji := "✅"
		if result.Status >= 400 {
			statusEmoji = "⚠️"
		}
		
		fmt.Printf("%s %-30s | %d | %7v\n", statusEmoji, result.URL, result.Status, result.Latency)
	}
}

func main() {
workerCount := flag.Int("w", 5, "Number of workers")
        method := flag.String("m", "GET", "HTTP method (GET, HEAD)")
	flag.Parse()
	urls := flag.Args()

	if len(urls) == 0 {
		fmt.Println("No URLs provided. Using default list.")
		urls = []string{
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
}

	client := NewHTTPClient() // Create a single shared HTTP client for all workers
	if tr, ok := client.Transport.(*http.Transport); ok {
		defer tr.CloseIdleConnections()
	}

	jobs := make(chan string, len(urls))
	results := make(chan Result, len(urls))
	var wg sync.WaitGroup

	// 1. Start the Reporter Goroutine first (Consumer)
	// We don't add this to the WaitGroup because it stops when the results channel is closed
	// TODO if you want Reporter to be in a separate goroutine, you should add a done channel to signal when it's finished, and wait for it in the main function before exiting.
	// go reporter(results)
	
	// 2. Start Workers (Producers) Goroutines
	ctx := context.Background() // Base context for all workers
	for w := 1; w <= *workerCount; w++ {
		wg.Add(1)
		go worker(ctx, client, *method, jobs, results, &wg)
	}

	// 3. Send Jobs (URLs) to the Jobs Channel
	for _, url := range urls {
		jobs <- url
	}
	close(jobs) // Signal that no more jobs will be sent

	// 4. Wait and Close Results in a separate goroutine
	go func() {
		wg.Wait() // Wait for all workers to finish
		close(results) // Close results channel after all workers are done
	}()

	reporter(results) // Start the reporter in the main goroutine to print results as they come in
	fmt.Println("\n--- Scan Complete ---")
}