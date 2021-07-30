package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	postgres "newProject/db"
	"newProject/models"
)

func HomePage(w http.ResponseWriter, r *http.Request) {
	_admin := GetContextValue(r, ADMINKEY("ADMIN"))
	if _admin != nil {
		admin := _admin.(ADMIN)
		j, _ := json.Marshal(admin)
		Response(r, w, j, 200)
		return
	}

}
func HomePost(w http.ResponseWriter, r *http.Request) {
	setHeader(w, "Content-Type", "application/json")
	var users []models.User
	bodyUser := new(models.User)
	RequestBody(r, bodyUser)
	db := postgres.PSQLConnect()
	if len(bodyUser.Email) < 1 {
		Response(r, w, []byte(`{"message":"No user matches your request"}`), 422)
		return
	}
	sql := "SELECT member_id,email,name,last_name,uuid,password FROM members WHERE email = $1 LIMIT 1"
	rows, err := db.Query(sql, bodyUser.Email)
	if err != nil {
		Response(r, w, []byte(`{"message":"No user matches your request"}`), 422)
		return
	}
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.Member_id, &user.Email, &user.Name, &user.Last_name, &user.Uuid, &user.Password)
		if err != nil {
			Response(r, w, []byte("Something went wrong"), 422)
			return
		}
		users = append(users, user)
	}
	if len(users) < 1 {
		Response(r, w, []byte(`{"message":"No user matches your request"}`), 422)
		return
	}
	JsonResponse(r, w, users[0], 201)
	defer rows.Close()
	defer db.Close()
}

func AboutPage(w http.ResponseWriter, r *http.Request) {
	names := []string{"About", "John"}
	j, err := json.Marshal(names)
	if err != nil {
		newErr := fmt.Sprintf("%s", err)
		w.WriteHeader(422)
		w.Write([]byte(newErr))
		return
	}
	w.WriteHeader(200)
	w.Write(j)
}
