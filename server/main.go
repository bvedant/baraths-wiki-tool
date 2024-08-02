package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

const apiUrl = "https://en.wikipedia.org/w/api.php"
const wikipediaBaseURL = "https://en.wikipedia.org"

type PageData struct {
	Title   string
	Content template.HTML
}

func fetchWikipediaContent(title string) (PageData, error) {
	params := url.Values{}
	params.Add("action", "parse")
	params.Add("format", "json")
	params.Add("page", title)
	params.Add("prop", "text")

	resp, err := http.Get(apiUrl + "?" + params.Encode())
	if err != nil {
		return PageData{}, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
	}()

	var result struct {
		Parse struct {
			Title string `json:"title"`
			Text  struct {
				Content string `json:"*"`
			} `json:"text"`
		} `json:"parse"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PageData{}, err
	}

	if result.Parse.Title == "" {
		return PageData{}, fmt.Errorf("no content found for %s", title)
	}

	// Parse the HTML content
	doc, err := html.Parse(strings.NewReader(result.Parse.Text.Content))
	if err != nil {
		return PageData{}, err
	}

	// Remove the infobox and edit links, and process links
	removeInfobox(doc)
	removeEditLinks(doc)
	processLinks(doc)

	// Convert the modified HTML back to a string
	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return PageData{}, err
	}

	return PageData{
		Title:   result.Parse.Title,
		Content: template.HTML(buf.String()),
	}, nil
}

func removeInfobox(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "table" {
		for _, a := range n.Attr {
			if a.Key == "class" && strings.Contains(a.Val, "infobox") {
				if n.Parent != nil {
					n.Parent.RemoveChild(n)
				}
				return
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		removeInfobox(c)
	}
}

func removeEditLinks(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "span" {
		for _, a := range n.Attr {
			if a.Key == "class" && strings.Contains(a.Val, "mw-editsection") {
				if n.Parent != nil {
					n.Parent.RemoveChild(n)
				}
				return
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		removeEditLinks(c)
	}
}

func processLinks(n *html.Node) {
	if n.Type == html.ElementNode && n.Data == "a" {
		for i, a := range n.Attr {
			if a.Key == "href" {
				if strings.HasPrefix(a.Val, "/wiki/") {
					n.Attr[i].Val = wikipediaBaseURL + a.Val
				}
			}
		}
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		processLinks(c)
	}
}

func handleWikipediaRequest(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		http.Error(w, "Missing 'title' parameter", http.StatusBadRequest)
		return
	}

	pageData, err := fetchWikipediaContent(title)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(filepath.Join("templates", "index.html"))
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.Execute(w, pageData); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Error rendering page", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Serve static files
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Handle Wikipedia requests
	http.HandleFunc("/pageContent", handleWikipediaRequest)

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
