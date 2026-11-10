package handlers

import (
	"api/internal/models"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

var jwtKey []byte
var db *sql.DB

type Claims struct {
	PlayerId   string `json:"playerId"`
	PlayerName string `json:"playerName"`
	jwt.RegisteredClaims
}
type RegisterRequest struct {
	Playername string `json:"playerName"`
	Password   string `json:"password"`
}

type RegisterResponse struct {
	Playername string `json:"playerName"`
	Playerid   string `json:"playerId"`
}

type LoginRequest struct {
	Playername string `json:"playerName"`
	Password   string `json:"password"`
}

type LoginResponse struct {
	Token      string `json:"token"`
	Playername string `json:"playerName"`
	Playerid   string `json:"playerId"`
}

func InitDB(d *sql.DB) {
	db = d
	jwtKey = []byte(os.Getenv("JWT_SECRET"))
}

func Register(w http.ResponseWriter, r *http.Request) {
	log.Println("Register handler hit")
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	name := req.Playername

	if name == "" {
		http.Error(w, "Playername is required", http.StatusUnprocessableEntity)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusUnprocessableEntity)
		return
	}

	playerid := uuid.NewString()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Filed to hash password", http.StatusInternalServerError)
		return
	}

	player := models.Player{
		ID:       playerid,
		Name:     name,
		Password: string(hashedPassword),
	}

	log.Println("Inserting player:", player.ID, player.Name)
	if err := models.InsertPlayer(db, player); err != nil {
		log.Println("Insert error:", err)
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RegisterResponse{
		Playername: name,
		Playerid:   playerid,
	})
}

func Login(w http.ResponseWriter, r *http.Request) {
	log.Println("Login handler called")
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if req.Playername == "" {
		http.Error(w, "Playername is required", http.StatusUnprocessableEntity)
		return
	}

	if req.Password == "" {
		http.Error(w, "Password is required", http.StatusUnprocessableEntity)
		return
	}

	player, err := models.GetPlayer(db, req.Playername)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(player.Password), []byte(req.Password))
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		PlayerId:   player.ID,
		PlayerName: player.Name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		http.Error(w, "Internal Server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token:      tokenString,
		Playername: claims.PlayerName,
		Playerid:   claims.PlayerId,
	})
}
