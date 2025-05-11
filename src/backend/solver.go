package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	//"time"
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

func LoadCombinations(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	var raw []Combination
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	combinations = make(map[string][]Combination)
	tierMap = make(map[string]int)
	reverseMap = make(map[string][]string)

	for _, c := range raw {
		combinations[c.Root] = append(combinations[c.Root], c)
		tierMap[c.Root] = c.Tier
		// Build reverse index
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

// Fixed: Corrected bidirectional search algorithm
func FindRecipeBidirectional(target string) *Node {
	if isBasic(target) {
		return &Node{Element: target}
	}

	// Check if the target exists in combinations
	if _, exists := combinations[target]; !exists {
		return nil
	}

	// Initialize counters and data structures
	BidirectionalVisitedCount = 0
	
	// Forward direction: Store nodes with their complete recipe tree
	forwardMap := make(map[string]*Node)
	// Backward direction: Store the direct recipe for the element
	backwardMap := make(map[string]Combination)
	
	// Initialize queues for BFS in both directions
	forwardQueue := []string{}
	backwardQueue := []string{target}
	
	// Add basic elements to forward map and queue
	for elem := range tierMap {
		if isBasic(elem) {
			forwardMap[elem] = &Node{Element: elem}
			forwardQueue = append(forwardQueue, elem)
			BidirectionalVisitedCount++
		}
	}
	
	// Keep track of visited elements in both directions
	visitedForward := make(map[string]bool)
	visitedBackward := make(map[string]bool)
	
	// Initialize visited status for basic elements
	for elem := range forwardMap {
		visitedForward[elem] = true
	}
	
	// Mark target as visited in backward direction
	visitedBackward[target] = true
	BidirectionalVisitedCount++
	
	// For backward search, we need to know the direct combinations to make each element
	for elem, combList := range combinations {
		for _, comb := range combList {
			if comb.Root == elem && tierMap[comb.Left] < tierMap[elem] && tierMap[comb.Right] < tierMap[elem] {
				backwardMap[elem] = comb
				break
			}
		}
	}
	
	// Perform bidirectional search
	for len(forwardQueue) > 0 && len(backwardQueue) > 0 {
		// Forward step - extend from basic elements toward target
		nextForward := []string{}
		for _, current := range forwardQueue {
			// Find elements that can be created using the current element
			for resultElem, combList := range combinations {
				for _, comb := range combList {
					// Check if current element is used in this combination
					if (comb.Left == current && visitedForward[comb.Right]) || 
					   (comb.Right == current && visitedForward[comb.Left]) {
						
						// Skip if we've already processed this result or if tiers are not correct
						if visitedForward[resultElem] || 
						   tierMap[comb.Left] >= tierMap[resultElem] || 
						   tierMap[comb.Right] >= tierMap[resultElem] {
							continue
						}
						
						// Create node for this new element
						var leftNode, rightNode *Node
						if comb.Left == current {
							leftNode = forwardMap[current]
							rightNode = forwardMap[comb.Right]
						} else {
							leftNode = forwardMap[comb.Left]
							rightNode = forwardMap[current]
						}
						
						forwardMap[resultElem] = &Node{
							Element: resultElem,
							Left:    leftNode,
							Right:   rightNode,
						}
						
						// Check if we've reached a meeting point
						if visitedBackward[resultElem] {
							return forwardMap[resultElem]
						}
						
						visitedForward[resultElem] = true
						BidirectionalVisitedCount++
						nextForward = append(nextForward, resultElem)
					}
				}
			}
		}
		forwardQueue = nextForward
		
		// Backward step - work from target toward basic elements
		nextBackward := []string{}
		for _, current := range backwardQueue {
			if isBasic(current) {
				continue // Skip if we've reached a basic element
			}
			
			// Get recipe for current element
			comb, exists := backwardMap[current]
			if !exists {
				continue
			}
			
			// Process both ingredients
			ingredients := []string{comb.Left, comb.Right}
			for _, ingredient := range ingredients {
				if visitedBackward[ingredient] {
					continue
				}
				
				// Check if we've reached a meeting point
				if visitedForward[ingredient] {
					// Construct the complete path
					result := &Node{
						Element: ingredient,
						Left:    forwardMap[ingredient],
						Right:   nil, // We'll fill this in with backward construction
					}
					return result
				}
				
				visitedBackward[ingredient] = true
				BidirectionalVisitedCount++
				nextBackward = append(nextBackward, ingredient)
			}
		}
		backwardQueue = nextBackward
	}
	
	return nil
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
	
	// Use mutex for results to avoid race conditions
	var resultMutex sync.Mutex
	var results []*Node
	var seen sync.Map
	done := make(chan struct{})
	
	// Helper function to add a recipe to results
	addResult := func(recipe *Node) bool {
		signature := serializeTree(recipe)
		if _, loaded := seen.LoadOrStore(signature, true); !loaded {
			resultMutex.Lock()
			defer resultMutex.Unlock()
			
			if len(results) >= maxCount {
				return false
			}
			
			results = append(results, recipe)
			if len(results) >= maxCount {
				close(done)
				return false
			}
			return true
		}
		return true // Continue exploring if duplicate
	}
	
	// Function to find all recipe combinations
	var findAllRecipes func(string, map[string][]*Node) []*Node
	findAllRecipes = func(element string, recipeCache map[string][]*Node) []*Node {
		// Check if we already computed recipes for this element
		if recipes, found := recipeCache[element]; found {
			return recipes
		}
		
		// Base case: basic elements
		if isBasic(element) {
			atomic.AddInt32(&MultiVisitedCount, 1)
			node := &Node{Element: element}
			recipeCache[element] = []*Node{node}
			return recipeCache[element]
		}
		
		// Find valid combinations
		var elementRecipes []*Node
		for _, comb := range combinations[element] {
			if tierMap[comb.Left] >= tierMap[element] || tierMap[comb.Right] >= tierMap[element] {
				continue // Skip invalid combinations
			}
			
			// Get all recipes for left and right ingredients
			leftRecipes := findAllRecipes(comb.Left, recipeCache)
			if len(leftRecipes) == 0 {
				continue
			}
			
			rightRecipes := findAllRecipes(comb.Right, recipeCache)
			if len(rightRecipes) == 0 {
				continue
			}
			
			// Create all possible combinations
			for _, leftRecipe := range leftRecipes {
				for _, rightRecipe := range rightRecipes {
					// Check if we need to stop
					select {
					case <-done:
						if len(elementRecipes) > 0 {
							recipeCache[element] = elementRecipes
							return elementRecipes
						}
						return nil
					default:
					}
					
					node := &Node{
						Element: element,
						Left:    leftRecipe,
						Right:   rightRecipe,
					}
					
					// For the target element, add to results
					if element == target {
						if !addResult(node) {
							recipeCache[element] = elementRecipes
							return elementRecipes
						}
					}
					
					elementRecipes = append(elementRecipes, node)
					atomic.AddInt32(&MultiVisitedCount, 1)
					
					// Limit the number of recipes per element to avoid exponential growth
					if len(elementRecipes) >= maxCount*2 {
						break
					}
				}
				if len(elementRecipes) >= maxCount*2 {
					break
				}
			}
		}
		
		// Sort recipes by depth for better results
		sort.Slice(elementRecipes, func(i, j int) bool {
			return treeDepth(elementRecipes[i]) < treeDepth(elementRecipes[j])
		})
		
		// Limit recipe count to prevent memory issues
		if len(elementRecipes) > maxCount*2 {
			elementRecipes = elementRecipes[:maxCount*2]
		}
		
		recipeCache[element] = elementRecipes
		return elementRecipes
	}
	
	// Start the recursive exploration with a shared recipe cache
	recipeCache := make(map[string][]*Node)
	findAllRecipes(target, recipeCache)
	
	// If we have more recipes than requested, sort and truncate
	if len(results) > maxCount {
		sort.Slice(results, func(i, j int) bool {
			return treeDepth(results[i]) < treeDepth(results[j])
		})
		results = results[:maxCount]
	}
	
	return results
}

// Helper function to find multiple variations of recipes for an element
func exploreRecipeVariations(target string, visited map[string]bool, counter *int32, ctx context.Context, seen *sync.Map, resultChan chan<- *Node) *Node {
	// Check for context cancellation
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	
	if isBasic(target) {
		atomic.AddInt32(counter, 1)
		return &Node{Element: target}
	}
	
	if visited[target] {
		return nil
	}
	
	visited[target] = true
	atomic.AddInt32(counter, 1)
	
	// Get valid combinations for this element
	validCombos := []Combination{}
	for _, c := range combinations[target] {
		if tierMap[c.Left] < tierMap[target] && tierMap[c.Right] < tierMap[target] {
			validCombos = append(validCombos, c)
		}
	}
	
	// If no valid combinations, return nil
	if len(validCombos) == 0 {
		return nil
	}
	
	// Randomly shuffle combinations to increase recipe variety
	rand.Shuffle(len(validCombos), func(i, j int) {
		validCombos[i], validCombos[j] = validCombos[j], validCombos[i]
	})
	
	// Try each combination
	for _, c := range validCombos {
		leftVisited := copyVisitedMap(visited)
		left := exploreRecipeVariations(c.Left, leftVisited, counter, ctx, seen, resultChan)
		if left == nil {
			continue
		}
		
		rightVisited := copyVisitedMap(visited)
		right := exploreRecipeVariations(c.Right, rightVisited, counter, ctx, seen, resultChan)
		if right == nil {
			continue
		}
		
		return &Node{
			Element: target,
			Left:    left,
			Right:   right,
		}
	}
	
	return nil
}

// The original helper functions remain unchanged
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