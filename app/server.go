package main

import (
	"fmt"
	"net/http"
	"os"
	"sol/llm_app/endpoints"
)

func RunServer() {
	fmt.Printf("Starting server...\n")

	http.HandleFunc("/", endpoints.GetRoot)
	http.HandleFunc("/test", endpoints.GetTest)
	err := http.ListenAndServe(":5001", nil)

	if err != nil {
		fmt.Printf("error: %v", err)
		os.Exit(1)
	}
}

// POST/DELETE chat
// ws /chat
