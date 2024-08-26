package middleware

import (
	"log"
	"net/http"

	"github.com/AxMdv/go-gophermart/internal/service/auth"
)

const cookieName = "login"

func ValidateUserMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		valid, err := auth.ValidateCookie(cookie)
		if !valid || err != nil {
			// кука не валидна
			log.Println("cookie is not valid or error during validation of cookie", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// c айди все ок - передаём в контексте реквеста айди
		id := auth.GetIDFromCookie(cookie.Value)
		cr := auth.SetUUIDToRequestContext(r, id)
		h.ServeHTTP(w, cr)
	}
}
