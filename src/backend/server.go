package backend
import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)
type SearchRequest struct {
	Element string `json:"element"`
	Mode    string `json:"mode"`
}
type SearchResponse struct {
	Path  []Step `json:"path"`
	Found bool   `json:"found"`
	Steps int    `json:"steps"`
}
type Step struct {
	Ingredients [2]string `json:"ingredients"`
	Result      string    `json:"result"`
}
type State struct {
	Available map[string]bool
	Path      []Step
	Depth     int
	ID        string
	ParentID  string
}
type Graph map[string][][]string
type InverseGraph map[string]map[string][]string
func LoadGraph(path string) (Graph, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var graph Graph
	err = json.Unmarshal(file, &graph)
	return graph, err
}

func CreateInverseGraph(graph Graph) InverseGraph {
	inverse := make(InverseGraph)
	for result, combos := range graph {
		for _, combo := range combos {
			a, b := combo[0], combo[1]
			if _, exists := inverse[a]; !exists {
				inverse[a] = make(map[string][]string)
			}
			inverse[a][b] = append(inverse[a][b], result)
			if a != b {
				if _, exists := inverse[b]; !exists {
					inverse[b] = make(map[string][]string)
				}
				inverse[b][a] = append(inverse[b][a], result)
			}
		}
	}
	return inverse
}

func BFSRecipe(graph Graph, start []string, target string, timeout time.Duration) ([]Step, bool, int) {
	log.Printf("BFS: Starting search for target: %s", target)
	inverse := CreateInverseGraph(graph)
	queueSize := 10000 
	forwardQueue := make([]State, 0, queueSize)
	forwardQueue = append(forwardQueue, State{Available: sliceToSet(start), Path: []Step{}, Depth: 0})
	visitedStates := make(map[string]struct{})
	visitedStates[stateHash(sliceToSet(start))] = struct{}{}
	nodesVisited := 0
	startTime := time.Now()
	if forwardQueue[0].Available[target] {
		log.Printf("BFS: Target %s is already available in start elements", target)
		return []Step{}, true, 1
	}
	maxDepth := 7 
	canLeadToTarget := make(map[string]bool)
	for result, combos := range graph {
		if result == target {
			for _, combo := range combos {
				canLeadToTarget[combo[0]] = true
				canLeadToTarget[combo[1]] = true
			}
		}
	}
	intermediateElements := make(map[string]bool)
	for result, combos := range graph {
		for _, combo := range combos {
			if canLeadToTarget[result] {
				intermediateElements[combo[0]] = true
				intermediateElements[combo[1]] = true
			}
		}
	}
	for len(forwardQueue) > 0 {
		if time.Since(startTime) > timeout {
			log.Printf("BFS: Search timed out after visiting %d nodes", nodesVisited)
			return nil, false, nodesVisited
		}
		current := forwardQueue[0]
		forwardQueue = forwardQueue[1:]
		nodesVisited++
		if nodesVisited%1000 == 0 {
			log.Printf("BFS: Visited %d nodes, queue size: %d, current depth: %d", nodesVisited, len(forwardQueue), current.Depth)
		}
		if current.Depth >= maxDepth {
			continue
		}
		available := make([]string, 0, len(current.Available))
		for elem := range current.Available {
			available = append(available, elem)
		}
		sort.Slice(available, func(i, j int) bool {
			iCanLead := canLeadToTarget[available[i]] || intermediateElements[available[i]]
			jCanLead := canLeadToTarget[available[j]] || intermediateElements[available[j]]
			return iCanLead && !jCanLead
		})
		for i := 0; i < len(available); i++ {
			a := available[i]
			if !canLeadToTarget[a] && !intermediateElements[a] && current.Depth > 0 {
				continue
			}
			for j := i; j < len(available); j++ {
				b := available[j]
				if !canLeadToTarget[a] && !intermediateElements[a] && 
				   !canLeadToTarget[b] && !intermediateElements[b] {
					continue
				}
				if _, exists := inverse[a]; !exists {
					continue
				}
				results, exists := inverse[a][b]
				if !exists {
					continue
				}
				for _, result := range results {
					if result == target {
						log.Printf("BFS: Found target %s after visiting %d nodes", target, nodesVisited)
						newPath := make([]Step, len(current.Path))
						copy(newPath, current.Path)
						newPath = append(newPath, Step{
							Ingredients: [2]string{a, b},
							Result:      result,
						})
						return newPath, true, nodesVisited
					}
					if !current.Available[result] {
						newAvailable := copyMap(current.Available)
						newAvailable[result] = true
						stateKey := stateHash(newAvailable)
						if _, visited := visitedStates[stateKey]; !visited {
							visitedStates[stateKey] = struct{}{}
							if len(forwardQueue) < queueSize {
								newPath := make([]Step, len(current.Path))
								copy(newPath, current.Path)
								newPath = append(newPath, Step{
									Ingredients: [2]string{a, b},
									Result:      result,
								})
								forwardQueue = append(forwardQueue, State{
									Available: newAvailable,
									Path:      newPath,
									Depth:     current.Depth + 1,
								})
							}
						}
					}
				}
			}
		}
	}
	log.Printf("BFS: Target %s not found after visiting %d nodes", target, nodesVisited)
	return nil, false, nodesVisited
}

func DFSRecipe(graph Graph, start []string, target string, timeout time.Duration) ([]Step, bool, int) {
	log.Printf("DFS: Starting search for target: %s", target)
	inverse := CreateInverseGraph(graph)
	stack := []State{{Available: sliceToSet(start), Path: []Step{}, Depth: 0}}
	visitedStates := make(map[string]struct{})
	visitedStates[stateHash(sliceToSet(start))] = struct{}{}
	nodesVisited := 0
	startTime := time.Now()
	if stack[0].Available[target] {
		log.Printf("DFS: Target %s is already available in start elements", target)
		return []Step{}, true, 1
	}
	maxDepth := 10
	for len(stack) > 0 {
		if time.Since(startTime) > timeout {
			log.Printf("DFS: Search timed out after visiting %d nodes", nodesVisited)
			return nil, false, nodesVisited
		}
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		nodesVisited++
		if nodesVisited%1000 == 0 {
			log.Printf("DFS: Visited %d nodes, stack size: %d, current depth: %d", nodesVisited, len(stack), current.Depth)
		}
		if current.Depth >= maxDepth {
			continue
		}
		available := keys(current.Available)
		for i := 0; i < len(available); i++ {
			for j := i; j < len(available); j++ {
				a, b := available[i], available[j]
				if _, exists := inverse[a]; !exists {
					continue
				}
				results, exists := inverse[a][b]
				if !exists {
					continue
				}
				for _, result := range results {
					if result == target {
						log.Printf("DFS: Found target %s after visiting %d nodes", target, nodesVisited)
						newPath := make([]Step, len(current.Path))
						copy(newPath, current.Path)
						newPath = append(newPath, Step{
							Ingredients: [2]string{a, b},
							Result:      result,
						})
						return newPath, true, nodesVisited
					}
					if !current.Available[result] {
						newAvailable := copyMap(current.Available)
						newAvailable[result] = true
						stateKey := stateHash(newAvailable)
						if _, visited := visitedStates[stateKey]; !visited {
							visitedStates[stateKey] = struct{}{}
							newPath := make([]Step, len(current.Path))
							copy(newPath, current.Path)
							newPath = append(newPath, Step{
								Ingredients: [2]string{a, b},
								Result:      result,
							})
							stack = append(stack, State{
								Available: newAvailable,
								Path:      newPath,
								Depth:     current.Depth + 1,
							})
						}
					}
				}
			}
		}
	}
	log.Printf("DFS: Target %s not found after visiting %d nodes", target, nodesVisited)
	return nil, false, nodesVisited
}

func stateHash(available map[string]bool) string {
	keys := make([]string, 0, len(available))
	for k := range available {
		keys = append(keys, k)
	}
	return strings.Join(keys, ",")
}

func sliceToSet(slice []string) map[string]bool {
	set := make(map[string]bool)
	for _, s := range slice {
		set[s] = true
	}
	return set
}

func keys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func copyMap(m map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "86400") 
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}

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
	inverse := CreateInverseGraph(graph)
	isBaseElement := false
	for _, base := range []string{"air", "earth", "fire", "water"} {
		if element == base {
			isBaseElement = true
			break
		}
	}
	elementExists := isBaseElement || graph[element] != nil
	if !elementExists {
		for _, combos := range inverse {
			for _, results := range combos {
				for _, result := range results {
					if result == element {
						elementExists = true
						break
					}
				}
				if elementExists {
					break
				}
			}
			if elementExists {
				break
			}
		}
	}
	if !elementExists {
		log.Printf("Element %s not found in graph and cannot be created", element)
		response := SearchResponse{
			Path:  nil,
			Found: false,
			Steps: 0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}
	if isBaseElement {
		log.Printf("Element %s is a base element", element)
	} else if combos, exists := graph[element]; exists {
		log.Printf("Element %s found in graph with %d combinations:", element, len(combos))
		for _, combo := range combos {
			log.Printf("  - %s + %s", combo[0], combo[1])
		}
	} else {
		log.Printf("Element %s can be created from other combinations", element)
	}
	baseElements := []string{"air", "earth", "fire", "water"}
	log.Printf("Starting with base elements: %v", baseElements)
	var steps []Step
	var found bool
	var nodesVisited int
	timeout := 60 * time.Second
	if mode == "dfs" {
		log.Printf("Starting DFS search for element: %s", element)
		steps, found, nodesVisited = DFSRecipe(graph, baseElements, element, timeout)
	} else {
		log.Printf("Starting BFS search for element: %s", element)
		steps, found, nodesVisited = BFSRecipe(graph, baseElements, element, timeout)
	}
	log.Printf("Search completed. Found: %v, Steps: %d", found, nodesVisited)
	if found {
		log.Printf("Path length: %d", len(steps))
		for i, step := range steps {
			log.Printf("Step %d: %s + %s = %s", i+1, step.Ingredients[0], step.Ingredients[1], step.Result)
		}
	}
	response := SearchResponse{
		Path:  steps,
		Found: found,
		Steps: nodesVisited,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type VisualizationNode struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Children []string `json:"children"`
	Parent   string   `json:"parent"`
	Depth    int      `json:"depth"`
}

type VisualizationData struct {
	Nodes []VisualizationNode `json:"nodes"`
	Edges []struct {
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"edges"`
}

func visualizationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	element := strings.ToLower(r.URL.Query().Get("element"))
	mode := strings.ToLower(r.URL.Query().Get("mode"))
	if element == "" {
		http.Error(w, "Element parameter is required", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Max-Age", "86400")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	visChan := make(chan VisualizationData)
	go func() {
		graph, err := LoadGraph("combinations.json")
		if err != nil {
			log.Printf("Error loading graph: %v", err)
			visChan <- VisualizationData{
				Nodes: []VisualizationNode{{
					ID:       "error",
					Label:    "Error loading graph",
					Children: make([]string, 0),
					Parent:   "",
					Depth:    0,
				}},
				Edges: make([]struct {
					From string `json:"from"`
					To   string `json:"to"`
				}, 0),
			}
			close(visChan)
			return
		}
		baseElements := []string{"air", "earth", "fire", "water"}
		timeout := 60 * time.Second
		if mode == "dfs" {
			DFSRecipeWithVisualization(graph, baseElements, element, timeout, visChan)
		} else {
			BFSRecipeWithVisualization(graph, baseElements, element, timeout, visChan)
		}
		close(visChan)
	}()
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	initialData := VisualizationData{
		Nodes: []VisualizationNode{{
			ID:       "root",
			Label:    "Root",
			Children: make([]string, 0),
			Parent:   "",
			Depth:    0,
		}},
		Edges: make([]struct {
			From string `json:"from"`
			To   string `json:"to"`
		}, 0),
	}
	jsonData, err := json.Marshal(initialData)
	if err != nil {
		log.Printf("Error marshaling initial data: %v", err)
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()
	for data := range visChan {
		jsonData, err := json.Marshal(data)
		if err != nil {
			log.Printf("Error marshaling visualization data: %v", err)
			continue
		}
		fmt.Fprintf(w, "data: %s\n\n", jsonData)
		flusher.Flush()
	}
	fmt.Fprintf(w, "event: end\ndata: {}\n\n")
	flusher.Flush()
}

func BFSRecipeWithVisualization(graph Graph, start []string, target string, timeout time.Duration, visChan chan VisualizationData) ([]Step, bool, int) {
	inverse := CreateInverseGraph(graph)
	queueSize := 10000
	forwardQueue := make([]State, 0, queueSize)
	initialState := State{
		Available: sliceToSet(start),
		Path:      []Step{},
		Depth:     0,
		ID:        "root",
		ParentID:  "",
	}
	forwardQueue = append(forwardQueue, initialState)
	visitedStates := make(map[string]struct{})
	visitedStates[stateHash(sliceToSet(start))] = struct{}{}
	nodesVisited := 0
	startTime := time.Now()
	maxDepth := 7
	sendVisualizationData(visChan, []State{initialState}, State{})
	if initialState.Available[target] {
		return []Step{}, true, 1
	}
	for len(forwardQueue) > 0 {
		if time.Since(startTime) > timeout {
			return nil, false, nodesVisited
		}
		current := forwardQueue[0]
		forwardQueue = forwardQueue[1:]
		nodesVisited++
		sendVisualizationData(visChan, forwardQueue, current)
		time.Sleep(500 * time.Millisecond) 
		if current.Depth >= maxDepth {
			continue
		}
		available := make([]string, 0, len(current.Available))
		for elem := range current.Available {
			available = append(available, elem)
		}
		for i := 0; i < len(available); i++ {
			a := available[i]
			for j := i; j < len(available); j++ {
				b := available[j]
				if _, exists := inverse[a]; !exists {
					continue
				}
				results, exists := inverse[a][b]
				if !exists {
					continue
				}
				for _, result := range results {
					if result == target {
						newPath := make([]Step, len(current.Path))
						copy(newPath, current.Path)
						newPath = append(newPath, Step{
							Ingredients: [2]string{a, b},
							Result:      result,
						})
						return newPath, true, nodesVisited
					}
					if !current.Available[result] {
						newAvailable := copyMap(current.Available)
						newAvailable[result] = true
						stateKey := stateHash(newAvailable)
						if _, visited := visitedStates[stateKey]; !visited {
							visitedStates[stateKey] = struct{}{}
							if len(forwardQueue) < queueSize {
								newState := State{
									Available: newAvailable,
									Path:      append(make([]Step, len(current.Path)), current.Path...),
									Depth:     current.Depth + 1,
									ID:        fmt.Sprintf("%s_%d", result, nodesVisited),
									ParentID:  current.ID,
								}
								newState.Path = append(newState.Path, Step{
									Ingredients: [2]string{a, b},
									Result:      result,
								})
								forwardQueue = append(forwardQueue, newState)
							}
						}
					}
				}
			}
		}
	}
	return nil, false, nodesVisited
}

func DFSRecipeWithVisualization(graph Graph, start []string, target string, timeout time.Duration, visChan chan VisualizationData) ([]Step, bool, int) {
	inverse := CreateInverseGraph(graph)
	stack := make([]State, 0, 10000)
	initialState := State{
		Available: sliceToSet(start),
		Path:      []Step{},
		Depth:     0,
		ID:        "root",
		ParentID:  "",
	}
	stack = append(stack, initialState)
	visitedStates := make(map[string]struct{})
	visitedStates[stateHash(sliceToSet(start))] = struct{}{}
	nodesVisited := 0
	startTime := time.Now()
	maxDepth := 7
	sendVisualizationData(visChan, []State{initialState}, State{})
	if initialState.Available[target] {
		return []Step{}, true, 1
	}
	for len(stack) > 0 {
		if time.Since(startTime) > timeout {
			return nil, false, nodesVisited
		}
		current := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		nodesVisited++
		sendVisualizationData(visChan, stack, current)
		time.Sleep(500 * time.Millisecond) 
		if current.Depth >= maxDepth {
			continue
		}
		available := make([]string, 0, len(current.Available))
		for elem := range current.Available {
			available = append(available, elem)
		}
		for i := 0; i < len(available); i++ {
			a := available[i]
			for j := i; j < len(available); j++ {
				b := available[j]
				if _, exists := inverse[a]; !exists {
					continue
				}
				results, exists := inverse[a][b]
				if !exists {
					continue
				}
				for _, result := range results {
					if result == target {
						newPath := make([]Step, len(current.Path))
						copy(newPath, current.Path)
						newPath = append(newPath, Step{
							Ingredients: [2]string{a, b},
							Result:      result,
						})
						return newPath, true, nodesVisited
					}
					if !current.Available[result] {
						newAvailable := copyMap(current.Available)
						newAvailable[result] = true
						stateKey := stateHash(newAvailable)
						if _, visited := visitedStates[stateKey]; !visited {
							visitedStates[stateKey] = struct{}{}
							if len(stack) < 10000 {
								newState := State{
									Available: newAvailable,
									Path:      append(make([]Step, len(current.Path)), current.Path...),
									Depth:     current.Depth + 1,
									ID:        fmt.Sprintf("%s_%d", result, nodesVisited),
									ParentID:  current.ID,
								}
								newState.Path = append(newState.Path, Step{
									Ingredients: [2]string{a, b},
									Result:      result,
								})
								stack = append(stack, newState)
							}
						}
					}
				}
			}
		}
	}
	return nil, false, nodesVisited
}

func sendVisualizationData(visChan chan VisualizationData, queue []State, currentState State) {
	data := VisualizationData{
		Nodes: make([]VisualizationNode, 0),
		Edges: make([]struct {
			From string `json:"from"`
			To   string `json:"to"`
		}, 0),
	}
	for _, state := range queue {
		label := "Root"
		if len(state.Path) > 0 {
			lastStep := state.Path[len(state.Path)-1]
			label = fmt.Sprintf("%s + %s\n= %s", 
				lastStep.Ingredients[0], 
				lastStep.Ingredients[1], 
				lastStep.Result)
		}
		node := VisualizationNode{
			ID:       state.ID,
			Label:    label,
			Children: make([]string, 0),
			Parent:   state.ParentID,
			Depth:    state.Depth,
		}
		data.Nodes = append(data.Nodes, node)
		if state.ParentID != "" {
			data.Edges = append(data.Edges, struct {
				From string `json:"from"`
				To   string `json:"to"`
			}{
				From: state.ParentID,
				To:   state.ID,
			})
		}
	}
	if currentState.ID != "" {
		label := "Root"
		if len(currentState.Path) > 0 {
			lastStep := currentState.Path[len(currentState.Path)-1]
			label = fmt.Sprintf("%s + %s\n= %s", 
				lastStep.Ingredients[0], 
				lastStep.Ingredients[1], 
				lastStep.Result)
		}
		node := VisualizationNode{
			ID:       currentState.ID,
			Label:    label,
			Children: make([]string, 0),
			Parent:   currentState.ParentID,
			Depth:    currentState.Depth,
		}
		data.Nodes = append(data.Nodes, node)
		if currentState.ParentID != "" {
			data.Edges = append(data.Edges, struct {
				From string `json:"from"`
				To   string `json:"to"`
			}{
				From: currentState.ParentID,
				To:   currentState.ID,
			})
		}
	}
	if len(data.Nodes) == 0 {
		data.Nodes = append(data.Nodes, VisualizationNode{
			ID:       "root",
			Label:    "Root",
			Parent:   "",
			Depth:    0,
		})
	}
	log.Printf("Sending visualization data: %+v", data)
	visChan <- data
}

func main() {
	log.Println("Starting server...")
	log.Println("Server will listen on :5000")
	http.HandleFunc("/search", enableCORS(searchHandler))
	http.HandleFunc("/visualize", enableCORS(visualizationHandler))
	if err := http.ListenAndServe(":5000", nil); err != nil {
		log.Fatal("Server failed to start:", err)
	}
} 