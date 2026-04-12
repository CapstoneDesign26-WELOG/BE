package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"welog/router"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("경고: .env 파일을 찾을 수 없습니다.")
	}
	r := router.NewRouter()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	fmt.Printf("WELOG API Server is running on port %s\n", port)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}
