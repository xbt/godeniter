package middleware

import (
	"net/http"
)

// AccessControlMiddleware is a middleware for controlling access to routes
type AccessControlMiddleware struct {
	// This could be a map of routes to roles or any other structure
	allowedPaths map[string]bool
}

// NewAccessControlMiddleware creates a new AccessControlMiddleware
func NewAccessControlMiddleware() *AccessControlMiddleware {
	return &AccessControlMiddleware{
		allowedPaths: map[string]bool{
			"/public": true,  // Example of a public route
			"/admin":  false, // Example of a protected route
		},
	}
}

// Handle checks if the request is allowed to proceed
func (a *AccessControlMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the path is allowed
		if allowed, ok := a.allowedPaths[r.URL.Path]; ok && !allowed {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
