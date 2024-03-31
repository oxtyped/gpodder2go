package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	b64 "encoding/base64"
	"log"
	"net/http"
)

func Verify(key string, noAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			if noAuth {
				next.ServeHTTP(w, r)
				return
			}

			ck, err := r.Cookie("sessionid")
			if err != nil {
				w.WriteHeader(400)
				log.Println(err)
				return
			}

			session, err := b64.StdEncoding.DecodeString(ck.Value)
			if err != nil {
				w.WriteHeader(400)
				log.Println(err)
				return
			}

			i := bytes.LastIndexByte(session, '.')
			if i < 0 {
				w.WriteHeader(400)
				log.Println("invalid cookie format")
				return
			}

			var (
				sign = session[:i]
				user = session[i+1:] // FIXME: how to handle usernames with a dot '.' ?
			)

			mac := hmac.New(sha256.New, []byte(key))
			mac.Write(user)

			if !hmac.Equal([]byte(sign), mac.Sum(nil)) {
				w.WriteHeader(401)
				return

			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func Verifier(key string, noAuth bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Verify(key, noAuth)(next)
	}
}

// CheckBasicAuth is a middleware that checks the authenticity of the user attempting to access secured endpoints
