package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func text(w http.ResponseWriter, r *http.Request) {
	header := w.Header()
	header.Set("Content-Type", "application/json")
	//fmt.Fprintln(w, "hello world")
	w.Write([]byte("hello world"))
	w.WriteHeader(http.StatusOK)
}

func index(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html := `<doctype html>
        <html>
        <head>
          <title>Hello World</title>
        </head>
        <body>
        <p>
          <a href="/welcome">Welcome</a> |  <a href="/message">Message</a>
        </p>
        </body>
      </html>`
	fmt.Fprintln(w, html)
}

func hello(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/hello" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		http.ServeFile(w, r, "form.html")
	case http.MethodPost:
		result, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
		fmt.Printf("%s\n", result)

		// Call ParseForm() to parse the raw query and update r.PostForm and r.Form.
		if err := r.ParseForm(); err != nil {
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			return
		}
		fmt.Fprintf(w, "Post from website! r.PostFrom = %v\n", r.PostForm)
		name := r.FormValue("name")
		address := r.FormValue("address")
		fmt.Fprintf(w, "Name = %s\n", name)
		fmt.Fprintf(w, "Address = %s\n", address)
	default:
		fmt.Fprintf(w, "Sorry, only GET and POST methods are supported.")
	}
}

//中间件
func middlewareHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
		log.Printf("Comleted %s in %v", r.URL.Path, time.Since(start))
	})
}

func main() {
	// mux := http.NewServeMux()
	// mux.HandleFunc("/", index)

	http.Handle("/hello", middlewareHandler(http.HandlerFunc(hello)))
	http.Handle("/", middlewareHandler(http.HandlerFunc(index)))
	http.Handle("/text", middlewareHandler(http.HandlerFunc(text)))
	http.ListenAndServe(":8000", nil)

	server := &http.Server{
		Addr:         ":8000",
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		//Handler:      mux,
		MaxHeaderBytes: 1 << 20,
	}
	server.ListenAndServe()

	return
}
