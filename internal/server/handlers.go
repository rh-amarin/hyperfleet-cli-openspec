package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	"github.com/rh-amarin/hyperfleet-cli/internal/api"
	"github.com/rh-amarin/hyperfleet-cli/internal/resource"
)

// ClusterEntry wraps a Cluster with its adapter statuses for the dashboard list.
type ClusterEntry struct {
	resource.Cluster
	AdapterStatuses []resource.AdapterStatus `json:"adapter_statuses"`
}

// ClustersResponse is the shape returned by GET /api/clusters.
type ClustersResponse struct {
	Items []ClusterEntry `json:"items"`
	Kind  string         `json:"kind"`
	Page  int32          `json:"page"`
	Size  int32          `json:"size"`
	Total int32          `json:"total"`
}

// handleIndex serves the embedded dashboard HTML.
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(s.indexHTML)
}

// handleClusters fetches the cluster list, concurrently fetches adapter statuses
// for each cluster (max 5 goroutines), merges them, and returns the combined JSON.
func (s *Server) handleClusters(w http.ResponseWriter, r *http.Request) {
	list, err := api.Get[resource.ListResponse[resource.Cluster]](r.Context(), s.client, "clusters")
	if err != nil {
		s.writeError(w, err)
		return
	}

	entries := make([]ClusterEntry, len(list.Items))
	for i, c := range list.Items {
		entries[i].Cluster = c
	}

	// Concurrently fetch statuses with a semaphore of 5.
	sem := make(chan struct{}, 5)
	var mu sync.Mutex
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	for i := range entries {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			statuses, err := api.Get[resource.ListResponse[resource.AdapterStatus]](
				ctx, s.client, "clusters/"+entries[idx].ID+"/statuses",
			)
			if err == nil {
				mu.Lock()
				entries[idx].AdapterStatuses = statuses.Items
				mu.Unlock()
			}
		}(i)
	}
	wg.Wait()

	resp := ClustersResponse{
		Items: entries,
		Kind:  list.Kind,
		Page:  list.Page,
		Size:  list.Size,
		Total: list.Total,
	}
	writeJSON(w, http.StatusOK, resp)
}

// handleCluster proxies GET /api/clusters/{id} → clusters/{id}.
func (s *Server) handleCluster(w http.ResponseWriter, r *http.Request, id string) {
	s.proxyJSON(w, r, "clusters/"+id)
}

// handleClusterStatuses proxies GET /api/clusters/{id}/statuses → clusters/{id}/statuses.
func (s *Server) handleClusterStatuses(w http.ResponseWriter, r *http.Request, id string) {
	s.proxyJSON(w, r, "clusters/"+id+"/statuses")
}

// handleNodePools proxies GET /api/clusters/{id}/nodepools → clusters/{id}/nodepools.
func (s *Server) handleNodePools(w http.ResponseWriter, r *http.Request, id string) {
	s.proxyJSON(w, r, "clusters/"+id+"/nodepools")
}

// handleNodePool proxies GET /api/clusters/{id}/nodepools/{npid} → clusters/{id}/nodepools/{npid}.
func (s *Server) handleNodePool(w http.ResponseWriter, r *http.Request, clusterID, npID string) {
	s.proxyJSON(w, r, "clusters/"+clusterID+"/nodepools/"+npID)
}

// handleNodePoolStatuses proxies GET /api/clusters/{id}/nodepools/{npid}/statuses.
func (s *Server) handleNodePoolStatuses(w http.ResponseWriter, r *http.Request, clusterID, npID string) {
	s.proxyJSON(w, r, "clusters/"+clusterID+"/nodepools/"+npID+"/statuses")
}

// handlePostClusterStatuses proxies POST /api/clusters/{id}/statuses → clusters/{id}/statuses.
func (s *Server) handlePostClusterStatuses(w http.ResponseWriter, r *http.Request, id string) {
	s.proxyPOST(w, r, "clusters/"+id+"/statuses")
}

// handlePostNodePoolStatuses proxies POST /api/clusters/{id}/nodepools/{npid}/statuses.
func (s *Server) handlePostNodePoolStatuses(w http.ResponseWriter, r *http.Request, clusterID, npID string) {
	s.proxyPOST(w, r, "clusters/"+clusterID+"/nodepools/"+npID+"/statuses")
}

// proxyPOST reads the request body, forwards it verbatim to the upstream HyperFleet
// API via POST, and writes the raw response back to w.
func (s *Server) proxyPOST(w http.ResponseWriter, r *http.Request, apiPath string) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "reading request body: " + err.Error()})
		return
	}
	raw, err := api.Post[json.RawMessage](r.Context(), s.client, apiPath, json.RawMessage(body))
	if err != nil {
		s.writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(raw)
}

// proxyJSON fetches apiPath from the upstream HyperFleet API and writes the raw
// JSON response to w, preserving the upstream status code on errors.
func (s *Server) proxyJSON(w http.ResponseWriter, r *http.Request, apiPath string) {
	raw, err := api.Get[json.RawMessage](r.Context(), s.client, apiPath)
	if err != nil {
		s.writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(raw)
}

// writeError writes an upstream APIError (preserving status code) or a 502 for
// unexpected errors, both as JSON.
func (s *Server) writeError(w http.ResponseWriter, err error) {
	var apiErr *api.APIError
	if errors.As(err, &apiErr) {
		status := apiErr.Status
		if status == 0 {
			status = http.StatusBadGateway
		}
		writeJSON(w, status, apiErr)
		return
	}
	writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
}

// writeJSON encodes v as JSON and writes it with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
