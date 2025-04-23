package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
)

// Graph: result → list of ingredient pairs
type Graph map[string][][]string

type Step struct {
	Ingredients [2]string `json:"ingredients"`
	Result      string    `json:"result"`
}

// LoadGraph loads graph from graph_combinations.json
func LoadGraph(path string) (Graph, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var graph Graph
	err = json.Unmarshal(file, &graph)
	return graph, err
}

// BFSRecipe searches for target element using BFS
func BFSRecipe(graph Graph, start []string, target string) ([]Step, bool) {
	type State struct {
		Available map[string]bool
		Path      []Step
	}

	queue := []State{{Available: sliceToSet(start), Path: []Step{}}}
	visited := map[string]bool{}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		if curr.Available[target] {
			return curr.Path, true
		}

		elements := keys(curr.Available)
		n := len(elements)

		for i := 0; i < n; i++ {
			for j := i + 1; j < n; j++ {
				a, b := elements[i], elements[j]

				for result, combos := range graph {
					for _, combo := range combos {
						if isPair(combo, a, b) && !curr.Available[result] {
							newAvail := copyMap(curr.Available)
							newAvail[result] = true

							step := Step{Ingredients: [2]string{a, b}, Result: result}
							newPath := append([]Step{}, curr.Path...)
							newPath = append(newPath, step)

							key := sortedKey(newAvail)
							if !visited[key] {
								visited[key] = true
								queue = append(queue, State{Available: newAvail, Path: newPath})
							}
						}
					}
				}
			}
		}
	}
	return nil, false
}

// Util

func sliceToSet(slice []string) map[string]bool {
	set := make(map[string]bool)
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
	newMap := make(map[string]bool)
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

func isPair(pair []string, a, b string) bool {
	return (pair[0] == a && pair[1] == b) || (pair[0] == b && pair[1] == a)
}

func sortedKey(m map[string]bool) string {
	k := keys(m)
	sort.Strings(k)
	return fmt.Sprintf("%v", k)
}
