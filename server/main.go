package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

const apiURL = "https://en.wikipedia.org/w/api.php?action=query&format=json&prop=extracts&exintro=true&titles=Go_(programming_language)"

type WikiResponse struct {
	Query struct {
		Pages map[string]struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

func main() {
	client := &http.Client{}

	resp, err := client.Get(apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Error: unexpected status code %d\n", resp.StatusCode)
		os.Exit(1)
	}

	var result WikiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
		os.Exit(1)
	}

	for _, page := range result.Query.Pages {
		fmt.Printf("Title: %s\n\nExtract:\n%s\n", page.Title, page.Extract)
		break // We only need the first (and likely only) page
	}
}