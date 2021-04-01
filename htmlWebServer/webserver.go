package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

var (
	PORT = "8080"
)

type Middleware func(http.HandlerFunc) http.HandlerFunc

// Logging logs all requests with its path and the time it took to process
func Logging() Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			// Do middleware things
			start := time.Now()
			defer func() { log.Println(r.URL.Path, time.Since(start)) }()

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

// Method ensures that url can only be requested with a specific method, else returns a 400 Bad Request
func Method(m string) Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			// Do middleware things
			if r.Method != m {
				http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
				return
			}

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

// Chain applies middlewares to a http.HandlerFunc
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

func main() {
	http.HandleFunc("/",
		Chain(
			func(w http.ResponseWriter, r *http.Request) {
				f, _ := ioutil.ReadFile("html/mainPage.html")
				fmt.Fprintf(w, "%s", f)
			},
			Method("GET"),
			Logging()))

	http.HandleFunc("/user/",
		Chain(
			func(w http.ResponseWriter, r *http.Request) {
				f, _ := ioutil.ReadFile("html/userPage.html")
				fmt.Fprintf(w, "%s", f)
			},
			Method("GET"),
			Logging()))

	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Printf("Starting Server on Port: %s", PORT)
	log.Fatalln(http.ListenAndServe(":"+PORT, nil))

	http.ListenAndServe(":"+PORT, nil)
}
