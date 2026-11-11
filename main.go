package main

import (
	"api/internal/handlers"
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq" // PostgreSQLドライバ
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors" // ★ 1. CORSパッケージをインポート
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .envファイルが見つかりません")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB接続失敗: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("DB疎通確認失敗: %v", err)
	}

	initTables(db)
	handlers.InitDB(db)

	r := mux.NewRouter()

	r.HandleFunc("/register", handlers.Register).Methods("POST", "OPTIONS")
	r.HandleFunc("/login", handlers.Login).Methods("POST", "OPTIONS")
	r.HandleFunc("/createroom", handlers.Create).Methods("POST", "OPTIONS")
	r.HandleFunc("/joinroom", handlers.Join).Methods("POST", "OPTIONS")
	r.HandleFunc("/getroom", handlers.Get).Methods("GET", "OPTIONS")
	r.HandleFunc("/ws", handlers.HandleConnections)

	// ★ 2. CORSのミドルウェア設定を追加
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // すべてのサイト（unityroom含む）からのアクセスを許可
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	// ★ 3. ルーター(r)をCORSでラップしてハンドラーを作る
	handler := c.Handler(r)

	// ★ 4. ListenAndServeには "r" ではなく "handler" を渡す
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", handler))
}

func initTables(db *sql.DB) {
    // （既存のテーブル作成コードそのまま）
    query := `
    CREATE TABLE IF NOT EXISTS players (
        id VARCHAR(255) PRIMARY KEY,
        name VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(255) NOT NULL,
        room_id VARCHAR(255),
        position_x REAL DEFAULT 0,
        position_y REAL DEFAULT 0,
        position_z REAL DEFAULT 0
    );
    CREATE TABLE IF NOT EXISTS rooms (
        id VARCHAR(255) PRIMARY KEY
    );
    CREATE TABLE IF NOT EXISTS room_players (
        room_id VARCHAR(255),
        player_id VARCHAR(255),
        joined_at TIMESTAMP
    );
    `
    _, err := db.Exec(query)
    if err != nil {
        log.Fatalf("テーブルの自動作成に失敗しました: %v", err)
    }
    log.Println("データベースのテーブル確認・作成が完了しました！")
}