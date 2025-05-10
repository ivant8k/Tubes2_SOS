package backend
import (
	"encoding/json"
	"os"
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
        "Air": true,
        "Water": true,
        "Fire": true,
    }
    return basics[element]
}

func FindRecipeBFS(target string) *Node {
    if isBasic(target) {
        BFSVisitedCount = 1
        return &Node{Element: target}
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
            if tierMap[comb.Left] < tierMap[comb.Root] && tierMap[comb.Right] < tierMap[comb.Root] {
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
                if tierMap[comb.Left] < tierMap[comb.Root] && tierMap[comb.Right] < tierMap[comb.Root] {
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
    if isBasic(target) {
        DFSVisitedCount++
        return &Node{Element: target}
    }
    if visited[target] {
        return nil
    }
    visited[target] = true
    defer func() { visited[target] = false }() 
    DFSVisitedCount++
    for _, comb := range combinations[target] {
        if tierMap[comb.Left] < tierMap[comb.Root] && tierMap[comb.Right] < tierMap[comb.Root] {
            left := FindRecipeDFS(comb.Left, visited)
            if left == nil {
                continue
            }
            right := FindRecipeDFS(comb.Right, visited)
            if right != nil {
                return &Node{Element: comb.Root, Left: left, Right: right}
            }
        }
    }
    return nil
}

func FindMultipleRecipes(target string, maxCount int) []*Node {
    var results []*Node
    var mu sync.Mutex
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, 10) 
    MultiVisitedCount = 0
    if isBasic(target) {
        return []*Node{{Element: target}}
    }
    validCombos := make([]Combination, 0)
    for _, comb := range combinations[target] {
        if tierMap[comb.Left] < tierMap[comb.Root] && tierMap[comb.Right] < tierMap[comb.Root] {
            validCombos = append(validCombos, comb)
        }
    }
    for _, comb := range validCombos {
        if len(results) >= maxCount {
            break
        }
        wg.Add(1)
        semaphore <- struct{}{}
        go func(c Combination) {
            defer wg.Done()
            defer func() { <-semaphore }()
            visited := make(map[string]bool)
            left := countAndTrackDFS(c.Left, visited)
            if left == nil {
                return
            }
            right := countAndTrackDFS(c.Right, visited)
            if right == nil {
                return
            }
            tree := &Node{
                Element: c.Root,
                Left:    left,
                Right:   right,
            }
            mu.Lock()
            if len(results) < maxCount {
                results = append(results, tree)
            }
            mu.Unlock()
        }(comb)
    }
    wg.Wait()
    return results
}

func countAndTrackDFS(target string, visited map[string]bool) *Node {
    if isBasic(target) {
        incrementMultiVisited()
        return &Node{Element: target}
    }
    if visited[target] {
        return nil
    }
    visited[target] = true
    incrementMultiVisited()
    for _, comb := range combinations[target] {
        if tierMap[comb.Left] < tierMap[comb.Root] && tierMap[comb.Right] < tierMap[comb.Root] {
            left := countAndTrackDFS(comb.Left, visited)
            if left == nil {
                continue
            }
            right := countAndTrackDFS(comb.Right, visited)
            if right != nil {
                return &Node{
                    Element: comb.Root,
                    Left:    left,
                    Right:   right,
                }
            }
        }
    }
    return nil
}

var multiVisitedMutex sync.Mutex
func incrementMultiVisited() {
    multiVisitedMutex.Lock()
    defer multiVisitedMutex.Unlock()
    MultiVisitedCount++
}

func buildTree(c Combination) *Node {
	left := &Node{Element: c.Left}
	right := &Node{Element: c.Right}
	return &Node{Element: c.Root, Left: left, Right: right}
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
	return MultiVisitedCount
}
