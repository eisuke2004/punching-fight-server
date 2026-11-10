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
		"INSERT INTO players(id, name, password) VALUES($1, $2, $3)",
		player.ID,
		player.Name,
		player.Password)
	return err
}

func GetPlayer(db *sql.DB, name string) (*Player, error) {
	row := db.QueryRow("SELECT id, name, password FROM players WHERE name = $1", name)
	player := &Player{}

	if err := row.Scan(&player.ID, &player.Name, &player.Password); err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return player, nil
}
```[cite: 11]

#### 2. `internal/models/room.go` も同様に修正
`room_3.go` の中にあるすべての `?` も、順番に応じて `$1`, `$2`, `$3` に書き換える必要があります（例: `WHERE id = ?` なら `WHERE id = $1` に変更）[cite: 12]。

これらを修正してGitHubにプッシュし、Renderのデプロイが完了したあとにUnityから再度「登録」を試してみてください。エラーが解消され、無事に登録とログインができるようになります！