package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
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
var BidirectionalVisitedCount int
var reverseMap map[string][]string

// Fungsi LoadCombinations bertujuan untuk memuat data kombinasi dari file JSON
// File berisi daftar kombinasi (Root, Left, Right, Tier) untuk setiap elemen
func LoadCombinations(filename string) error {
	// [BACA FILE] Membaca file JSON
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var raw []Combination

	// [UNMARSHAL] Memproses JSON menjadi slice of Combination
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// [INISIALISASI] Map untuk menyimpan kombinasi dan tier dari setiap elemen
	combinations = make(map[string][]Combination)
	tierMap = make(map[string]int)
	reverseMap = make(map[string][]string)  // Map untuk bidirectional search

	// [PERULANGAN] Memasukkan semua kombinasi ke dalam map
	for _, c := range raw {
		combinations[c.Root] = append(combinations[c.Root], c)
		tierMap[c.Root] = c.Tier

		// [REVERSE MAP] Menyimpan hubungan kebalikan untuk pencarian bidirectional
		reverseMap[c.Left] = append(reverseMap[c.Left], c.Root)
		reverseMap[c.Right] = append(reverseMap[c.Right], c.Root)
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

// Struktur RecipePath menyimpan jalur bahan untuk bidirectional search
type RecipePath struct {
	Left  string
	Right string
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

func FindMultipleRecipes(target string, maxCount int, algorithm string) []*Node {
	if isBasic(target) {
		atomic.StoreInt32(&MultiVisitedCount, 1)
		return []*Node{{Element: target}}
	}
	if _, exists := combinations[target]; !exists {
		return nil
	}

	atomic.StoreInt32(&MultiVisitedCount, 0)
	var results []*Node
	resultChan := make(chan *Node, maxCount*16) // Increased buffer size
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 300) // Increased concurrent searches
	var seen sync.Map
	validCombos := []Combination{}

	targetTier := tierMap[target]
	
	// Filter and sort combinations by tier difference and complexity
	for _, c := range combinations[target] {
		if tierMap[c.Left] < targetTier && tierMap[c.Right] < targetTier {
			// Calculate tier difference to prioritize combinations with ingredients closer to target tier
			leftTierDiff := targetTier - tierMap[c.Left]
			rightTierDiff := targetTier - tierMap[c.Right]
			avgTierDiff := (leftTierDiff + rightTierDiff) / 2
			
			// Add tier difference information to the combination
			c.Tier = avgTierDiff
			validCombos = append(validCombos, c)
		}
	}

	// Sort combinations by tier difference and complexity
	sort.SliceStable(validCombos, func(i, j int) bool {
		// First sort by tier difference (prefer combinations with ingredients closer to target tier)
		if validCombos[i].Tier != validCombos[j].Tier {
			return validCombos[i].Tier < validCombos[j].Tier
		}
		// Then sort by complexity
		return len(validCombos[i].Left)+len(validCombos[i].Right) < len(validCombos[j].Left)+len(validCombos[j].Right)
	})

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second) // Increased timeout
	defer cancel()

	// Function to explore a single combination
	exploreCombination := func(c Combination) {
		defer wg.Done()
		defer func() { <-semaphore }()

		// Try variations based on tier difference
		variations := []struct {
			left  string
			right string
		}{
			{c.Left, c.Right},
			{c.Right, c.Left},
		}

		for _, v := range variations {
			// Try different combinations for both left and right sides
			for _, leftComb := range combinations[target] {
				if tierMap[leftComb.Left] < targetTier && tierMap[leftComb.Right] < targetTier {
					visited := make(map[string]bool)
					var left *Node
					
					// Use the specified algorithm to find the left ingredient
					switch algorithm {
					case "bfs":
						left = FindRecipeBFS(v.left)
						atomic.AddInt32(&MultiVisitedCount, int32(GetBFSVisited()))
					case "dfs":
						left = FindRecipeDFS(v.left, visited)
						atomic.AddInt32(&MultiVisitedCount, int32(GetDFSVisited()))
					case "bidirectional":
						// For bidirectional, we'll use the first basic element as start
						left = FindRecipeBidirectional(v.left, "Earth")
						atomic.AddInt32(&MultiVisitedCount, int32(GetBidirectionalVisited()))
					default:
						left = exploreRecipe(v.left, visited, &MultiVisitedCount, algorithm)
					}

					if left == nil {
						continue
					}

					for _, rightComb := range combinations[target] {
						if tierMap[rightComb.Left] < targetTier && tierMap[rightComb.Right] < targetTier {
							rightVisited := copyVisitedMap(visited)
							var right *Node
							
							// Use the specified algorithm to find the right ingredient
							switch algorithm {
							case "bfs":
								right = FindRecipeBFS(v.right)
								atomic.AddInt32(&MultiVisitedCount, int32(GetBFSVisited()))
							case "dfs":
								right = FindRecipeDFS(v.right, rightVisited)
								atomic.AddInt32(&MultiVisitedCount, int32(GetDFSVisited()))
							case "bidirectional":
								// For bidirectional, we'll use the first basic element as start
								right = FindRecipeBidirectional(v.right, "Earth")
								atomic.AddInt32(&MultiVisitedCount, int32(GetBidirectionalVisited()))
							default:
								right = exploreRecipe(v.right, rightVisited, &MultiVisitedCount, algorithm)
							}

							if right == nil {
								continue
							}

							// Try both variations of the tree
							trees := []*Node{
								{Element: target, Left: left, Right: right},
								{Element: target, Left: right, Right: left},
							}

							for _, tree := range trees {
								signature := serializeTree(tree)
								if _, ok := seen.LoadOrStore(signature, true); !ok {
									select {
									case resultChan <- tree:
									case <-ctx.Done():
										return
									}
								}
							}
						}
					}
				}
			}
		}
	}

	// Launch goroutines for each valid combination
	for _, comb := range validCombos {
		select {
		case <-ctx.Done():
			break
		default:
			wg.Add(1)
			semaphore <- struct{}{}
			go exploreCombination(comb)
		}
	}

	// Close result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results with respect to maxCount
	results = make([]*Node, 0, maxCount)
	for recipe := range resultChan {
		results = append(results, recipe)
		if len(results) >= maxCount {
			cancel()
			break
		}
	}

	// If we found more recipes than requested, randomly select maxCount recipes
	if len(results) > maxCount {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(results), func(i, j int) {
			results[i], results[j] = results[j], results[i]
		})
		results = results[:maxCount]
	}

	return results
}

func exploreRecipe(target string, visited map[string]bool, counter *int32, algorithm string) *Node {
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
	validCombos := []Combination{}

	targetTier := tierMap[target]

	// Filter and sort combinations by tier difference
	for _, comb := range combinations[target] {
		if tierMap[comb.Left] < targetTier && tierMap[comb.Right] < targetTier {
			leftTierDiff := targetTier - tierMap[comb.Left]
			rightTierDiff := targetTier - tierMap[comb.Right]
			avgTierDiff := (leftTierDiff + rightTierDiff) / 2
			comb.Tier = avgTierDiff
			validCombos = append(validCombos, comb)
		}
	}

	// Sort combinations by tier difference and complexity
	sort.SliceStable(validCombos, func(i, j int) bool {
		if validCombos[i].Tier != validCombos[j].Tier {
			return validCombos[i].Tier < validCombos[j].Tier
		}
		return len(validCombos[i].Left)+len(validCombos[i].Right) < len(validCombos[j].Left)+len(validCombos[j].Right)
	})

	// Try all valid combinations
	for _, comb := range validCombos {
		variations := []struct {
			left  string
			right string
		}{
			{comb.Left, comb.Right},
			{comb.Right, comb.Left},
		}

		for _, v := range variations {
			leftVisited := copyVisitedMap(visited)
			var left *Node
			
			// Use the specified algorithm to find the left ingredient
			switch algorithm {
			case "bfs":
				left = FindRecipeBFS(v.left)
			case "dfs":
				left = FindRecipeDFS(v.left, leftVisited)
			case "bidirectional":
				// For bidirectional, we'll use the first basic element as start
				left = FindRecipeBidirectional(v.left, "Earth")
			default:
				left = exploreRecipe(v.left, leftVisited, counter, algorithm)
			}

			if left == nil {
				continue
			}

			rightVisited := copyVisitedMap(visited)
			var right *Node
			
			// Use the specified algorithm to find the right ingredient
			switch algorithm {
			case "bfs":
				right = FindRecipeBFS(v.right)
			case "dfs":
				right = FindRecipeDFS(v.right, rightVisited)
			case "bidirectional":
				// For bidirectional, we'll use the first basic element as start
				right = FindRecipeBidirectional(v.right, "Earth")
			default:
				right = exploreRecipe(v.right, rightVisited, counter, algorithm)
			}

			if right == nil {
				continue
			}

			candidates = append(candidates,
				&Node{Element: target, Left: left, Right: right},
				&Node{Element: target, Left: right, Right: left},
			)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	// Return a random candidate to increase variety
	rand.Seed(time.Now().UnixNano())
	return candidates[rand.Intn(len(candidates))]
}

func getSortedBasicElements() []string {
    basics := []string{}
    for elem := range tierMap {
        if isBasic(elem) {
            basics = append(basics, elem)
        }
    }
    sort.Strings(basics)
    return basics
}

// Fungsi FindRecipeBidirectional bertujuan mencari resep menggunakan pencarian dua arah (dari elemen dasar dan dari target)
// Metode ini memungkinkan pencarian lebih cepat dengan mempertemukan dua pencarian di tengah
func FindRecipeBidirectional(target string, startElement string) *Node {
    fmt.Printf("\n=== Starting Bidirectional Search ===\n")
    fmt.Printf("Target: %s (Tier: %d)\n", target, tierMap[target])
    fmt.Printf("Start Element: %s (Tier: %d)\n", startElement, tierMap[startElement])

    if isBasic(target) {
        fmt.Printf("Target is a basic element, returning direct node\n")
        return &Node{Element: target}
    }
    if _, exists := combinations[target]; !exists {
        fmt.Printf("Target element not found in combinations\n")
        return nil
    }
    if !isBasic(startElement) {
        fmt.Printf("Start element is not a basic element\n")
        return nil
    }

    // Forward search (start → target)
    forwardVisited := make(map[string]*Node)
    forwardQueue := []string{startElement}
    forwardVisited[startElement] = &Node{Element: startElement}
    
    // Backward search (target → basic)
    backwardVisited := make(map[string]bool)
    backwardQueue := []string{target}
    backwardVisited[target] = true
    
    BidirectionalVisitedCount = 2 // Start with both elements
    fmt.Printf("Initialized bidirectional search\n")

    // Main search loop
    for len(forwardQueue) > 0 && len(backwardQueue) > 0 {
        // Forward expansion
        currentForward := forwardQueue[0]
        forwardQueue = forwardQueue[1:]
        fmt.Printf("\nForward exploring from: %s (Tier: %d)\n", currentForward, tierMap[currentForward])

        // Check if current forward node is in backward visited
        if backwardVisited[currentForward] {
            fmt.Printf("Found intersection at: %s\n", currentForward)
            // Build path from start to intersection
            forwardPath := forwardVisited[currentForward]
            // Build path from intersection to target
            backwardPath := buildBackwardPath(currentForward, target, backwardVisited)
            if backwardPath != nil {
                return &Node{
                    Element: target,
                    Left:   forwardPath,
                    Right:  backwardPath,
                }
            }
        }

        // Expand forward
        for _, comb := range combinations {
            for _, c := range comb {
                if (c.Left == currentForward || c.Right == currentForward) &&
                    forwardVisited[c.Left] != nil && forwardVisited[c.Right] != nil &&
                    tierMap[c.Left] < tierMap[c.Root] && tierMap[c.Right] < tierMap[c.Root] {
                    
                    if _, exists := forwardVisited[c.Root]; !exists {
                        fmt.Printf("  Forward found: %s + %s = %s\n", c.Left, c.Right, c.Root)
                        forwardVisited[c.Root] = &Node{
                            Element: c.Root,
                            Left:    forwardVisited[c.Left],
                            Right:   forwardVisited[c.Right],
                        }
                        forwardQueue = append(forwardQueue, c.Root)
                        BidirectionalVisitedCount++
                    }
                }
            }
        }

        // Backward expansion
        currentBackward := backwardQueue[0]
        backwardQueue = backwardQueue[1:]
        fmt.Printf("\nBackward exploring from: %s (Tier: %d)\n", currentBackward, tierMap[currentBackward])

        // Check if current backward node is in forward visited
        if node, exists := forwardVisited[currentBackward]; exists {
            fmt.Printf("Found intersection at: %s\n", currentBackward)
            // Build path from start to intersection
            forwardPath := node
            // Build path from intersection to target
            backwardPath := buildBackwardPath(currentBackward, target, backwardVisited)
            if backwardPath != nil {
                return &Node{
                    Element: target,
                    Left:   forwardPath,
                    Right:  backwardPath,
                }
            }
        }

        // Expand backward
        for _, comb := range combinations[currentBackward] {
            for _, ingredient := range []string{comb.Left, comb.Right} {
                if !backwardVisited[ingredient] && tierMap[ingredient] < tierMap[currentBackward] {
                    fmt.Printf("  Backward found ingredient: %s\n", ingredient)
                    backwardVisited[ingredient] = true
                    backwardQueue = append(backwardQueue, ingredient)
                    BidirectionalVisitedCount++
                }
            }
        }
    }

    fmt.Printf("\nNo recipe found from %s to %s\n", startElement, target)
    return nil
}

func buildBackwardPath(from, to string, visited map[string]bool) *Node {
    if from == to {
        return &Node{Element: to}
    }

    // Try to find a valid combination that leads to the target
    for _, comb := range combinations[to] {
        if visited[comb.Left] && visited[comb.Right] &&
           tierMap[comb.Left] < tierMap[to] && tierMap[comb.Right] < tierMap[to] {
            leftNode := buildBackwardPath(from, comb.Left, visited)
            rightNode := buildBackwardPath(from, comb.Right, visited)
            if leftNode != nil && rightNode != nil {
                return &Node{
                    Element: to,
                    Left:   leftNode,
                    Right:  rightNode,
                }
            }
        }
    }

    return nil
}

func expandForwardBuild(queue []string, visited map[string]*Node, otherVisited map[string]bool, intersection *string) ([]string, bool) {
    next := []string{}
    sort.Strings(queue)

    for _, current := range queue {
        // Find all combinations where current is one of the components
        for root, combList := range combinations {
            if _, exists := visited[root]; exists {
                continue
            }

            for _, comb := range combList {
                // Ensure tier is valid and current is involved in this combination
                if (comb.Left == current || comb.Right == current) &&
                   tierMap[comb.Left] < tierMap[root] && 
                   tierMap[comb.Right] < tierMap[root] {
                    
                    leftNode, leftExists := visited[comb.Left]
                    rightNode, rightExists := visited[comb.Right]
                    
                    if leftExists && rightExists {
                        visited[root] = &Node{
                            Element: root,
                            Left:    leftNode,
                            Right:   rightNode,
                        }
                        next = append(next, root)
                        BidirectionalVisitedCount++

                        if otherVisited[root] {
                            *intersection = root
                            return nil, true
                        }
                    }
                }
            }
        }
    }
    return next, false
}

func expandBackwardTrack(queue []string, visited map[string]bool, otherVisited map[string]*Node, intersection *string) ([]string, bool) {
    next := []string{}
    sort.Strings(queue)

    for _, current := range queue {
        for _, comb := range combinations[current] {
            for _, ingredient := range []string{comb.Left, comb.Right} {
                if !visited[ingredient] && tierMap[ingredient] < tierMap[current] {
                    visited[ingredient] = true
                    next = append(next, ingredient)
                    BidirectionalVisitedCount++

                    if _, exists := otherVisited[ingredient]; exists {
                        *intersection = ingredient
                        return nil, true
                    }
                }
            }
        }
    }
    return next, false
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

func GetBidirectionalVisited() int {
	return BidirectionalVisitedCount
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
	recipeMode := r.URL.Query().Get("recipe_mode")
	fmt.Printf("\n=== Search Request ===\n")
	fmt.Printf("Element: %s (Tier: %d)\n", element, tierMap[element])
	fmt.Printf("Mode: %s\n", mode)
	fmt.Printf("Recipe Mode: %s\n", recipeMode)

	var result *Node
	var results []*Node
	var visited int
	var executionTime time.Duration
	var response struct {
		Found         bool     `json:"found"`
		Steps         int      `json:"steps"`
		Paths         [][]Step `json:"paths"`
		Target        struct {
			Element string `json:"element"`
			Tier    int    `json:"tier"`
		} `json:"target"`
		ExecutionTime float64 `json:"executionTime"` // in milliseconds
	}

	// Set target information
	response.Target.Element = element
	response.Target.Tier = tierMap[element]

	startTime := time.Now()

	// Handle single recipe mode
	if recipeMode == "single" {
		switch mode {
		case "bfs":
			result = FindRecipeBFS(element)
			visited = GetBFSVisited()
			if result != nil {
				path := convertRecipeToPath(result)
				response.Found = true
				response.Steps = visited
				response.Paths = [][]Step{path}
			}
		case "dfs":
			result = FindRecipeDFS(element, nil)
			visited = GetDFSVisited()
			if result != nil {
				path := convertRecipeToPath(result)
				response.Found = true
				response.Steps = visited
				response.Paths = [][]Step{path}
			}
		case "bidirectional":
			startElement := r.URL.Query().Get("start_element")
			if startElement == "" {
				// If no start element is provided, use the first basic element
				startElement = "Earth"
			}
			// Check if start element is the same as target element
			if startElement == element {
				http.Error(w, "The target element can't be the same as the starting element", http.StatusBadRequest)
				return
			}
			result = FindRecipeBidirectional(element, startElement)
			visited = BidirectionalVisitedCount
			if result != nil {
				path := convertRecipeToPath(result)
				response.Found = true
				response.Steps = visited
				response.Paths = [][]Step{path}
			}
		default:
			fmt.Printf("Invalid mode: %s\n", mode)
			http.Error(w, "Invalid mode", http.StatusBadRequest)
			return
		}
	} else if recipeMode == "multiple" {
		maxRecipes := 10 // Default value
		if maxRecipesStr := r.URL.Query().Get("max_recipes"); maxRecipesStr != "" {
			if parsed, err := strconv.Atoi(maxRecipesStr); err == nil && parsed > 0 {
				maxRecipes = parsed
			}
		}
		fmt.Printf("Requested max recipes: %d\n", maxRecipes)
		results = FindMultipleRecipes(element, maxRecipes, mode)
		visited = GetMultiVisited()
		if len(results) > 0 {
			paths := make([][]Step, 0, len(results))
			for _, result := range results {
				path := convertRecipeToPath(result)
				paths = append(paths, path)
			}
			response.Found = true
			response.Steps = visited
			response.Paths = paths
		}
	} else {
		http.Error(w, "Invalid recipe mode", http.StatusBadRequest)
		return
	}

	executionTime = time.Since(startTime)
	response.ExecutionTime = float64(executionTime.Microseconds()) / 1000.0

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Encode and send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}
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

	modes := []string{"bfs", "dfs", "bidirectional", "multi"}
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