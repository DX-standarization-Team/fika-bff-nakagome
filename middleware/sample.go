package middleware

import (
	"context"
	"net/http"
)

func authenticationMiddlewareSample(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for the presence of an "Authorization" header
		token := r.Header.Get("Authorization")

		// Replace this with your actual authentication logic
		if token != "your_secret_token" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
		}

		// Store the token in the request context
		ctx := context.WithValue(r.Context(), "token", token)

		// Authentication passed; call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}