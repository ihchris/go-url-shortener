package main

import (
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type URLStorage struct {
	urls map[string]string
	mu   sync.RWMutex
}

var (
	storage = URLStorage{urls: make(map[string]string)}
	charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
    <title>URL Shortener</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        form { display: flex; gap: 10px; margin-bottom: 20px; }
        input[type="url"] { flex-grow: 1; padding: 8px; }
        button { padding: 8px 16px; background: #007bff; color: white; border: none; cursor: pointer; }
        .result { padding: 15px; background: #f8f9fa; border-radius: 4px; }
    </style>
</head>
<body>
    <h1>URL Shortener</h1>
    <form action="/shorten" method="post">
        <input type="url" name="url" placeholder="Enter URL" required>
        <button type="submit">Shorten</button>
    </form>
    {{if .ShortURL}}
    <div class="result">
        Short URL: <a href="{{.ShortURL}}">{{.ShortURL}}</a>
    </div>
    {{end}}
</body>
</html>
`))

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/shorten", shortenHandler)
	http.HandleFunc("/short/", redirectHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	homeTemplate.Execute(w, nil)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	originalURL := r.FormValue("url")
	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	shortCode := generateShortCode(6)
	storage.mu.Lock()
	storage.urls[shortCode] = originalURL
	storage.mu.Unlock()

	shortURL := "http://" + r.Host + "/short/" + shortCode
	data := struct{ ShortURL string }{ShortURL: shortURL}
	homeTemplate.Execute(w, data)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/short/")
	if code == "" {
		http.NotFound(w, r)
		return
	}

	storage.mu.RLock()
	originalURL, ok := storage.urls[code]
	storage.mu.RUnlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}

func generateShortCode(length int) string {
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[rand.Intn(len(charset))]
	}
	return string(code)
}
