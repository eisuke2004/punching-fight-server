package models

import "database/sql"

type Player struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Password  string  `json:"password"`
	RoomID    string  `json:"roomId"`
	PositionX float32 `json:"positionX"`
	PositionY float32 `json:"positionY"`
	PositionZ float32 `json:"positionZ"`
}

func InsertPlayer(db *sql.DB, player Player) error {
	_, err := db.Exec(
		"INSERT INTO players(id,name,password)VALUES(?,?,?)",
		player.ID,
		player.Name,
		player.Password)
	return err
}

func GetPlayer(db *sql.DB, name string) (*Player, error) {
	row := db.QueryRow("SELECT id, name, password FROM players WHERE name = ?", name)
	player := &Player{}

	if err := row.Scan(&player.ID, &player.Name, &player.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return player, nil
}
