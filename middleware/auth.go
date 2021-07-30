package middleware

import (
	"net/http"
	app "newProject/app"
)

const USERKEY app.USER = "USER"
const USERVALUE app.USER = "TEST"
const KEY app.ADMINKEY = "ADMIN"

var ADMIN app.ADMIN

func Auth(w http.ResponseWriter, r *http.Request) (*http.Request, bool, []byte, int) {
	var admin app.ADMIN
	admin.Name = "JIM"
	r = app.SetContextValue(r, KEY, admin)
	return r, true, []byte(`{"error":"Unauthorized action1"}`), 401
}

func Test2(w http.ResponseWriter, r *http.Request) (*http.Request, bool, []byte, int) {

	_admin := app.GetContextValue(r, app.ADMINKEY(KEY))
	if _admin != nil && _admin.(app.ADMIN).Name == "JIM" {
		admin := _admin.(app.ADMIN)
		admin.Id = "1"
		r = app.SetContextValue(r, KEY, admin)
		return r, true, []byte(`{"message":"Authorized"}`), 201
	}
	return r, false, []byte(`{"error":"Unauthorized action2"}`), 401
}
