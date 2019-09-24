package handler

import (
	dblayer "filestore-server/db"
	"net/http"
)

func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			userName := r.Form.Get("username")
			token := r.Form.Get("token")

			if len(userName) < 3 || !dblayer.IsTokenValid(token) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			h(w, r)
		})
}
