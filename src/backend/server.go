package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Combination struct {
	Root  string `json:"root"`
	Left  string `json:"left"`
	Right string `json:"right"`
	Tier  int    `json:"tier,string"`
}

type Node struct {
	Element string
	Left    *Node
	Right   *Node
}

var combinations map[string][]Combination
var tierMap map[string]int
var BFSVisitedCount int
var MultiVisitedCount int32
var DFSVisitedCount int

func LoadCombinations(filename string) error {
	fmt.Println("Loading combinations from:", filename)
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file: %v", err)
	}
	var raw []Combination
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %v", err)
	}
	combinations = make(map[string][]Combination)
	tierMap = make(map[string]int)
	for _, c := range raw {
		combinations[c.Root] = append(combinations[c.Root], c)
		tierMap[c.Root] = c.Tier
	}
	fmt.Printf("Loaded %d combinations\n", len(raw))
	return nil
}

func isBasic(element string) bool {
	basics := map[string]bool{
		"Earth": true,
		"Air":   true,
		"Water": true,
		"Fire":  true,
		"Time":  true,
	}
	return basics[element]
}

func FindRecipeBFS(target string) *Node {
	fmt.Printf("\n=== Starting BFS search for: %s ===\n", target)
	
	if isBasic(target) {
		fmt.Printf("Found basic element: %s\n", target)
		BFSVisitedCount = 1
		return &Node{Element: target}
	}

	if _, exists := combinations[target]; !exists {
		fmt.Printf("Element %s not found in combinations\n", target)
		BFSVisitedCount = 0
		return nil
	}

	fmt.Printf("Found %d combinations for %s\n", len(combinations[target]), target)
	visited := make(map[string]bool)
	recipeMap := make(map[string]*Node)
	queue := []string{target}
	BFSVisitedCount = 0

	// First pass: collect all possible combinations
	fmt.Println("\nFirst pass: Collecting combinations...")
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		BFSVisitedCount++
		fmt.Printf("Visiting: %s (visited count: %d)\n", current, BFSVisitedCount)

		if isBasic(current) {
			fmt.Printf("Found basic element: %s\n", current)
			recipeMap[current] = &Node{Element: current}
			continue
		}

		// Add all possible combinations to the queue
		for _, comb := range combinations[current] {
			// Check if both ingredients are of lower tier than the target
			if tierMap[comb.Left] < tierMap[current] && tierMap[comb.Right] < tierMap[current] {
				fmt.Printf("  Checking combination: %s (tier %d) + %s (tier %d) = %s (tier %d)\n", 
					comb.Left, tierMap[comb.Left], comb.Right, tierMap[comb.Right], comb.Root, comb.Tier)
				if !visited[comb.Left] {
					queue = append(queue, comb.Left)
					fmt.Printf("    Added to queue: %s\n", comb.Left)
				}
				if !visited[comb.Right] {
					queue = append(queue, comb.Right)
					fmt.Printf("    Added to queue: %s\n", comb.Right)
				}
			} else {
				fmt.Printf("  Skipping invalid combination: %s (tier %d) + %s (tier %d) = %s (tier %d)\n",
					comb.Left, tierMap[comb.Left], comb.Right, tierMap[comb.Right], comb.Root, comb.Tier)
			}
		}
	}

	// Second pass: build recipes from basic elements up
	fmt.Println("\nSecond pass: Building recipes...")
	changed := true
	for changed {
		changed = false
		for elem := range visited {
			if recipeMap[elem] != nil {
				continue
			}

			// Try all combinations for this element
			for _, comb := range combinations[elem] {
				// Check if both ingredients are of lower tier than the target
				if tierMap[comb.Left] < tierMap[elem] && tierMap[comb.Right] < tierMap[elem] {
					leftRecipe := recipeMap[comb.Left]
					rightRecipe := recipeMap[comb.Right]
					if leftRecipe != nil && rightRecipe != nil {
						fmt.Printf("Found recipe for %s: %s + %s\n", 
							elem, comb.Left, comb.Right)
						recipeMap[elem] = &Node{
							Element: elem,
							Left:    leftRecipe,
							Right:   rightRecipe,
						}
						changed = true
						break
					}
				}
			}
		}
	}

	result := recipeMap[target]
	if result != nil {
		fmt.Printf("\nSuccessfully found recipe for %s\n", target)
	} else {
		fmt.Printf("\nNo valid recipe found for %s\n", target)
	}
	return result
}

func FindRecipeDFS(target string, visited map[string]bool) *Node {
	if _, exists := combinations[target]; !exists && !isBasic(target) {
		return nil
	}

	if isBasic(target) {
		DFSVisitedCount++
		return &Node{Element: target}
	}

	if visited == nil {
		visited = make(map[string]bool)
		DFSVisitedCount = 0
	}

	if visited[target] {
		return nil
	}

	visited[target] = true
	DFSVisitedCount++
	defer func() { visited[target] = false }()

	for _, comb := range combinations[target] {
		if tierMap[comb.Left] < tierMap[target] && tierMap[comb.Right] < tierMap[target] {
			left := FindRecipeDFS(comb.Left, visited)
			if left == nil {
				continue
			}
			right := FindRecipeDFS(comb.Right, visited)
			if right != nil {
				return &Node{Element: target, Left: left, Right: right}
			}
		}
	}

	return nil
}

func FindMultipleRecipes(target string, maxCount int) []*Node {
	if isBasic(target) {
		atomic.StoreInt32(&MultiVisitedCount, 1)
		return []*Node{{Element: target}}
	}
	if _, exists := combinations[target]; !exists {
		return nil
	}

	atomic.StoreInt32(&MultiVisitedCount, 0)
	var results []*Node
	resultChan := make(chan *Node, maxCount)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 30) // Increased from 20 to 30 concurrent searches
	var seen sync.Map
	validCombos := []Combination{}

	for _, c := range combinations[target] {
		if tierMap[c.Left] < tierMap[target] && tierMap[c.Right] < tierMap[target] {
			validCombos = append(validCombos, c)
		}
	}

	sort.SliceStable(validCombos, func(i, j int) bool {
		return validCombos[i].Left+validCombos[i].Right < validCombos[j].Left+validCombos[j].Right
	})

	for _, comb := range validCombos {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(c Combination) {
			defer wg.Done()
			defer func() { <-semaphore }()
			
			visited := make(map[string]bool)
			left := exploreRecipe(c.Left, visited, &MultiVisitedCount)
			if left == nil {
				return
			}
			right := exploreRecipe(c.Right, visited, &MultiVisitedCount)
			if right == nil {
				return
			}
			tree := &Node{Element: target, Left: left, Right: right}
			signature := serializeTree(tree)
			if _, ok := seen.LoadOrStore(signature, true); !ok {
				resultChan <- tree
			}
		}(comb)
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	for recipe := range resultChan {
		results = append(results, recipe)
		if len(results) >= maxCount {
			break
		}
	}

	return results
}

func exploreRecipe(target string, visited map[string]bool, counter *int32) *Node {
	if isBasic(target) {
		atomic.AddInt32(counter, 1)
		return &Node{Element: target}
	}
	if visited[target] {
		return nil
	}
	visited[target] = true
	atomic.AddInt32(counter, 1)

	var candidates []*Node
	for _, comb := range combinations[target] {
		if tierMap[comb.Left] < tierMap[target] && tierMap[comb.Right] < tierMap[target] {
			leftVisited := copyVisitedMap(visited)
			left := exploreRecipe(comb.Left, leftVisited, counter)
			if left == nil {
				continue
			}
			rightVisited := copyVisitedMap(visited)
			right := exploreRecipe(comb.Right, rightVisited, counter)
			if right == nil {
				continue
			}
			candidates = append(candidates, &Node{Element: target, Left: left, Right: right})
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Return a random candidate instead of always the first one
	// This helps find different recipe paths
	rand.Seed(time.Now().UnixNano())
	return candidates[rand.Intn(len(candidates))]
}

func copyVisitedMap(original map[string]bool) map[string]bool {
	copy := make(map[string]bool)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

func treeDepth(node *Node) int {
	if node == nil {
		return 0
	}
	leftDepth := treeDepth(node.Left)
	rightDepth := treeDepth(node.Right)
	if leftDepth > rightDepth {
		return leftDepth + 1
	}
	return rightDepth + 1
}

func serializeTree(n *Node) string {
	if n == nil {
		return ""
	}
	if n.Left == nil && n.Right == nil {
		return n.Element
	}
	leftStr := serializeTree(n.Left)
	rightStr := serializeTree(n.Right)
	if leftStr > rightStr {
		return n.Element + "(" + rightStr + "," + leftStr + ")"
	}
	return n.Element + "(" + leftStr + "," + rightStr + ")"
}

func IsBasic(element string) bool {
	return isBasic(element)
}

func GetCombinations(element string) []Combination {
	return combinations[element]
}

func IsLowerTier(c Combination) bool {
	return tierMap[c.Left] < tierMap[c.Root] && tierMap[c.Right] < tierMap[c.Root]
}

func GetBFSVisited() int {
	return BFSVisitedCount
}

func GetDFSVisited() int {
	return DFSVisitedCount
}

func GetMultiVisited() int {
	return int(atomic.LoadInt32(&MultiVisitedCount))
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	element := r.URL.Query().Get("element")
	if element == "" {
		http.Error(w, "Element parameter is required", http.StatusBadRequest)
		return
	}

	mode := r.URL.Query().Get("mode")
	fmt.Printf("\n=== Search Request ===\n")
	fmt.Printf("Element: %s (Tier: %d)\n", element, tierMap[element])
	fmt.Printf("Mode: %s\n", mode)

	var result *Node
	var visited int
	var response struct {
		Found  bool     `json:"found"`
		Steps  int      `json:"steps"`
		Paths  [][]Step `json:"paths"`
		Target struct {
			Element string `json:"element"`
			Tier    int    `json:"tier"`
		} `json:"target"`
	}

	// Set target information
	response.Target.Element = element
	response.Target.Tier = tierMap[element]

	switch mode {
	case "bfs":
		result = FindRecipeBFS(element)
		visited = GetBFSVisited()
		if result != nil {
			path := convertRecipeToPath(result)
			response = struct {
				Found  bool     `json:"found"`
				Steps  int      `json:"steps"`
				Paths  [][]Step `json:"paths"`
				Target struct {
					Element string `json:"element"`
					Tier    int    `json:"tier"`
				} `json:"target"`
			}{
				Found: true,
				Steps: visited,
				Paths: [][]Step{path},
				Target: struct {
					Element string `json:"element"`
					Tier    int    `json:"tier"`
				}{
					Element: element,
					Tier:    tierMap[element],
				},
			}
		}
	case "dfs":
		result = FindRecipeDFS(element, nil)
		visited = GetDFSVisited()
		if result != nil {
			path := convertRecipeToPath(result)
			response = struct {
				Found  bool     `json:"found"`
				Steps  int      `json:"steps"`
				Paths  [][]Step `json:"paths"`
				Target struct {
					Element string `json:"element"`
					Tier    int    `json:"tier"`
				} `json:"target"`
			}{
				Found: true,
				Steps: visited,
				Paths: [][]Step{path},
				Target: struct {
					Element string `json:"element"`
					Tier    int    `json:"tier"`
				}{
					Element: element,
					Tier:    tierMap[element],
				},
			}
		}
	case "multi":
		results := FindMultipleRecipes(element, 10)
		visited = GetMultiVisited()
		if len(results) > 0 {
			paths := make([][]Step, 0, len(results))
			for _, result := range results {
				path := convertRecipeToPath(result)
				paths = append(paths, path)
			}
			response = struct {
				Found  bool     `json:"found"`
				Steps  int      `json:"steps"`
				Paths  [][]Step `json:"paths"`
				Target struct {
					Element string `json:"element"`
					Tier    int    `json:"tier"`
				} `json:"target"`
			}{
				Found: true,
				Steps: visited,
				Paths: paths,
				Target: struct {
					Element string `json:"element"`
					Tier    int    `json:"tier"`
				}{
					Element: element,
					Tier:    tierMap[element],
				},
			}
		}
	default:
		fmt.Printf("Invalid mode: %s\n", mode)
		http.Error(w, "Invalid mode", http.StatusBadRequest)
		return
	}

	if len(response.Paths) == 0 {
		fmt.Printf("No recipe found for %s (Tier: %d)\n", element, tierMap[element])
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	fmt.Printf("Found %d recipes with %d visited nodes\n", len(response.Paths), visited)
	for i, path := range response.Paths {
		fmt.Printf("\nRecipe %d:\n", i+1)
		for _, step := range path {
			fmt.Printf("%s (Tier %d) + %s (Tier %d) = %s (Tier %d)\n",
				step.Ingredients[0], step.Tiers.Left,
				step.Ingredients[1], step.Tiers.Right,
				step.Result, step.Tiers.Result)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

type Step struct {
	Ingredients []string `json:"ingredients"`
	Result      string   `json:"result"`
	Tiers       struct {
		Left   int `json:"left"`
		Right  int `json:"right"`
		Result int `json:"result"`
	} `json:"tiers"`
}

func convertRecipeToPath(node *Node) []Step {
	if node == nil {
		return nil
	}

	// If it's a leaf node (basic element), return empty
	if node.Left == nil && node.Right == nil {
		return nil
	}

	// Get paths for left and right subtrees
	leftSteps := convertRecipeToPath(node.Left)
	rightSteps := convertRecipeToPath(node.Right)

	// Create the current step
	currentStep := Step{
		Ingredients: []string{node.Left.Element, node.Right.Element},
		Result:      node.Element,
	}
	currentStep.Tiers.Left = tierMap[node.Left.Element]
	currentStep.Tiers.Right = tierMap[node.Right.Element]
	currentStep.Tiers.Result = tierMap[node.Element]

	// Combine all steps in the correct order
	// First add all steps from left subtree, then right subtree, then current step
	steps := make([]Step, 0)
	steps = append(steps, leftSteps...)
	steps = append(steps, rightSteps...)
	steps = append(steps, currentStep)

	return steps
}

func handleMode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	modes := []string{"bfs", "dfs", "multi"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(modes)
}

func main() {
	fmt.Println("Starting server...")
	
	// Load combinations from file
	err := LoadCombinations("combinations.json")
	if err != nil {
		fmt.Printf("Error loading combinations: %v\n", err)
		panic(err)
	}

	// Define HTTP handlers with CORS
	http.HandleFunc("/search", enableCORS(handleSearch))
	http.HandleFunc("/mode", enableCORS(handleMode))

	// Start server
	port := ":5000"
	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		panic(err)
	}
}