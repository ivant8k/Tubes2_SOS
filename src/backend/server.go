package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Combination struct {
	Root  string `json:"root"`
	Left  string `json:"left"`
	Right string `json:"right"`
	Tier  string `json:"tier"`
}

type Node struct {
	Element string
	Left    *Node
	Right   *Node
}

type SearchResponse struct {
	Found bool     `json:"found"`
	Steps int      `json:"steps"`
	Paths [][]Step `json:"paths"`
}

type Step struct {
	Ingredients []string `json:"ingredients"`
	Result      string   `json:"result"`
}

var combinations map[string][]Combination
var tierMap map[string]int
var BFSVisitedCount int
var MultiVisitedCount int
var DFSVisitedCount int

func LoadCombinations(filename string) error {
	log.Printf("Loading combinations from %s", filename)
	
	data, err := os.ReadFile(filename)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return err
	}
	
	var raw []Combination
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("Error unmarshaling JSON: %v", err)
		return err
	}
	
	log.Printf("Successfully loaded %d combinations", len(raw))
	
	combinations = make(map[string][]Combination)
	tierMap = make(map[string]int)
	
	for _, c := range raw {
		// Convert tier from string to int
		tier, err := strconv.Atoi(c.Tier)
		if err != nil {
			log.Printf("Warning: Invalid tier value for %s: %s", c.Root, c.Tier)
			tier = 0
		}
		
		// Store the combination
		combinations[c.Root] = append(combinations[c.Root], c)
		tierMap[c.Root] = tier
		
		// Debug: Log some combinations to verify loading
		if c.Root == "Lava" || c.Root == "lava" {
			log.Printf("Loaded Lava combination: %s = %s + %s (tier: %d)", 
				c.Root, c.Left, c.Right, tier)
		}
	}
	
	// Debug: Print some statistics
	log.Printf("Loaded combinations for %d unique elements", len(combinations))
	
	// Debug: Check if Lava exists in combinations
	if combos, exists := combinations["Lava"]; exists {
		log.Printf("Found %d combinations for Lava", len(combos))
		for _, c := range combos {
			log.Printf("Lava combination: %s = %s + %s (tier: %d)", 
				c.Root, c.Left, c.Right, tierMap[c.Root])
		}
	} else {
		log.Printf("No combinations found for Lava")
	}
	
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
	log.Printf("BFS: Starting search for %s", target)
	
	// Find the actual case of the element in our combinations
	var actualElement string
	for root := range combinations {
		if strings.EqualFold(root, target) {
			actualElement = root
			break
		}
	}
	
	if actualElement == "" {
		log.Printf("BFS: No combinations found for %s", target)
		BFSVisitedCount = 0
		return nil
	}
	
	target = actualElement // Use the actual case from combinations
	
	if isBasic(target) {
		log.Printf("BFS: %s is a basic element", target)
		BFSVisitedCount = 1
		return &Node{Element: target}
	}

	// Check if the target exists in our combinations
	if _, exists := combinations[target]; !exists {
		log.Printf("BFS: No combinations found for %s", target)
		BFSVisitedCount = 0
		return nil
	}

	log.Printf("BFS: Found combinations for %s, tier: %d", target, tierMap[target])

	visited := make(map[string]bool)
	recipeMap := make(map[string]*Node)
	queue := []string{target}
	BFSVisitedCount = 0

	log.Printf("BFS: Starting BFS traversal for %s", target)

	// First pass: BFS to find all relevant elements
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		BFSVisitedCount++

		log.Printf("BFS: Visiting %s (visited count: %d, tier: %d)", current, BFSVisitedCount, tierMap[current])

		if isBasic(current) {
			recipeMap[current] = &Node{Element: current}
			continue
		}

		for _, comb := range combinations[current] {
			leftTier := tierMap[comb.Left]
			rightTier := tierMap[comb.Right]
			currentTier := tierMap[current]

			log.Printf("BFS: Checking combination %s = %s(%d) + %s(%d)", 
				current, comb.Left, leftTier, comb.Right, rightTier)

			// More lenient tier comparison
			if leftTier <= currentTier && rightTier <= currentTier {
				if !visited[comb.Left] {
					queue = append(queue, comb.Left)
					log.Printf("BFS: Added %s to queue (tier: %d)", comb.Left, leftTier)
				}
				if !visited[comb.Right] {
					queue = append(queue, comb.Right)
					log.Printf("BFS: Added %s to queue (tier: %d)", comb.Right, rightTier)
				}
			} else {
				log.Printf("BFS: Skipping combination due to tier constraints")
			}
		}
	}

	log.Printf("BFS: Starting recipe building phase")

	// Second pass: Bottom-up build recipes
	changed := true
	for changed {
		changed = false
		for elem := range visited {
			if recipeMap[elem] != nil {
				continue
			}
			for _, comb := range combinations[elem] {
				leftTier := tierMap[comb.Left]
				rightTier := tierMap[comb.Right]
				elemTier := tierMap[elem]

				log.Printf("BFS: Trying to build recipe for %s (tier: %d) from %s(%d) + %s(%d)",
					elem, elemTier, comb.Left, leftTier, comb.Right, rightTier)

				if leftTier <= elemTier && rightTier <= elemTier {
					leftRecipe := recipeMap[comb.Left]
					rightRecipe := recipeMap[comb.Right]
					if leftRecipe != nil && rightRecipe != nil {
						recipeMap[elem] = &Node{
							Element: elem,
							Left:    leftRecipe,
							Right:   rightRecipe,
						}
						changed = true
						log.Printf("BFS: Built recipe for %s", elem)
						break
					} else {
						log.Printf("BFS: Missing recipe for %s or %s", comb.Left, comb.Right)
					}
				} else {
					log.Printf("BFS: Tier constraints not met for %s", elem)
				}
			}
		}
	}

	result := recipeMap[target]
	if result == nil {
		log.Printf("BFS: No recipe found for %s", target)
	} else {
		log.Printf("BFS: Found recipe for %s", target)
	}
	return result
}

func FindRecipeDFS(target string, visited map[string]bool) *Node {
	// Check if the target exists in our combinations
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
	
	// Save original state to restore later
	defer func() { visited[target] = false }()

	for _, comb := range combinations[target] {
		// Allow ingredients of the same tier or lower
		if tierMap[comb.Left] <= tierMap[target] && tierMap[comb.Right] <= tierMap[target] {
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
	var results []*Node
	
	// Reset the visited count
	MultiVisitedCount = 0
	
	// Check if target exists
	if _, exists := combinations[target]; !exists && !isBasic(target) {
		return results
	}
	
	// Early return for basic elements
	if isBasic(target) {
		MultiVisitedCount = 1
		return []*Node{{Element: target}}
	}

	// Use channels to collect results from goroutines
	resultChan := make(chan *Node, maxCount*2)
	var wg sync.WaitGroup
	
	// Create a semaphore to limit concurrent goroutines
	semaphore := make(chan struct{}, 5)
	
	// Track visited nodes globally for the multivisited count
	var visitedCountMutex sync.Mutex
	
	// Find all valid combinations for the target
	validCombos := make([]Combination, 0)
	for _, comb := range combinations[target] {
		if tierMap[comb.Left] <= tierMap[target] && tierMap[comb.Right] <= tierMap[target] {
			validCombos = append(validCombos, comb)
		}
	}
	
	// Process each valid combination
	for _, comb := range validCombos {
		wg.Add(1)
		semaphore <- struct{}{}  // Acquire semaphore
		
		go func(c Combination) {
			defer wg.Done()
			defer func() { <-semaphore }()  // Release semaphore
			
			// Find recipe for left ingredient
			leftVisited := make(map[string]bool)
			left := exploreRecipe(c.Left, leftVisited, &visitedCountMutex)
			if left == nil {
				return
			}
			
			// Find recipe for right ingredient
			rightVisited := make(map[string]bool)
			right := exploreRecipe(c.Right, rightVisited, &visitedCountMutex)
			if right == nil {
				return
			}
			
			// Create the node and send to result channel
			resultChan <- &Node{
				Element: target,
				Left:    left,
				Right:   right,
			}
		}(comb)
	}
	
	// Wait in a separate goroutine and close the channel when done
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// Collect results and ensure uniqueness
	seen := make(map[string]bool)
	for node := range resultChan {
		signature := serializeTree(node)
		if !seen[signature] {
			seen[signature] = true
			results = append(results, node)
			if len(results) >= maxCount {
				break
			}
		}
	}
	
	// Sort results by depth for consistent output
	sort.Slice(results, func(i, j int) bool {
		return treeDepth(results[i]) < treeDepth(results[j])
	})
	
	return results
}

// Helper function to explore a recipe with tracked visited count
func exploreRecipe(target string, visited map[string]bool, mutex *sync.Mutex) *Node {
	if isBasic(target) {
		mutex.Lock()
		MultiVisitedCount++
		mutex.Unlock()
		return &Node{Element: target}
	}
	
	if visited[target] {
		return nil
	}
	
	visited[target] = true
	mutex.Lock()
	MultiVisitedCount++
	mutex.Unlock()
	
	// Store valid candidates to avoid duplicate work
	var validCandidates []*Node
	
	for _, comb := range combinations[target] {
		if tierMap[comb.Left] <= tierMap[target] && tierMap[comb.Right] <= tierMap[target] {
			leftVisited := copyVisitedMap(visited)
			left := exploreRecipe(comb.Left, leftVisited, mutex)
			if left == nil {
				continue
			}
			
			rightVisited := copyVisitedMap(visited)
			right := exploreRecipe(comb.Right, rightVisited, mutex)
			if right == nil {
				continue
			}
			
			validCandidates = append(validCandidates, &Node{
				Element: target,
				Left:    left,
				Right:   right,
			})
		}
	}
	
	if len(validCandidates) == 0 {
		return nil
	}
	
	// Return the simplest candidate (lowest tree depth)
	sort.Slice(validCandidates, func(i, j int) bool {
		return treeDepth(validCandidates[i]) < treeDepth(validCandidates[j])
	})
	
	return validCandidates[0]
}

// Helper to copy a visited map
func copyVisitedMap(original map[string]bool) map[string]bool {
	copy := make(map[string]bool)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// Calculate tree depth
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

// Serialize a tree to a unique string representation
func serializeTree(n *Node) string {
	if n == nil {
		return ""
	}
	if n.Left == nil && n.Right == nil {
		return n.Element
	}
	// Create a deterministic string representation
	leftStr := serializeTree(n.Left)
	rightStr := serializeTree(n.Right)
	
	// Sort children for consistency
	if leftStr > rightStr {
		return n.Element + "(" + rightStr + "," + leftStr + ")"
	}
	return n.Element + "(" + leftStr + "," + rightStr + ")"
}

// Public API
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
	return MultiVisitedCount
}

func main() {
	// Load combinations from file
	err := LoadCombinations("combinations.json")
	if err != nil {
		log.Fatalf("Error loading combinations: %v", err)
	}

	// Set up CORS middleware
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
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

	// Set up routes
	http.HandleFunc("/search", corsMiddleware(handleSearch))

	// Start server
	port := ":5000"
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	element := r.URL.Query().Get("element")
	mode := r.URL.Query().Get("mode")

	log.Printf("Searching for element: %s with mode: %s", element, mode)

	if element == "" {
		http.Error(w, "Element parameter is required", http.StatusBadRequest)
		return
	}

	// Debug: Check if element exists in combinations (case-insensitive)
	elementFound := false
	var elementCombos []Combination
	for root, combos := range combinations {
		if strings.EqualFold(root, element) {
			elementFound = true
			elementCombos = combos
			log.Printf("Found %d combinations for element %s (case-insensitive match with %s)", 
				len(combos), element, root)
			break
		}
	}

	if elementFound {
		for _, c := range elementCombos {
			log.Printf("Combination: %s = %s + %s (tier: %s)", c.Root, c.Left, c.Right, c.Tier)
		}
	} else {
		log.Printf("No combinations found for element %s", element)
	}

	var response SearchResponse
	var paths [][]Step

	switch mode {
	case "bfs":
		log.Printf("Using BFS search")
		node := FindRecipeBFS(element)
		if node != nil {
			response.Found = true
			response.Steps = BFSVisitedCount
			paths = append(paths, nodeToPath(node))
			log.Printf("BFS found recipe with %d steps", BFSVisitedCount)
		} else {
			log.Printf("BFS did not find recipe")
		}
	case "dfs":
		log.Printf("Using DFS search")
		node := FindRecipeDFS(element, nil)
		if node != nil {
			response.Found = true
			response.Steps = DFSVisitedCount
			paths = append(paths, nodeToPath(node))
			log.Printf("DFS found recipe with %d steps", DFSVisitedCount)
		} else {
			log.Printf("DFS did not find recipe")
		}
	case "multi":
		log.Printf("Using Multi-Recipe search")
		nodes := FindMultipleRecipes(element, 5) // Limit to 5 recipes
		if len(nodes) > 0 {
			response.Found = true
			response.Steps = MultiVisitedCount
			for _, node := range nodes {
				paths = append(paths, nodeToPath(node))
			}
			log.Printf("Multi-Recipe found %d recipes with %d steps", len(nodes), MultiVisitedCount)
		} else {
			log.Printf("Multi-Recipe did not find any recipes")
		}
	default:
		http.Error(w, "Invalid search mode", http.StatusBadRequest)
		return
	}

	response.Paths = paths

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func nodeToPath(node *Node) []Step {
	if node == nil {
		return nil
	}

	var path []Step
	if node.Left != nil && node.Right != nil {
		// Add paths from children first
		leftPath := nodeToPath(node.Left)
		rightPath := nodeToPath(node.Right)
		path = append(path, leftPath...)
		path = append(path, rightPath...)

		// Add this node's step
		path = append(path, Step{
			Ingredients: []string{node.Left.Element, node.Right.Element},
			Result:      node.Element,
		})
	}

	return path
}