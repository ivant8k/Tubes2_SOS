package backend

import (
	"container/heap"
	"log"
	"sort"
	"strings"
	"time"
)

// Step represents a combination step with ingredients and result.
type Step struct {
	Ingredients [2]string `json:"ingredients"`
	Result      string    `json:"result"`
}

// State represents the current state in the search process.
type State struct {
	Available map[string]bool
	Path      []Step
	Depth     int
	Priority  int // For A* search
}

// Graph maps a result to a list of ingredient pairs.
type Graph map[string][][]string

// InverseGraph maps ingredients to a list of possible results.
type InverseGraph map[string]map[string][]string

// PriorityQueue implementation
type PriorityQueue []*State

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*State)
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// BFSRecipe finds a recipe for the target element using Breadth-First Search with optimizations.
func BFSRecipe(graph Graph, start []string, target string, timeout time.Duration) ([]Step, bool, int) {
	inverse := CreateInverseGraph(graph)
	
	// Check if target exists in the graph to fail early
	if _, exists := graph[target]; !exists {
		found := false
		for _, pairs := range graph {
			for _, pair := range pairs {
				for _, elem := range pair {
					if elem == target {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return nil, false, 0
		}
	}
	
	queue := []State{{Available: sliceToSet(start), Path: nil, Depth: 0}}
	startHash := stateHash(sliceToSet(start))
	visited := map[string]bool{startHash: true}
	startTime := time.Now()
	nodesVisited := 0
	
	// Pre-check if target is already in start elements
	if sliceToSet(start)[target] {
		return nil, true, 0 // Target is already available
	}

	for len(queue) > 0 {
		if time.Since(startTime) > timeout {
			log.Printf("BFS timeout after %v", timeout)
			return nil, false, nodesVisited
		}

		current := queue[0]
		queue = queue[1:]
		nodesVisited++

		// Found target
		if current.Available[target] {
			return current.Path, true, nodesVisited
		}

		// Get possible combinations
		combinations := getPossibleCombinations(current.Available, inverse)
		
		for _, combo := range combinations {
			a, b, res := combo[0], combo[1], combo[2]
			
			// Skip if we already have this result
			if current.Available[res] {
				continue
			}
			
			// Create new state
			newAvail := copyMap(current.Available)
			newAvail[res] = true
			hash := stateHash(newAvail)
			
			if !visited[hash] {
				visited[hash] = true
				newPath := append(append([]Step{}, current.Path...), Step{
					Ingredients: [2]string{a, b},
					Result:      res,
				})
				
				// Early exit if target found
				if res == target {
					return newPath, true, nodesVisited + 1
				}
				
				queue = append(queue, State{
					Available: newAvail,
					Path:      newPath,
					Depth:     current.Depth + 1,
				})
			}
		}
	}
	return nil, false, nodesVisited
}

// DFSRecipe with optimizations
func DFSRecipe(graph Graph, start []string, target string, timeout time.Duration) ([]Step, bool, int) {
	inverse := CreateInverseGraph(graph)
	
	// Check if target exists in the graph to fail early
	if _, exists := graph[target]; !exists {
		found := false
		for _, pairs := range graph {
			for _, pair := range pairs {
				for _, elem := range pair {
					if elem == target {
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return nil, false, 0
		}
	}
	
	// Use priority queue for DFS with some heuristic
	pq := &PriorityQueue{}
	heap.Init(pq)
	
	startState := &State{
		Available: sliceToSet(start),
		Path:      nil,
		Depth:     0,
		Priority:  0, // Lower is better
	}
	
	heap.Push(pq, startState)
	visited := map[string]bool{stateHash(startState.Available): true}
	startTime := time.Now()
	nodesVisited := 0
	maxDepth := 20 // Limit depth to avoid deep recursion
	
	// Pre-check if target is already in start elements
	if sliceToSet(start)[target] {
		return nil, true, 0 // Target is already available
	}

	for pq.Len() > 0 {
		if time.Since(startTime) > timeout {
			log.Printf("DFS timeout after %v", timeout)
			return nil, false, nodesVisited
		}

		current := heap.Pop(pq).(*State)
		nodesVisited++

		// Found target
		if current.Available[target] {
			return current.Path, true, nodesVisited
		}
		
		// Skip if we've reached max depth
		if current.Depth >= maxDepth {
			continue
		}

		// Get possible combinations efficiently
		combinations := getPossibleCombinations(current.Available, inverse)
		
		// Process combinations in reverse order for DFS
		for i := len(combinations) - 1; i >= 0; i-- {
			combo := combinations[i]
			a, b, res := combo[0], combo[1], combo[2]
			
			// Skip if we already have this result
			if current.Available[res] {
				continue
			}
			
			// Create new state
			newAvail := copyMap(current.Available)
			newAvail[res] = true
			hash := stateHash(newAvail)
			
			if !visited[hash] {
				visited[hash] = true
				newPath := append(append([]Step{}, current.Path...), Step{
					Ingredients: [2]string{a, b},
					Result:      res,
				})
				
				// Early exit if target found
				if res == target {
					return newPath, true, nodesVisited + 1
				}
				
				// Calculate priority - prefer states with fewer steps
				priority := current.Depth + 1
				
				// Push to priority queue
				heap.Push(pq, &State{
					Available: newAvail,
					Path:      newPath,
					Depth:     current.Depth + 1,
					Priority:  priority,
				})
			}
		}
	}
	return nil, false, nodesVisited
}

// Helper function to get all possible new combinations from current state
func getPossibleCombinations(available map[string]bool, inverse InverseGraph) [][3]string {
	var combinations [][3]string
	seen := make(map[string]bool)
	
	// Get all available elements
	availableSlice := keys(available)
	
	// For each pair of available elements
	for i := range availableSlice {
		a := availableSlice[i]
		
		// Check for combinations with this element
		if resultMap, exists := inverse[a]; exists {
			for j := i; j < len(availableSlice); j++ {
				b := availableSlice[j]
				
				// Check if this combination produces any results
				if results, ok := resultMap[b]; ok {
					for _, res := range results {
						// Skip if we've seen this result before
						key := a + "+" + b + "=" + res
						if !seen[key] && !seen[b + "+" + a + "=" + res] {
							seen[key] = true
							combinations = append(combinations, [3]string{a, b, res})
						}
					}
				}
			}
		}
	}
	
	return combinations
}

// CreateInverseGraph creates an inverse mapping of the graph for efficient lookup.
func CreateInverseGraph(graph Graph) InverseGraph {
	inv := make(InverseGraph)
	for result, pairs := range graph {
		for _, pair := range pairs {
			a, b := pair[0], pair[1]
			if inv[a] == nil {
				inv[a] = make(map[string][]string)
			}
			inv[a][b] = append(inv[a][b], result)
			if a != b {
				if inv[b] == nil {
					inv[b] = make(map[string][]string)
				}
				inv[b][a] = append(inv[b][a], result)
			}
		}
	}
	return inv
}

// stateHash generates a unique hash for a state based on available elements.
func stateHash(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

// sliceToSet converts a slice of strings to a set (map).
func sliceToSet(slice []string) map[string]bool {
	set := make(map[string]bool)
	for _, v := range slice {
		set[v] = true
	}
	return set
}

// copyMap creates a deep copy of a map.
func copyMap(m map[string]bool) map[string]bool {
	newMap := make(map[string]bool)
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

// keys returns a slice of all keys in a map.
func keys(m map[string]bool) []string {
	res := make([]string, 0, len(m))
	for k := range m {
		res = append(res, k)
	}
	return res
}