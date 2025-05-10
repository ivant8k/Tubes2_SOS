package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"littlealchemy/backend"
)

// ComboEntry represents a combination entry from the JSON data.
type ComboEntry struct {
	Root  string `json:"root"`
	Left  string `json:"left"`
	Right string `json:"right"`
	Tier  string `json:"tier"`
}

// LoadGraphFromCombinationJSON loads combination data from a JSON file into a backend.Graph.
func LoadGraphFromCombinationJSON(path string) (backend.Graph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var entries []ComboEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	graph := make(backend.Graph)
	for _, entry := range entries {
		if entry.Left != "" && entry.Right != "" {
			graph[entry.Root] = append(graph[entry.Root], []string{entry.Left, entry.Right})
		}
	}
	return graph, nil
}

func main() {
	startTime := time.Now()

	start := []string{"Earth", "Water", "Fire", "Air"}
	target := "Airplane"

	// Load graph from combination.json
	graph, err := LoadGraphFromCombinationJSON("combination.json")
	if err != nil {
		panic(err)
	}

	// BFS search
	fmt.Println("=== BFS ===")
	bfsPath, bfsFound, bfsVisited := backend.BFSRecipe(graph, start, target, 5*time.Second)
	if bfsFound {
		for _, step := range bfsPath {
			fmt.Printf("%s + %s → %s\n", step.Ingredients[0], step.Ingredients[1], step.Result)
		}
		fmt.Printf("Total langkah: %d\n", len(bfsPath))
	} else {
		fmt.Println("Go find by yourself dawg💀💀💀")
	}
	fmt.Printf("Node dikunjungi: %d\n", bfsVisited)

	// DFS search
	fmt.Println("\n=== DFS ===")
	dfsPath, dfsFound, dfsVisited := backend.DFSRecipe(graph, start, target, 5*time.Second)
	if dfsFound {
		for _, step := range dfsPath {
			fmt.Printf("%s + %s → %s\n", step.Ingredients[0], step.Ingredients[1], step.Result)
		}
		fmt.Printf("Total langkah: %d\n", len(dfsPath))
	} else {
		fmt.Println("Go find by yourself dawg💀💀💀")
	}
	fmt.Printf("Node dikunjungi: %d\n", dfsVisited)

	// Total execution time
	fmt.Printf("\nTotal waktu eksekusi: %s\n", time.Since(startTime))
}