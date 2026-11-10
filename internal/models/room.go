package models

import (
	"database/sql"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Room struct {
	ID string `json:"id"`
}

type RoomPlayer struct {
	RoomID   string `json:"room_id"`
	PlayerID string `json:"player_id"`
}

type Ranking struct {
	Score    int    `json:"score"`
	PlayerID string `json:"player_id"`
}

type WsRoom struct {
	Players map[*websocket.Conn]*Player
	Mutex   sync.Mutex
}

func (room *WsRoom) AddPlayer(conn *websocket.Conn, player *Player) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()
	room.Players[conn] = player
}

func (room *WsRoom) RemovePlayer(conn *websocket.Conn) {
	room.Mutex.Lock()
	defer room.Mutex.Unlock()
	delete(room.Players, conn)
}

func (room *WsRoom) BroadcastToAll(msg interface{}, exclude *websocket.Conn) {
	messageBytes, err := json.Marshal(msg)
	if err != nil {
		log.Printf("JSON Marshal error: %v", err)
		return
	}

	room.Mutex.Lock()
	var toRemove []*websocket.Conn

	for conn := range room.Players {
		if conn == exclude {
			continue // 自分には送らない
		}
		if err := conn.WriteMessage(websocket.TextMessage, messageBytes); err != nil {
			log.Printf("送信エラー、接続削除予定: %v", err)
			toRemove = append(toRemove, conn)
		}
	}

	for _, conn := range toRemove {
		delete(room.Players, conn)
	}
	room.Mutex.Unlock()

	for _, conn := range toRemove {
		conn.Close()
	}
}

func InsertRoom(db *sql.DB, room Room) error {
	_, err := db.Exec(
		"INSERT INTO rooms(id) VALUES($1)",
		room.ID)
	return err
}

func InsertRoomPlayer(db *sql.DB, roomPlayers RoomPlayer) error {
	_, err := db.Exec(
		"INSERT INTO room_players(room_id, player_id, joined_at) VALUES($1, $2, $3)",
		roomPlayers.RoomID, roomPlayers.PlayerID, time.Now())
	return err
}

func RoomExists(db *sql.DB, roomID string) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM rooms WHERE id = $1)", roomID).Scan(&exists)
	return exists, err
}

func GetRoomPlayerCount(db *sql.DB, roomID string) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM room_players WHERE room_id = $1", roomID).Scan(&count)
	return count, err
}

func GetActiveRooms(db *sql.DB) ([]Room, error) {
	rows, err := db.Query(`
		SELECT r.id
		FROM rooms r
		LEFT JOIN room_players rp ON r.id = rp.room_id
		GROUP BY r.id
		HAVING COUNT(rp.player_id) < 2
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []Room
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.ID); err != nil {
			return nil, err
		}
		rooms = append(rooms, room)
	}

	return rooms, nil
}

func GetPlayerId(db *sql.DB, roomID string) (string, error) {
	var playerId string
	err := db.QueryRow("SELECT player_id FROM room_players WHERE room_id = $1", roomID).Scan(&playerId)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return playerId, err
}

func GetPlayerName(db *sql.DB, playerId string) (string, error) {
	var playerName string
	err := db.QueryRow("SELECT name FROM players WHERE id = $1", playerId).Scan(&playerName)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return playerName, err
}

func GetRoomCount(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM rooms").Scan(&count)
	return count, err
}

func DeleteRoomPlayer(db *sql.DB, roomID, playerID string) error {
	_, err := db.Exec("DELETE FROM room_players WHERE room_id = $1 AND player_id = $2", roomID, playerID)
	return err
}

func DeleteRoom(db *sql.DB, roomID string) error {
	_, err := db.Exec("DELETE FROM rooms WHERE id = $1", roomID)
	return err
}