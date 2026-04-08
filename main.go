package main

import (
	"fmt"
	"log"
	"net/http"
	"welog/router"
)

func main() {
	mux := router.NewRouter()
	port := ":8080"
	fmt.Printf("WELOG API Server is running on port %s\n", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("서버 시작 실패")
	}
}
