package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const apiUrl = "https://en.wikipedia.org/w/api.php?action=query&format=json&prop=extracts&explaintext=1&titles=%s"

type WikiResponse struct {
	Query struct {
		Pages map[string]struct {
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

type CleanResponse struct {
	Title   string   `json:"title"`
	Content []string `json:"content"`
}

func fetchWikipediaData(title string) (*CleanResponse, error) {
	encodedTitle := url.QueryEscape(title)
	fullUrl := fmt.Sprintf(apiUrl, encodedTitle)

	client := &http.Client{}

	resp, err := client.Get(fullUrl)
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
		cleanText := cleanContent(page.Extract)
		return &CleanResponse{
			Title:   page.Title,
			Content: cleanText,
		}, nil
	}

	return nil, fmt.Errorf("no content found")
}

func cleanContent(content string) []string {
	paragraphs := strings.Split(content, "\n")

	var cleanParagraphs []string
	for _, p := range paragraphs {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			cleanParagraphs = append(cleanParagraphs, trimmed)
		}
	}

	return cleanParagraphs
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "Missing 'title' query parameter", http.StatusBadRequest)
		return
	}

	data, err := fetchWikipediaData(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func main() {
	http.HandleFunc("/api/info", infoHandler)

	port := "8080"
	fmt.Printf("Server is running on http://localhost:%s\n", port)
	fmt.Println("Use /api/info?title=Your_Wikipedia_Page_Title to get information")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}