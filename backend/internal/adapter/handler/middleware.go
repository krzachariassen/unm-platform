package handler

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"
)

// loggingMiddleware logs each request method, path, and elapsed duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

// makeCORSMiddleware creates a CORS middleware that allows the given origins.
// If origins is empty or contains "*", it allows all origins.
func makeCORSMiddleware(origins []string) func(http.Handler) http.Handler {
	allowAll := false
	if len(origins) == 0 {
		allowAll = true
	} else {
		for _, o := range origins {
			if o == "*" {
				allowAll = true
				break
			}
		}
	}

	// Build a fast lookup set for the allowed origins list.
	allowedSet := make(map[string]bool, len(origins))
	for _, o := range origins {
		allowedSet[o] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if allowAll {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			} else {
				origin := r.Header.Get("Origin")
				if allowedSet[origin] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
				}
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Replace-Model")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// corsMiddleware adds permissive CORS headers and handles preflight requests.
// Kept for backward compatibility — delegates to makeCORSMiddleware with "*".
func corsMiddleware(next http.Handler) http.Handler {
	return makeCORSMiddleware([]string{"*"})(next)
}

// chain wraps h with middleware in order (outermost first).
func chain(h http.Handler, middleware ...func(http.Handler) http.Handler) http.Handler {
	for i := len(middleware) - 1; i >= 0; i-- {
		h = middleware[i](h)
	}
	return h
}

// recoveryMiddleware catches panics in handlers, logs the stack trace, and returns HTTP 500.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("PANIC: %v\n%s", rec, debug.Stack())
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
