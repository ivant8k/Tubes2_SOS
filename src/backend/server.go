package main

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"sync/atomic"
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

	if _, exists := combinations[target]; !exists {
		BFSVisitedCount = 0
		return nil
	}

	visited := make(map[string]bool)
	recipeMap := make(map[string]*Node)
	queue := []string{target}
	BFSVisitedCount = 0

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
	semaphore := make(chan struct{}, 10)
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

	found := make(chan struct{})

	for _, comb := range validCombos {
		wg.Add(1)
		semaphore <- struct{}{}

		go func(c Combination) {
			defer wg.Done()
			defer func() { <-semaphore }()
			select {
			case <-found:
				return
			default:
			}
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
			close(found)
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