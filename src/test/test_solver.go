package main

import (
	"fmt"
	"time"
	"littlealchemy-scraper/backend"
)

func main() {
    startTime := time.Now()

    start := []string{"earth", "water", "fire", "air"}
    target := "airplane"

    graph, err := backend.LoadGraph("graph_combinations.json")
    if err != nil {
        panic(err)
    }

    path, found, visited := backend.BFSRecipe(graph, start, target)

    if found {
        for _, step := range path {
            fmt.Printf("%s + %s → %s\n", step.Ingredients[0], step.Ingredients[1], step.Result)
        }
        fmt.Printf("Total langkah: %d\n", len(path))
    } else {
        fmt.Printf("Go find by yourself dawg💀💀💀\n")
    }

    duration := time.Since(startTime)
    fmt.Printf("Node dikunjungi: %d\n", visited)
    fmt.Printf("Waktu eksekusi: %s\n", duration)
}