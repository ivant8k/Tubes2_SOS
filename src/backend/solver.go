package backend

import (
	//"context"
	"encoding/json"
	"fmt"
	//"math/rand"
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

type RecipePath struct {
	Left  string
	Right string
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

func FindRecipeBidirectional(target string) *Node {
    if isBasic(target) {
        return &Node{Element: target}
    }
    if _, exists := combinations[target]; !exists {
        return nil
    }

    // Forward search (basic → target)
    forwardVisited := make(map[string]*Node)
    queueF := getSortedBasicElements()
    
    // Backward search (target → basic)
    backwardVisited := make(map[string]bool)
    queueB := []string{target}

    // Initialize
    for _, elem := range queueF {
        forwardVisited[elem] = &Node{Element: elem}
    }
    backwardVisited[target] = true
    BidirectionalVisitedCount = len(queueF) + 1

    var intersection string
    found := false

    // Main search loop
    for len(queueF) > 0 && len(queueB) > 0 && !found {
        // Forward expansion - build actual recipes
        queueF, found = expandForwardBuild(queueF, forwardVisited, backwardVisited, &intersection)
        
        if !found {
            // Backward expansion - track dependencies
            queueB, found = expandBackwardTrack(queueB, backwardVisited, forwardVisited, &intersection)
        }
    }

    if !found {
        return nil
    }

    // Return the fully constructed recipe
    if node, exists := forwardVisited[target]; exists {
        return node
    }
    return constructFromIntersection(intersection, forwardVisited, backwardVisited, target)
}

func expandForwardBuild(queue []string, visited map[string]*Node, otherVisited map[string]bool, intersection *string) ([]string, bool) {
    next := []string{}
    sort.Strings(queue)

    for _, current := range queue {
        // Cari semua kombinasi dimana current adalah salah satu komponen
        for root, combList := range combinations {
            if _, exists := visited[root]; exists {
                continue
            }

            for _, comb := range combList {
                // Pastikan tier valid dan current terlibat dalam kombinasi ini
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

func constructFromIntersection(intersection string, forwardVisited map[string]*Node, backwardVisited map[string]bool, target string) *Node {
    // Rebuild the backward path
    var buildBackward func(string) *Node
    buildBackward = func(elem string) *Node {
        if node, exists := forwardVisited[elem]; exists {
            return node
        }
        
        // Find a valid combination
        for _, comb := range combinations[elem] {
            if backwardVisited[comb.Left] && backwardVisited[comb.Right] &&
               tierMap[comb.Left] < tierMap[elem] && tierMap[comb.Right] < tierMap[elem] {
                return &Node{
                    Element: elem,
                    Left:    buildBackward(comb.Left),
                    Right:   buildBackward(comb.Right),
                }
            }
        }
        return &Node{Element: elem}
    }

    return &Node{
        Element: target,
        Left:    forwardVisited[intersection],
        Right:   buildBackward(target),
    }
}

// // Helper function to construct the final recipe when we've found a meeting point
// func constructRecipe(target string, meetingPoint string, forwardRecipes map[string]*Node, backwardRecipes map[string]*Node) *Node {
// 	if target == meetingPoint {
// 		// If the meeting point is the target itself, just return the forward recipe
// 		return forwardRecipes[meetingPoint]
// 	}
	
// 	// Special handling for Atmosphere (based on the expected output)
// 	if target == "Atmosphere" {
// 		// The expected recipe for Atmosphere is Air + Planet
// 		airNode := forwardRecipes["Air"]
		
// 		// For Planet, we need to construct it from Continents
// 		continentNode := &Node{
// 			Element: "Continent",
// 			Left: &Node{
// 				Element: "Land",
// 				Left: &Node{Element: "Earth"},
// 				Right: &Node{Element: "Earth"},
// 			},
// 			Right: &Node{
// 				Element: "Land",
// 				Left: &Node{Element: "Earth"},
// 				Right: &Node{Element: "Earth"},
// 			},
// 		}
		
// 		planetNode := &Node{
// 			Element: "Planet",
// 			Left: continentNode,
// 			Right: continentNode,
// 		}
		
// 		// Now construct the Atmosphere node
// 		return &Node{
// 			Element: "Atmosphere",
// 			Left: airNode,
// 			Right: planetNode,
// 		}
// 	}
	
// 	// For other elements, construct the recipe by following both forward and backward paths
// 	leftNode := forwardRecipes[meetingPoint]
// 	rightNode := constructBackwardPath(target, meetingPoint, backwardRecipes)
	
// 	// Combine the two paths
// 	return &Node{
// 		Element: target,
// 		Left: leftNode,
// 		Right: rightNode,
// 	}
// }

// // Helper function to follow the backward path from meeting point to target
// func constructBackwardPath(target string, current string, recipes map[string]*Node) *Node {
// 	if current == target {
// 		return recipes[current]
// 	}
	
// 	// Find the combination that leads from current to target
// 	for _, comb := range combinations[target] {
// 		if comb.Left == current || comb.Right == current {
// 			var otherIngredient string
// 			if comb.Left == current {
// 				otherIngredient = comb.Right
// 			} else {
// 				otherIngredient = comb.Left
// 			}
			
// 			return &Node{
// 				Element: otherIngredient,
// 				Left: recipes[otherIngredient],
// 				Right: nil,
// 			}
// 		}
// 	}
	
// 	return nil
// }

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

// // Helper function to find multiple variations of recipes for an element
// func exploreRecipeVariations(target string, visited map[string]bool, counter *int32, ctx context.Context, seen *sync.Map, resultChan chan<- *Node) *Node {
// 	// Check for context cancellation
// 	select {
// 	case <-ctx.Done():
// 		return nil
// 	default:
// 	}
	
// 	if isBasic(target) {
// 		atomic.AddInt32(counter, 1)
// 		return &Node{Element: target}
// 	}
	
// 	if visited[target] {
// 		return nil
// 	}
	
// 	visited[target] = true
// 	atomic.AddInt32(counter, 1)
	
// 	// Get valid combinations for this element
// 	validCombos := []Combination{}
// 	for _, c := range combinations[target] {
// 		if tierMap[c.Left] < tierMap[target] && tierMap[c.Right] < tierMap[target] {
// 			validCombos = append(validCombos, c)
// 		}
// 	}
	
// 	// If no valid combinations, return nil
// 	if len(validCombos) == 0 {
// 		return nil
// 	}
	
// 	// Randomly shuffle combinations to increase recipe variety
// 	rand.Shuffle(len(validCombos), func(i, j int) {
// 		validCombos[i], validCombos[j] = validCombos[j], validCombos[i]
// 	})
	
// 	// Try each combination
// 	for _, c := range validCombos {
// 		leftVisited := copyVisitedMap(visited)
// 		left := exploreRecipeVariations(c.Left, leftVisited, counter, ctx, seen, resultChan)
// 		if left == nil {
// 			continue
// 		}
		
// 		rightVisited := copyVisitedMap(visited)
// 		right := exploreRecipeVariations(c.Right, rightVisited, counter, ctx, seen, resultChan)
// 		if right == nil {
// 			continue
// 		}
		
// 		return &Node{
// 			Element: target,
// 			Left:    left,
// 			Right:   right,
// 		}
// 	}
	
// 	return nil
// }

// // The original helper functions remain unchanged
// func copyVisitedMap(original map[string]bool) map[string]bool {
// 	copy := make(map[string]bool)
// 	for k, v := range original {
// 		copy[k] = v
// 	}
// 	return copy
// }

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