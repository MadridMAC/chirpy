package main

import "net/http"

func main() {
	new_mux := http.NewServeMux()
	server := http.Server{
		Addr:    ":8080",
		Handler: new_mux,
	}
	server.ListenAndServe()
}
