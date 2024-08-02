package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

const apiUrl = "https://en.wikipedia.org/w/api.php?action=parse&format=json&prop=text&page=%s"

type WikiResponse struct {
	Parse struct {
		Title string `json:"title"`
		Text  struct {
			Content string `json:"*"`
		} `json:"text"`
	} `json:"parse"`
}

type ContentItem struct {
	Type    string
	Content template.HTML
}

type PageData struct {
	Title   string
	Content template.HTML
}

func fetchWikipediaData(title string) (*PageData, error) {
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

	return &PageData{
		Title:   result.Parse.Title,
		Content: template.HTML(result.Parse.Text.Content),
	}, nil
}

func infoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.URL.Query().Get("title")
	if title == "" {
		renderTemplate(w, "index.html", nil)
		return
	}

	data, err := fetchWikipediaData(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "result.html", data)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	templatesDir := filepath.Join("templates")
	t, err := template.ParseFiles(filepath.Join(templatesDir, tmpl))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

func main() {
	http.HandleFunc("/", infoHandler)

	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	port := "8080"
	fmt.Printf("Server is running on http://localhost:%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
