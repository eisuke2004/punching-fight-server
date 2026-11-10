package handlers

import (
	"api/internal/models"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/websocket"
)

// 共有データ
var players = make(map[*websocket.Conn]*models.Player) // 接続しているプレイヤーリスト
var rooms = make(map[string]*models.WsRoom)            // ルーム一覧
var mutex = &sync.Mutex{}                              // 共有データのロック
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleConnections(w http.ResponseWriter, r *http.Request) {
	log.Println("Start WS")
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("panic recovered in WS handler: %v", rec)
		}
	}()
	roomID := r.URL.Query().Get("room_id")
	tokenString := r.URL.Query().Get("token")
	log.Printf("WS params room_id='%s' token='%s'", roomID, tokenString)
	if tokenString == "" {
		http.Error(w, "Token missing", http.StatusUnauthorized)
		return
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		http.Error(w, "Invalid claims", http.StatusUnauthorized)
		return
	}

	if roomID == "" {
		http.Error(w, "ルームIDは必須です", http.StatusBadRequest)
		return
	}
	log.Printf("WS request room_id=%s token=%s", roomID, tokenString)

	player := &models.Player{ID: claims.PlayerId, Name: claims.PlayerName, RoomID: roomID}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	log.Println("WS Upgrade success")

	defer func() {
		ws.Close()

		var room *models.WsRoom
		var leavePayload map[string]interface{}

		mutex.Lock()
		player, ok := players[ws]
		if ok {
			delete(players, ws)
			r, exists := rooms[player.RoomID]
			if exists {
				r.RemovePlayer(ws)

				leavePayload = map[string]interface{}{
					"event":      "player_leave",
					"playerId":   player.ID,
					"playerName": player.Name,
				}
				if len(r.Players) == 0 {
					delete(rooms, player.RoomID)
				}
				room = r
				log.Printf("RemovePlayer後のプレイヤー数: %d", len(room.Players))

			}
		}
		mutex.Unlock()

		if room != nil {
			room.BroadcastToAll(leavePayload, nil)
		}

		if ok {
			if len(room.Players) == 0 {
				if err := models.DeleteRoom(db, player.RoomID); err != nil {
					log.Printf("DB削除失敗: %v", err)
				}
			}
			if err := models.DeleteRoomPlayer(db, player.RoomID, player.ID); err != nil {
				log.Printf("DB削除失敗: %v", err)
			}
		}
	}()

	mutex.Lock() // 共有データのロック
	players[ws] = player
	room, exists := rooms[roomID]
	if !exists {
		room = &models.WsRoom{
			Players: make(map[*websocket.Conn]*models.Player),
		}

		rooms[roomID] = room
	}
	room.AddPlayer(ws, player)
	mutex.Unlock()

	room.BroadcastToAll(map[string]interface{}{
		"event":      "player_join",
		"playerId":   player.ID,
		"playerName": player.Name,
	}, ws)

	// イベント処理ループ
	for {
		_, msgBytes, err := ws.ReadMessage()
		if err != nil {
			log.Printf("ReadMessageエラー: %v", err)
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		event, _ := msg["event"].(string)

		switch event {
		case "request_start_game":
			room.BroadcastToAll(map[string]interface{}{
				"event": "start_game",
			}, nil)

		case "action_input":
			room.BroadcastToAll(map[string]interface{}{
				"event":     "action_update",
				"playerId":  player.ID,
				"x":         msg["x"],
				"y":         msg["y"],
				"z":         msg["z"],
				"strikeFlg": msg["strikeFlg"],
			}, ws)
		}
	}
}
