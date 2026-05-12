package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"welog/internal/router"
	"welog/pkg/database"
	"welog/pkg/middleware"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("경고: .env 파일을 찾을 수 없습니다.")
	}
	dsn := os.Getenv("DSN")
	db := database.ConnectDB(dsn)
	jwtSecret := os.Getenv("JWT_SECRET")
	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	r, cleanup := router.NewRouter(db, jwtSecret, googleClientID)
	defer cleanup()

	r = middleware.CorsMiddleware(r)

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
