package backend
import (
	"context"
	"encoding/json"
	"fmt"
	//"math/rand"
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
	if _, exists := combinations[target]; !exists {
		return nil
	}
	forwardVisited := make(map[string]*Node)
	queue := []string{}
	visited := make(map[string]bool)
	BidirectionalVisitedCount = 0
		for elem := range tierMap {
		if isBasic(elem) {
			forwardVisited[elem] = &Node{Element: elem}
			queue = append(queue, elem)
			visited[elem] = true
			BidirectionalVisitedCount++
		}
	}
		for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, comb := range combinations {
			for _, c := range comb {
				if (c.Left == current || c.Right == current) &&
					forwardVisited[c.Left] != nil && forwardVisited[c.Right] != nil &&
					tierMap[c.Left] < tierMap[c.Root] && tierMap[c.Right] < tierMap[c.Root] {
					if _, exists := forwardVisited[c.Root]; !exists {
						forwardVisited[c.Root] = &Node{
							Element: c.Root,
							Left:    forwardVisited[c.Left],
							Right:   forwardVisited[c.Right],
						}
						queue = append(queue, c.Root)
						visited[c.Root] = true
						BidirectionalVisitedCount++
						if c.Root == target {
							return forwardVisited[c.Root]
						}
					}
				}
			}
		}
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
				for _, comb := range combinations[current] {
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
		fmt.Println("\nSecond pass: Building recipes...")
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
	var mu sync.Mutex
	var wg sync.WaitGroup
	recipeCache := sync.Map{}
	seen := sync.Map{}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
			selectCombination := func(combos []Combination, targetTier int) []Combination {
		if len(combos) <= 3 {
			return combos 		}
				sort.SliceStable(combos, func(i, j int) bool {
			iDiff := (targetTier - tierMap[combos[i].Left]) + (targetTier - tierMap[combos[i].Right])
			jDiff := (targetTier - tierMap[combos[j].Left]) + (targetTier - tierMap[combos[j].Right])
			return iDiff > jDiff
		})
				maxCombs := 3
		if len(combos) < maxCombs {
			maxCombs = len(combos)
		}
		return combos[:maxCombs]
	}
		var findRecipe func(ctx context.Context, elem string, depth int, targetDepth int) []*Node
		findRecipe = func(ctx context.Context, elem string, depth int, targetDepth int) []*Node {
				select {
		case <-ctx.Done():
			return nil
		default:
					}
				if isBasic(elem) {
			atomic.AddInt32(&MultiVisitedCount, 1)
			return []*Node{{Element: elem}}
		}
				if depth > targetDepth {
			return nil
		}
				if cached, ok := recipeCache.Load(elem); ok {
			return cached.([]*Node)
		}
				validCombos := []Combination{}
		for _, c := range combinations[elem] {
			if tierMap[c.Left] < tierMap[elem] && tierMap[c.Right] < tierMap[elem] {
				validCombos = append(validCombos, c)
			}
		}
				if len(validCombos) > 3 && tierMap[elem] > 3 {
			validCombos = selectCombination(validCombos, tierMap[elem])
		}
		var localResults []*Node
				for _, c := range validCombos {
			leftRecipes := findRecipe(ctx, c.Left, depth+1, targetDepth)
			if len(leftRecipes) == 0 {
				continue
			}
			rightRecipes := findRecipe(ctx, c.Right, depth+1, targetDepth)
			if len(rightRecipes) == 0 {
				continue
			}
						maxPairs := 2
			if tierMap[elem] <= 3 {
				maxPairs = 3 			}
			pairCount := 0
						for _, left := range leftRecipes {
				if pairCount >= maxPairs {
					break
				}
				for _, right := range rightRecipes {
					if pairCount >= maxPairs {
						break
					}
					node := &Node{Element: elem, Left: left, Right: right}
										if elem == target {
						signature := serializeTree(node)
						if _, exists := seen.LoadOrStore(signature, true); !exists {
							mu.Lock()
							if len(results) < maxCount {
								results = append(results, node)
							}
							mu.Unlock()
														if len(results) >= maxCount {
								cancel()
								return localResults
							}
						}
					}
					localResults = append(localResults, node)
					atomic.AddInt32(&MultiVisitedCount, 1)
					pairCount++
				}
			}
		}
				if len(localResults) > 0 {
						sort.Slice(localResults, func(i, j int) bool {
				return treeDepth(localResults[i]) < treeDepth(localResults[j])
			})
						maxCacheSize := 5
			if len(localResults) > maxCacheSize {
				localResults = localResults[:maxCacheSize]
			}
			recipeCache.Store(elem, localResults)
		}
		return localResults
	}
		initialDepth := 12
	if tierMap[target] > 6 {
		initialDepth = 8 	} else if tierMap[target] > 8 {
		initialDepth = 6 	}
		wg.Add(1)
	go func() {
		defer wg.Done()
		findRecipe(ctx, target, 0, initialDepth)
	}()
		wg.Wait()
		if len(results) > 0 {
		sort.Slice(results, func(i, j int) bool {
			return treeDepth(results[i]) < treeDepth(results[j])
		})
	}
	return results
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