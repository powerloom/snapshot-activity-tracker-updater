package dashboard

import (
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
)

// wrapCORS enables browser calls from datamarkets.powerloom.io (and other origins).
// If the edge nginx returns 404 for OPTIONS, add:
//
//	location /api/ {
//	    if ($request_method = OPTIONS) { add_header Access-Control-Allow-Origin $http_origin; ... }
//	    proxy_pass http://dashboard-api:8080;
//	}
//
// corsOriginsFromEnv returns allowed browser origins for dashboard JSON routes.
// DASHBOARD_CORS_ORIGINS: comma-separated list, or "*" (default) for any origin.
// Example: https://datamarkets.powerloom.io,http://localhost:5173
func corsOriginsFromEnv() []string {
	raw := strings.TrimSpace(os.Getenv("DASHBOARD_CORS_ORIGINS"))
	if raw == "" || raw == "*" {
		return []string{"*"}
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func wrapCORS(next http.Handler) http.Handler {
	origins := corsOriginsFromEnv()
	c := cors.New(cors.Options{
		AllowedOrigins:   origins,
		AllowedMethods:   []string{http.MethodGet, http.MethodOptions, http.MethodHead},
		AllowedHeaders:   []string{"Accept", "Content-Type", "Origin"},
		AllowCredentials: false,
		MaxAge:           86400,
	})
	return c.Handler(next)
}
