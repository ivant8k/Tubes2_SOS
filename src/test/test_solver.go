package main

import (
	"fmt"
	"littlealchemy/backend"
	"os"
	"strconv"
	"time"
)

func printTree(n *backend.Node, depth int) {
	if n == nil {
		return
	}
	for i := 0; i < depth; i++ {
		fmt.Print("  ")
	}
	fmt.Println(n.Element)
	printTree(n.Left, depth+1)
	printTree(n.Right, depth+1)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_solver.go <ElementTarget> [maxCount]")
		fmt.Println("Contoh:")
		fmt.Println("  go run test/test_solver.go Brick")
		fmt.Println("  go run test/test_solver.go Airplane 5")
		return
	}

	target := os.Args[1]
	maxCount := 3
	if len(os.Args) >= 3 {
		if val, err := strconv.Atoi(os.Args[2]); err == nil {
			maxCount = val
		}
	}

	err := backend.LoadCombinations("combinations.json")
	if err != nil {
		fmt.Println("Failed to load combination:", err)
		return
	}

	// === BFS ===
	fmt.Println("===BFS===")
	fmt.Println("elemen yang dicari:", target)
	start := time.Now()
	tree := backend.FindRecipeBFS(target)
	duration := time.Since(start)
	if tree != nil {
		printTree(tree, 0)
		fmt.Println("hasil: sukses")
	} else {
		fmt.Println("hasil: tidak ditemukan")
	}
	fmt.Println("waktu pencarian:", duration)
	fmt.Println("jumlah node yang dikunjungi:", backend.GetBFSVisited())

	// === DFS ===
	fmt.Println("\n===DFS===")
	fmt.Println("elemen yang dicari:", target)
	start = time.Now()
	tree2 := backend.FindRecipeDFS(target, make(map[string]bool))
	duration = time.Since(start)
	if tree2 != nil {
		printTree(tree2, 0)
		fmt.Println("hasil: sukses")
	} else {
		fmt.Println("hasil: tidak ditemukan")
	}
	fmt.Println("waktu pencarian:", duration)
	fmt.Println("jumlah node yang dikunjungi:", backend.GetDFSVisited())

	// === Multirecipe ===
	fmt.Println("\n===Multirecipe===")
	fmt.Println("elemen yang dicari:", target)
	fmt.Println("maksimal jumlah recipe:", maxCount)
	start = time.Now()
	trees := backend.FindMultipleRecipes(target, maxCount)
	duration = time.Since(start)
	if len(trees) > 0 {
		for i, t := range trees {
			fmt.Printf("Recipe #%d:\n", i+1)
			printTree(t, 0)
		}
		fmt.Println("hasil:", len(trees), "recipe ditemukan")
	} else {
		fmt.Println("hasil: tidak ditemukan")
	}
	fmt.Println("waktu pencarian:", duration)
	fmt.Println("jumlah node yang dikunjungi:", backend.GetMultiVisited())

	// === Bidirectional ===
	fmt.Println("\n===Bidirectional===")
	fmt.Println("elemen yang dicari:", target)
	start = time.Now()
	tree3 := backend.FindRecipeBidirectional(target)
	duration = time.Since(start)
	if tree3 != nil {
		printTree(tree3, 0)
		fmt.Println("hasil: sukses")
	} else {
		fmt.Println("hasil: tidak ditemukan")
	}
	fmt.Println("waktu pencarian:", duration)
	fmt.Println("jumlah node yang dikunjungi:", backend.GetBidirectionalVisited())

}
