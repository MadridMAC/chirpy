package main

import "net/http"

func main() {
	new_mux := http.NewServeMux()
	new_mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("."))))

	new_mux.HandleFunc("/healthz", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	server := http.Server{
		Addr:    ":8080",
		Handler: new_mux,
	}
	server.ListenAndServe()
}
