package main
import (
	"littlealchemy/backend"
	"fmt"
	"time"
)
var visitedNodeCount int
func resetVisitedCount() {
	visitedNodeCount = 0
	
}
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
	err := backend.LoadCombinations("combination.json")
	if err != nil {
		fmt.Println("Failed to load combination:", err)
		return
	}
	target := "Airplane"
	fmt.Println("===BFS===")
	fmt.Println("elemen yang dicari:", target)
	start := time.Now()
	resetVisitedCount()
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
	fmt.Println("\n===DFS===")
	fmt.Println("elemen yang dicari:", target)
	start = time.Now()
	resetVisitedCount()
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
	fmt.Println("\n===Multirecipe===")
	fmt.Println("elemen yang dicari:", target)
	start = time.Now()
	resetVisitedCount()
	trees := backend.FindMultipleRecipes(target, 3)
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
}
