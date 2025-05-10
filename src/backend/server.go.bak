package backend

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// SearchRequest defines the structure of a search request.
type SearchRequest struct {
	Element string `json:"element"`
	Mode    string `json:"mode"`
}

// SearchResponse defines the structure of a search response.
type SearchResponse struct {
	Path        []Step `json:"path"`
	Found       bool   `json:"found"`
	Steps       int    `json:"steps"`
}

// enableCORS adds CORS headers to HTTP responses.
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

// searchHandler handles search requests for elements.
func searchHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request: %s %s", r.Method, r.URL.String())

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	element := strings.ToLower(r.URL.Query().Get("element"))
	mode := strings.ToLower(r.URL.Query().Get("mode"))
	log.Printf("Searching for element: %s with mode: %s", element, mode)

	if element == "" {
		http.Error(w, "Element parameter is required", http.StatusBadRequest)
		return
	}

	graph, err := LoadGraph("combinations.json")
	if err != nil {
		log.Printf("Error loading graph: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Printf("Graph loaded with %d elements", len(graph))

	timeout := 60 * time.Second
	start := []string{"air", "earth", "fire", "water"}

	var steps []Step
	var found bool
	var nodesVisited int

	if mode == "dfs" {
		steps, found, nodesVisited = DFSRecipe(graph, start, element, timeout)
	} else {
		steps, found, nodesVisited = BFSRecipe(graph, start, element, timeout)
	}

	response := SearchResponse{
		Path:  steps,
		Found: found,
		Steps: nodesVisited,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// LoadGraph loads the graph data from a JSON file.
func LoadGraph(path string) (Graph, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var graph Graph
	err = json.Unmarshal(file, &graph)
	return graph, err
}

// main starts the HTTP server.
func main() {
	log.Println("Starting server on :5000")
	http.HandleFunc("/search", enableCORS(searchHandler))
	log.Fatal(http.ListenAndServe(":5000", nil))
}