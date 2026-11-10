package main

import (
	"api/internal/handlers"
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // PostgreSQLドライバに変更
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .envファイルが見つかりません")
	}

	// Renderの環境変数「DATABASE_URL」を直接取得して使用する
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// ドライバ名を "mysql" から "postgres" に変更
	db, err := sql.Open("postgres", dsn)
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