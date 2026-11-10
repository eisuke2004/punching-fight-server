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


initTables(db)
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

// データベースに接続した直後の位置にこれを追加する
func initTables(db *sql.DB) {
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