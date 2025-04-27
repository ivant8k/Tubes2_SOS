package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Graph: result → list of ingredient pairs
type Graph map[string][][]string

type InverseGraph map[string]map[string][]string

type Step struct {
	Ingredients [2]string `json:"ingredients"`
	Result      string    `json:"result"`
}

// LoadGraph loads combinations from JSON file
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

// BFSRecipe finds combination path from starters to target using optimized BFS
func BFSRecipe(graph Graph, start []string, target string) ([]Step, bool, int) {
	// Build inverse graph for faster lookups
	inverse := CreateInverseGraph(graph)
	
	type State struct {
		Available map[string]bool
		Path      []Step
	}

	// Initialize queue with starting elements
	queue := []State{{Available: sliceToSet(start), Path: []Step{}}}
	
	// Use string hash of available elements as visited key
	visitedStates := make(map[string]bool, 1000)
	visitedStates[stateHash(sliceToSet(start))] = true
	
	nodesVisited := 0

	// Early check if target is already in starting elements
	if queue[0].Available[target] {
		return queue[0].Path, true, nodesVisited
	}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		nodesVisited++
		fmt.Printf("Node visited: %d\r", nodesVisited)
		// if nodesVisited%1000 == 0 {
		// 	fmt.Printf("Nodes visited: %d, Queue size: %d\r", nodesVisited, len(queue))
		// }
		if nodesVisited == 2000000 {
			fmt.Printf("Shit I'm tired dawg💀💀💀\n")
			break
		}
		elements := keys(curr.Available)

		for i, a := range elements {
			for j := i; j < len(elements); j++ {
				b := elements[j]
				
				var possibleResults []string
				if pairs, exists := inverse[a]; exists {
					if results, hasPair := pairs[b]; hasPair {
						possibleResults = results
					}
				}
				
				for _, result := range possibleResults {
					if curr.Available[result] {
						continue
					}
					
					newAvail := copyMap(curr.Available)
					newAvail[result] = true
					
					isTarget := result == target
					
					step := Step{
						Ingredients: [2]string{a, b},
						Result:      result,
					}
					
					newPath := make([]Step, len(curr.Path)+1)
					copy(newPath, curr.Path)
					newPath[len(curr.Path)] = step
					
					if isTarget {
						return newPath, true, nodesVisited
					}
					
					stateKey := stateHash(newAvail)
					if !visitedStates[stateKey] {
						visitedStates[stateKey] = true
						queue = append(queue, State{
							Available: newAvail,
							Path:      newPath,
						})
					}
				}
			}
		}
	}
	
	return nil, false, nodesVisited
}

func stateHash(available map[string]bool) string {
	keys := make([]string, 0, len(available))
	for k := range available {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

//
// Helper Utilities
//

func sliceToSet(slice []string) map[string]bool {
	set := make(map[string]bool, len(slice))
	for _, v := range slice {
		set[v] = true
	}
	return set
}

func keys(m map[string]bool) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}

func copyMap(m map[string]bool) map[string]bool {
	newMap := make(map[string]bool, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}