package main

import (
	"ZG06/site-uptime-checker/urlparser"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

type Result struct {
	URL        string
	StatusCode int
	Latency    time.Duration
	Err        error
}

func main() {
	filePath := flag.String("f", "urls.json", "a data file containing target URLs")
	methodFlag := flag.String("m", "GET", "specifies an HTTP request method to check the URLs (HEAD/GET)")
	flag.Parse()

	method := strings.ToUpper(*methodFlag)
	// Check the HTTP request method and exit if it isn't GET or HEAD.
	if method != "GET" && method != "HEAD" {
		fmt.Println("Unsupported HTTP request method. Use HEAD or GET")
		os.Exit(1)
	}

	data, err := urlparser.ReadFile(*filePath)
	if err != nil {
		fmt.Println("Error reading file", err)
		os.Exit(1)
	}

	urls, err := urlparser.ParseJSON(data)
	if err != nil {
		fmt.Println("Error parsing URLs from JSON file", err)
		os.Exit(1)
	}

	var results = make(chan Result, len(urls))

	var wg sync.WaitGroup

	for _, url := range urls {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- checkURL(url, method)
		}()
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var finalResults []Result
	for res := range results {
		finalResults = append(finalResults, res)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "URL\tSTATUS\tLATENCY\tERROR")
	fmt.Fprintln(w, "---\t------\t-------\t-----")

	for _, res := range finalResults {
		statusStr := fmt.Sprintf("%d", res.StatusCode)
		errStr := "OK"

		if res.Err != nil {
			statusStr = "N/A"
			errStr = res.Err.Error()
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", res.URL, statusStr, res.Latency, errStr)
	}

	w.Flush()
}

func checkURL(url, method string) Result {
	var result Result
	// Save
	result.URL = url

	// Create a request with the specified method
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		result.Err = err
		return result
	}

	// Start the timer.
	startTime := time.Now()
	// Execute the HTTP request.
	res, err := http.DefaultClient.Do(req)
	// Stop the timer and calculate how long it took to execute the request.
	result.Latency = time.Since(startTime).Round(time.Millisecond)

	if err != nil {
		result.Err = err
		return result
	}

	// Close the *http.Response body.
	defer res.Body.Close()

	// Save the status code from the response.
	result.StatusCode = res.StatusCode

	return result
}
