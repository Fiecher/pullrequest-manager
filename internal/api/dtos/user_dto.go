package dtos

type User struct {
	IsActive bool   `json:"is_active"`
	TeamName string `json:"team_name"`
	UserId   string `json:"user_id"`
	Username string `json:"username"`
}
