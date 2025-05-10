package backend

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
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
var MultiVisitedCount int
var DFSVisitedCount int

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
	for _, c := range raw {
		combinations[c.Root] = append(combinations[c.Root], c)
		tierMap[c.Root] = c.Tier
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
	if isBasic(target) {
		BFSVisitedCount = 1
		return &Node{Element: target}
	}

	// Check if the target exists in our combinations
	if _, exists := combinations[target]; !exists {
		BFSVisitedCount = 0
		return nil
	}

	visited := make(map[string]bool)
	recipeMap := make(map[string]*Node)
	queue := []string{target}
	BFSVisitedCount = 0

	// First pass: BFS to find all relevant elements
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if visited[current] {
			continue
		}
		visited[current] = true
		BFSVisitedCount++

		if isBasic(current) {
			recipeMap[current] = &Node{Element: current}
			continue
		}

		for _, comb := range combinations[current] {
			// Only consider combinations with lower tier ingredients
			if tierMap[comb.Left] < tierMap[current] && tierMap[comb.Right] < tierMap[current] {
				if !visited[comb.Left] {
					queue = append(queue, comb.Left)
				}
				if !visited[comb.Right] {
					queue = append(queue, comb.Right)
				}
			}
		}
	}

	// Second pass: Bottom-up build recipes
	changed := true
	for changed {
		changed = false
		for elem := range visited {
			if recipeMap[elem] != nil {
				continue
			}
			for _, comb := range combinations[elem] {
				if tierMap[comb.Left] < tierMap[elem] && tierMap[comb.Right] < tierMap[elem] {
					leftRecipe := recipeMap[comb.Left]
					rightRecipe := recipeMap[comb.Right]
					if leftRecipe != nil && rightRecipe != nil {
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

	return recipeMap[target]
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