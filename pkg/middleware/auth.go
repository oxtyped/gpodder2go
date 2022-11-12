package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	b64 "encoding/base64"
	"log"
	"net/http"
	"strings"
)

func Verify(key string) func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			ck, err := r.Cookie("sessionid")
			if err != nil {
				w.WriteHeader(400)
				log.Println(err)
				return
			}

			decodedStr, err := b64.StdEncoding.DecodeString(ck.Value)
			if err != nil {
				w.WriteHeader(400)
				log.Println(err)
				return
			}

			parsedStr := strings.Split(string(decodedStr), ".")

			signature := parsedStr[0]
			username := parsedStr[1]

			mac := hmac.New(sha256.New, []byte(key))
			mac.Write([]byte(username))

			if !hmac.Equal([]byte(signature), mac.Sum(nil)) {
				w.WriteHeader(401)
				return

			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func Verifier(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Verify(key)(next)
	}
}

// CheckBasicAuth is a middleware that checks the authenticity of the user attempting to access secured endpoints
