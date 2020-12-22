package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
)

type HasAnyRoleFunc func(ctx context.Context, roles ...string) bool

func CheckRole(hasAnyRoleFunc HasAnyRoleFunc, roles ...string) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if !hasAnyRoleFunc(request.Context(), roles...) {
				http.Error(writer, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
			handler.ServeHTTP(writer, request)
		})
	}
}

// // Middleware ...
// type Middleware struct{
// 	secService *security.SecService
// }

// // NewMiddleware ...
// func NewMiddleware(secService *security.SecService) *Middleware{
// 	return &Middleware{secService: secService}
// }

var ErrNoAuthentication = errors.New("no authentication")

var authenticationContextKey = &contextKey{"authentication context"}

type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return c.name
}

type IDFunc func(ctx context.Context, token string) (int64, error)

//Authenticate ...
func Authenticate(idFunc IDFunc) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			token := request.Header.Get("Authorization")

			log.Print("мы тут тоже, в бейсике")
			id, err := idFunc(request.Context(), token)
			if err != nil {
				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(request.Context(), authenticationContextKey, id)
			request = request.WithContext(ctx)

			handler.ServeHTTP(writer, request)
		})
	}
}

// Authentication ...
func Authentication(ctx context.Context) (int64, error) {
	if value, ok := ctx.Value(authenticationContextKey).(int64); ok {
		return value, nil
	}
	return 0, ErrNoAuthentication
}

// // Basic ...
// func (mw *Middleware) Basic(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
// 		login, password, ok := request.BasicAuth()
// 		log.Println(login)
// 		log.Println(password)
// 		log.Println(ok)


// 		result := mw.secService.Auth(request.Context(), login, password)
// 		if result != true {
// 			http.Error(writer, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
// 			return
// 		}


// 		next.ServeHTTP(writer, request)
// 	})
// }




