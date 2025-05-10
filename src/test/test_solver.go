package main

import (
	"fmt"
	"time"
	"littlealchemy/backend"
)

func main() {
	startTime := time.Now()

	start := []string{"earth", "water", "fire", "air"}
	target := "airplane"

	graph, err := backend.LoadGraph("combinations.json")
	if err != nil {
		panic(err)
	}

	// ===== BFS =====
	fmt.Println("=== BFS ===")
	bfsPath, bfsFound, bfsVisited := backend.BFSRecipe(graph, start, target, 5*time.Second)

	if bfsFound {
		for _, step := range bfsPath {
			fmt.Printf("%s + %s → %s\n", step.Ingredients[0], step.Ingredients[1], step.Result)
		}
		fmt.Printf("Total langkah: %d\n", len(bfsPath))
	} else {
		fmt.Printf("Go find by yourself dawg💀💀💀\n")
	}

	fmt.Printf("Node dikunjungi: %d\n", bfsVisited)

	// ===== DFS =====
	fmt.Println("\n=== DFS ===")
	dfsPath, dfsFound, dfsVisited := backend.DFSRecipe(graph, start, target, 5*time.Second)

	if dfsFound {
		for _, step := range dfsPath {
			fmt.Printf("%s + %s → %s\n", step.Ingredients[0], step.Ingredients[1], step.Result)
		}
		fmt.Printf("Total langkah: %d\n", len(dfsPath))
	} else {
		fmt.Printf("Go find by yourself dawg💀💀💀\n")
	}

	fmt.Printf("Node dikunjungi: %d\n", dfsVisited)

	// Total waktu
	duration := time.Since(startTime)
	fmt.Printf("\nTotal waktu eksekusi: %s\n", duration)
}
