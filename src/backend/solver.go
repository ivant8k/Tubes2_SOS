package backend

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	//"net/http"
	"os"
	"sort"
	//"strconv"
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

func FindRecipeBidirectional(target string) *Node {
	if isBasic(target) {
		return &Node{Element: target}
	}

	forwardVisited := make(map[string]*Node)
	backwardVisited := make(map[string]*Node)
	queueF := []string{}
	queueB := []string{}
	visitedF := make(map[string]bool)
	visitedB := make(map[string]bool)
	BidirectionalVisitedCount = 0

	// Inisialisasi forward dengan elemen basic
	for e := range tierMap {
		if isBasic(e) {
			forwardVisited[e] = &Node{Element: e}
			queueF = append(queueF, e)
			visitedF[e] = true
			BidirectionalVisitedCount++
		}
	}

	// Inisialisasi backward dari target
	backwardVisited[target] = &Node{Element: target}
	queueB = append(queueB, target)
	visitedB[target] = true
	BidirectionalVisitedCount++

	for len(queueF) > 0 && len(queueB) > 0 {
		// Forward step
		nextF := []string{}
		for _, current := range queueF {
			for _, comb := range combinations {
				for _, c := range comb {
					if (c.Left == current || c.Right == current) &&
						forwardVisited[c.Left] != nil && forwardVisited[c.Right] != nil &&
						tierMap[c.Left] < tierMap[c.Root] && tierMap[c.Right] < tierMap[c.Root] {

						if forwardVisited[c.Root] == nil {
							forwardVisited[c.Root] = &Node{
								Element: c.Root,
								Left:    forwardVisited[c.Left],
								Right:   forwardVisited[c.Right],
							}
							if visitedB[c.Root] {
								// Gabungkan dua pohon di titik temu
								return &Node{
									Element: c.Root,
									Left:    forwardVisited[c.Root],
									Right:   backwardVisited[c.Root],
								}
							}
							visitedF[c.Root] = true
							BidirectionalVisitedCount++
							nextF = append(nextF, c.Root)
						}
					}
				}
			}
		}
		queueF = nextF

		// Backward step
		nextB := []string{}
		for _, current := range queueB {
			for _, parent := range reverseMap[current] {
				if !visitedB[parent] {
					for _, c := range combinations[parent] {
						if c.Left == current || c.Right == current {
							backwardVisited[parent] = &Node{Element: parent}
							if visitedF[parent] {
								return &Node{
									Element: parent,
									Left:    forwardVisited[parent],
									Right:   backwardVisited[parent],
								}
							}
							visitedB[parent] = true
							BidirectionalVisitedCount++
							nextB = append(nextB, parent)
							break
						}
					}
				}
			}
		}
		queueB = nextB
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
	var results []*Node
	resultChan := make(chan *Node, maxCount*8)
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 200)
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
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second) // Increased timeout for high-tier elements
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
					left := exploreRecipe(v.left, visited, &MultiVisitedCount)
					if left == nil {
						continue
					}

					for _, rightComb := range combinations[target] {
						if tierMap[rightComb.Left] < targetTier && tierMap[rightComb.Right] < targetTier {
							rightVisited := copyVisitedMap(visited)
							right := exploreRecipe(v.right, rightVisited, &MultiVisitedCount)
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

	sort.Slice(candidates, func(i, j int) bool {
		return treeDepth(candidates[i]) < treeDepth(candidates[j])
	})

	return candidates[0]
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
