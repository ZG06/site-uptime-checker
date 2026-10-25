package main

import (
	"ZG06/site-uptime-checker/urlparser"
	"flag"
	"fmt"
	"net/http"
	"os"
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
	flag.Parse()

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
			results <- checkURL(url)
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

func checkURL(url string) Result {
	var result Result
	result.URL = url

	startTime := time.Now()
	res, err := http.Get(url)
	result.Latency = time.Since(startTime).Round(time.Millisecond)

	if err != nil {
		result.Err = err
		return result
	}

	defer res.Body.Close()

	result.StatusCode = res.StatusCode

	return result
}
