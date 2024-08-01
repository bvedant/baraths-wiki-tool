package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const wikipediaAPIURL = "https://en.wikipedia.org/w/api.php?action=query&format=json&prop=extracts&exintro=true&titles=Go_(programming_language)"

type WikiResponse struct {
	Query struct {
		Pages map[string]struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

type CleanResponse struct {
	Title       string `json:"title"`
	FirstParagraph string `json:"first_paragraph"`
}

func fetchWikipediaData() (*CleanResponse, error) {
	client := &http.Client{}

	resp, err := client.Get(wikipediaAPIURL)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result WikiResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	for _, page := range result.Query.Pages {
		cleanText := cleanHTML(page.Extract)
		paragraphs := strings.Split(cleanText, "\n")
		if len(paragraphs) > 0 {
			return &CleanResponse{
				Title:       page.Title,
				FirstParagraph: paragraphs[0],
			}, nil
		}
	}

	return nil, fmt.Errorf("no content found")
}

func cleanHTML(html string) string {
	// Remove HTML tags
	re := regexp.MustCompile("<[^>]*>")
	cleanText := re.ReplaceAllString(html, "")

	// Remove extra whitespace
	cleanText = strings.TrimSpace(cleanText)
	cleanText = regexp.MustCompile(`\s+`).ReplaceAllString(cleanText, " ")

	return cleanText
}

func wikiHandler(w http.ResponseWriter, r *http.Request) {
	data, err := fetchWikipediaData()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/api/wiki", wikiHandler)

	port := "8080"
	fmt.Printf("Server is running on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}