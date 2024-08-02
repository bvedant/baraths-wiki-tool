package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
)

type WikipediaPage struct {
	Parse struct {
		Text json.RawMessage `json:"text"`
	} `json:"parse"`
}

type Page struct {
	Query string
	HTML  template.HTML
}

func main() {
	// Create a new HTTP request multiplexer
	mux := http.NewServeMux()

	// Define a handler function for the root URL
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Get the search query parameter
		query := r.URL.Query().Get("q")
		if query == "" {
			query = "Main_Page"
		}

		// Fetch the article content from Wikipedia
		url := fmt.Sprintf("https://en.wikipedia.org/w/api.php?action=parse&page=%s&format=json", query)
		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			err := resp.Body.Close()
			if err != nil {
				log.Println(err)
			}
		}()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Unmarshal the JSON response
		var page WikipediaPage
		err = json.Unmarshal(body, &page)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Unmarshal the text content
		var textContent map[string]string
		err = json.Unmarshal(page.Parse.Text, &textContent)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Replace relative URLs with absolute URLs
		html := textContent["*"]
		html = strings.ReplaceAll(html, `href="/wiki/`, `href="https://en.wikipedia.org/wiki/`)
		html = strings.ReplaceAll(html, `href="#`, `href="https://en.wikipedia.org/wiki/`+query+`#`)

		// Render the template
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Write the HTML content to the response writer
		w.Header().Set("Content-Type", "text/html")
		err = tmpl.Execute(w, Page{Query: query, HTML: template.HTML(html)})
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})

	// Start the web server
	fmt.Println("Server listening on port 8080")
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
