package main

import (
	"fmt"
	"io"
	"net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	fmt.Fprintf(w, "Hello, World! Received: %s", body)
}

func main() {
	http.HandleFunc("/", helloHandler)

	fmt.Println("Server is starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
