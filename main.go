package main

import (
	"api/internal/handlers"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .envファイルが見つかりません")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("DB疎通確認失敗: %v", err)
	}

	handlers.InitDB(db)

	r := mux.NewRouter()

	r.HandleFunc("/register", handlers.Register)
	r.HandleFunc("/login", handlers.Login)

	r.HandleFunc("/createroom", handlers.Create)
	r.HandleFunc("/joinroom", handlers.Join)
	r.HandleFunc("/getroom", handlers.Get)
	r.HandleFunc("/ws", handlers.HandleConnections)

	log.Fatal(http.ListenAndServe("0.0.0.0:8080", r))
}
