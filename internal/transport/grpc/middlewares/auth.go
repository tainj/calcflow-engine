package middlewares

import (
	"net/http"

	"github.com/tainj/distributed_calculator2/internal/auth"
)

func AuthMiddleware(jwtService auth.JWTService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// List of paths where authentication is not required
			if r.URL.Path == "/v1/register" || r.URL.Path == "/v1/login" {
				next.ServeHTTP(w, r)
				return
			}

			// Others - check
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				http.Error(w, "invalid authorization format", http.StatusUnauthorized)
				return
			}

			tokenStr := authHeader[7:]
			claims, err := jwtService.ParseToken(tokenStr)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := auth.WithUserID(r.Context(), claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
