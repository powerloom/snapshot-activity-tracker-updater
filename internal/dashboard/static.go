package dashboard

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
)

// staticFiles holds the embedded frontend files
// Will be populated by go:embed directive in main.go
var staticFiles embed.FS

// StaticServer serves the embedded frontend files
type StaticServer struct {
	fs http.FileSystem
}

// NewStaticServer creates a new static file server from embedded files
func NewStaticServer(efs embed.FS) *StaticServer {
	// Create a sub filesystem that strips the "dist" directory prefix
	sub, err := fs.Sub(efs, "frontend/dist")
	if err != nil {
		// If dist directory doesn't exist, use the root
		sub, _ = fs.Sub(efs, ".")
	}

	return &StaticServer{
		fs: http.FS(sub),
	}
}

// ServeHTTP implements http.Handler for serving static files with SPA fallback
func (s *StaticServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Try to serve the requested file
	upath := r.URL.Path
	if upath == "/" || upath == "" {
		upath = "/index.html"
	}

	// Open the file
	f, err := s.fs.Open(upath[1:]) // Remove leading slash
	if err != nil {
		// File not found, fall back to index.html for SPA routing
		http.FileServer(s.fs).ServeHTTP(w, r)
		return
	}
	defer f.Close()

	// Get file info to check if it's a directory
	stat, err := f.Stat()
	if err != nil {
		// Error getting file info, serve index.html
		s.serveIndexHTML(w, r)
		return
	}

	// If it's a directory, serve index.html
	if stat.IsDir() {
		s.serveIndexHTML(w, r)
		return
	}

	// Serve the file
	http.FileServer(s.fs).ServeHTTP(w, r)
}

// serveIndexHTML serves the index.html file for SPA routing
func (s *StaticServer) serveIndexHTML(w http.ResponseWriter, r *http.Request) {
	// Use a custom request to serve index.html
	indexReq := r.Clone(r.Context())
	indexReq.URL.Path = "/index.html"
	http.FileServer(s.fs).ServeHTTP(w, indexReq)
}

// WithStaticFiles sets the embedded static files
func WithStaticFiles(efs embed.FS) func(*Server) {
	return func(s *Server) {
		staticServer := NewStaticServer(efs)
		s.SetIndexHandler(staticServer)
	}
}

// FileServer creates a simple file server for development (non-embedded)
func FileServer(dir string) http.Handler {
	return http.FileServer(http.Dir(dir))
}

// SpaHandler wraps a file server to handle SPA routing
type SpaHandler struct {
	dir string
}

// NewSpaHandler creates a new SPA handler
func NewSpaHandler(dir string) *SpaHandler {
	return &SpaHandler{
		dir: dir,
	}
}

// ServeHTTP implements http.Handler with SPA fallback
func (h *SpaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Clean the URL path
	upath := path.Clean(r.URL.Path)

	// Try to open the file
	fs := http.Dir(h.dir)
	f, err := fs.Open(upath)
	if err != nil {
		// File not found, serve index.html
		http.ServeFile(w, r, path.Join(h.dir, "index.html"))
		return
	}
	defer f.Close()

	// Get file info
	stat, err := f.Stat()
	if err != nil || stat.IsDir() {
		// Error or directory, serve index.html
		http.ServeFile(w, r, path.Join(h.dir, "index.html"))
		return
	}

	// Serve the file using http.FileServer
	http.FileServer(fs).ServeHTTP(w, r)
}
