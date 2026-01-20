package middleware

import (
	"bytes"
	"strings"
	"errors"
	
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"net/http"
	d2 "github.com/oxtyped/gpodder2go/pkg/data"
)

// VerifyCookie verifies the session cookie and returns username.
// Returns http.ErrNoCookie if the cookie is missing.
func VerifyCookie(key string, r *http.Request) (string, error) {
	ck, err := r.Cookie("sessionid")
	if err != nil {
		return "", err // http.ErrNoCookie when missing
	}

	session, err := base64.StdEncoding.DecodeString(ck.Value)
	if err != nil {
		log.Printf("cookie-auth: invalid encoding: %#v", err)
		return "", err
	}

	i := bytes.LastIndexByte(session, '.')
	if i < 0 {
		return "", errors.New("cookie-auth: invalid cookie format; expected $sig.$user")
	}

	sign := session[:i]
	user := string(session[i+1:]) // last dot => supports dots inside username

	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(user))
	expected := mac.Sum(nil)

	if !hmac.Equal(sign, expected) {
		return "", errors.New("cookie-auth: invalid signature")
	}
	return user, nil
}

// VerifyBasic verifies HTTP Basic auth where credentials are "username:signature_b64".
// signature_b64 should be base64 of the raw HMAC-SHA256 bytes.
func VerifyBasic(key string, r *http.Request, db *d2.SQLite) (string, error) {
	auth := r.Header.Get("Authorization")
	if auth == "" || !strings.HasPrefix(auth, "Basic ") {
		return "", errors.New("no basic auth available")
	}

	credsB, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(auth, "Basic "))
	if err != nil {
		log.Printf("basic-auth: invalid encoding: %#v", err)
		return "", err
	}
	parts := strings.SplitN(string(credsB), ":", 2)
	if len(parts) != 2 {

		return "", errors.New("basic-auth: malformed credentials; expected $user:$pass")
	}
	user := parts[0]
	pass := parts[1]

	if !db.CheckUserPassword(user, pass) {
		return "", errors.New("basic-auth: invalid username or password")
	}

	return user, nil
}


func Verify(key string, noAuth bool, db *d2.SQLite) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		hfn := func(w http.ResponseWriter, r *http.Request) {
			if noAuth {
				next.ServeHTTP(w, r)
				return
			}

			user, err := VerifyCookie(key, r)
			if err != nil {
				user, err = VerifyBasic(key, r, db)
			}
			if err != nil {
				if err == http.ErrNoCookie {
					log.Printf("missing cookie, have you logged in yet: %#v", err)
					w.WriteHeader(401)
					return
				} else {
					w.WriteHeader(400)
					log.Printf("error retrieving cookie: %#v", err)
					return
				}
			}

			_ = user

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(hfn)
	}
}

func Verifier(key string, noAuth bool, db *d2.SQLite) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Verify(key, noAuth, db)(next)
	}
}

// CheckBasicAuth is a middleware that checks the authenticity of the user attempting to access secured endpoints
