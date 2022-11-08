package middleware

import (
	"net/http"
)

func CheckBasicAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		ck, err := r.Cookie("sessionid")
		if err != nil {
			w.WriteHeader(401)
		}

		next.ServeHTTP(w, r)

	}

	return http.HandlerFunc(fn)

}
