// Package server provides an HTTP server that serves the HyperFleet browser
// dashboard and proxies read requests to the HyperFleet API.
package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
)

// Server serves the HyperFleet browser UI and proxies API reads.
type Server struct {
	client    *api.Client
	port      int
	indexHTML []byte
}

// New creates a Server. indexHTML is the embedded dashboard HTML served at GET /.
func New(client *api.Client, port int, indexHTML []byte) *Server {
	return &Server{client: client, port: port, indexHTML: indexHTML}
}

// Listen starts the HTTP server and blocks until it exits.
func (s *Server) Listen() error {
	addr := fmt.Sprintf(":%d", s.port)
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.route)
	fmt.Printf("Serving HyperFleet UI at http://localhost:%d\n", s.port)
	return http.ListenAndServe(addr, mux)
}

// route dispatches all requests.
// Status paths (…/statuses) accept GET and POST; all other paths are GET-only.
// Path patterns (after trimming leading/trailing slashes):
//
//	""                                              → index.html            (GET)
//	api/clusters                                    → handleClusters        (GET)
//	api/clusters/{id}                               → handleCluster         (GET)
//	api/clusters/{id}/statuses                      → handleClusterStatuses (GET|POST)
//	api/clusters/{id}/nodepools                     → handleNodePools       (GET)
//	api/clusters/{id}/nodepools/{npid}              → handleNodePool        (GET)
//	api/clusters/{id}/nodepools/{npid}/statuses     → handleNodePoolStatuses(GET|POST)
func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(r.URL.Path, "/")

	// Root → dashboard HTML (GET only)
	if path == "" {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		s.handleIndex(w, r)
		return
	}

	parts := strings.Split(path, "/")

	switch {
	// /api/clusters  (GET only)
	case len(parts) == 2 && parts[0] == "api" && parts[1] == "clusters":
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		s.handleClusters(w, r)

	// /api/clusters/{id}  (GET only)
	case len(parts) == 3 && parts[0] == "api" && parts[1] == "clusters":
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		s.handleCluster(w, r, parts[2])

	// /api/clusters/{id}/statuses  (GET|POST)
	// /api/clusters/{id}/nodepools (GET only)
	case len(parts) == 4 && parts[0] == "api" && parts[1] == "clusters":
		switch parts[3] {
		case "statuses":
			switch r.Method {
			case http.MethodGet:
				s.handleClusterStatuses(w, r, parts[2])
			case http.MethodPost:
				s.handlePostClusterStatuses(w, r, parts[2])
			default:
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			}
		case "nodepools":
			if r.Method != http.MethodGet {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			s.handleNodePools(w, r, parts[2])
		default:
			http.NotFound(w, r)
		}

	// /api/clusters/{id}/nodepools/{npid}  (GET only)
	case len(parts) == 5 && parts[0] == "api" && parts[1] == "clusters" && parts[3] == "nodepools":
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		s.handleNodePool(w, r, parts[2], parts[4])

	// /api/clusters/{id}/nodepools/{npid}/statuses  (GET|POST)
	case len(parts) == 6 && parts[0] == "api" && parts[1] == "clusters" && parts[3] == "nodepools" && parts[5] == "statuses":
		switch r.Method {
		case http.MethodGet:
			s.handleNodePoolStatuses(w, r, parts[2], parts[4])
		case http.MethodPost:
			s.handlePostNodePoolStatuses(w, r, parts[2], parts[4])
		default:
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		}

	default:
		http.NotFound(w, r)
	}
}
