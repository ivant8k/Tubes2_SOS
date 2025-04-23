package main

import (
	"fmt"
	"littlealchemy-scraper/backend" // <- GANTI dengan nama module kamu (lihat go.mod)
)

func main() {
	graph, err := backend.LoadGraph("src/graph_combinations.json") // path relatif dari root project
	if err != nil {
		panic(err)
	}

	start := []string{"earth", "water", "fire"}
	target := "brick"

	path, found := backend.BFSRecipe(graph, start, target)
	if found {
		for _, step := range path {
			fmt.Printf("%s + %s → %s\n", step.Ingredients[0], step.Ingredients[1], step.Result)
		}
	} else {
		fmt.Println("❌ Target tidak ditemukan.")
	}
}
