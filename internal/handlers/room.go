package handlers

import (
	"api/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type CreateResponse struct {
	Roomid  string `json:"roomId"`
	Roomcnt int    `json:"roomCnt"`
}

type JoinRequest struct {
	Roomid string `json:"roomId"`
}

type JoinResponse struct {
	Roomid     string `json:"roomId"`
	Playername string `json:"playerName"`
}

type GetResponse struct {
	Rooms []models.Room `json:"rooms"`
}

func extractClaimsFromToken(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fmt.Errorf("missing or invalid authorization header")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid claims structure")
	}

	return claims, nil
}

func Create(w http.ResponseWriter, r *http.Request) {
	claims, err := extractClaimsFromToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	log.Printf("Room created by player ID: %s, name: %s", claims.PlayerId, claims.PlayerName)

	roomid := uuid.NewString()

	room := models.Room{
		ID: roomid,
	}

	if err := models.InsertRoom(db, room); err != nil {
		log.Println("Insert error:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roomplayer := models.RoomPlayer{
		RoomID:   room.ID,
		PlayerID: claims.PlayerId,
	}

	if err := models.InsertRoomPlayer(db, roomplayer); err != nil {
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	cnt, err := models.GetRoomCount(db)
	if err != nil {
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(CreateResponse{
		Roomid:  roomid,
		Roomcnt: cnt,
	})
}

func Join(w http.ResponseWriter, r *http.Request) {
	claims, err := extractClaimsFromToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var req JoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	log.Printf("Room join by player ID: %s, name: %s", claims.PlayerId, claims.PlayerName)

	exists, err := models.RoomExists(db, req.Roomid)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if !exists {
		http.Error(w, "Room not found", http.StatusGone)
		return
	}

	count, err := models.GetRoomPlayerCount(db, req.Roomid)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if count >= 2 {
		http.Error(w, "Room is full", http.StatusLocked)
		return
	}

	roomplayer := models.RoomPlayer{
		RoomID:   req.Roomid,
		PlayerID: claims.PlayerId,
	}

	playerId, err := models.GetPlayerId(db, req.Roomid)
	if err != nil {
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	playerName, err := models.GetPlayerName(db, playerId)
	if err != nil {
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	if err := models.InsertRoomPlayer(db, roomplayer); err != nil {
		log.Println("InsertRoomPlayer error:", err)
		http.Error(w, "Failed to join room", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(JoinResponse{
		Roomid:     req.Roomid,
		Playername: playerName,
	})
}

func Get(w http.ResponseWriter, r *http.Request) {
	_, err := extractClaimsFromToken(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	rooms, err := models.GetActiveRooms(db)
	if err != nil {
		http.Error(w, "Failed to fetch rooms", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetResponse{Rooms: rooms})
}
