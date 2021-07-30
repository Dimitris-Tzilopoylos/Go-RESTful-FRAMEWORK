package models

type User struct {
	Member_id int    `json:"member_id"`
	Name      string `json:"name"`
	Last_name string `json:"last_name"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Uuid      string `json:"uuid"`
}
