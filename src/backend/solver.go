package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

// Graph: result → list of ingredient pairs
type Graph map[string][][]string

// InverseGraph: ingredient → ingredient → list of results
type InverseGraph map[string]map[string][]string

// Step merepresentasikan satu langkah kombinasi
type Step struct {
	Ingredients [2]string `json:"ingredients"`
	Result      string    `json:"result"`
}

// LoadGraph memuat kombinasi dari file JSON
func LoadGraph(path string) (Graph, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var graph Graph
	err = json.Unmarshal(file, &graph)
	return graph, err
}

// CreateInverseGraph membuat graph kebalikan untuk pencarian lebih cepat
func CreateInverseGraph(graph Graph) InverseGraph {
	inverse := make(InverseGraph)
	
	for result, combos := range graph {
		for _, combo := range combos {
			a, b := combo[0], combo[1]
			if _, exists := inverse[a]; !exists {
				inverse[a] = make(map[string][]string)
			}
			inverse[a][b] = append(inverse[a][b], result)
			if a != b {
				if _, exists := inverse[b]; !exists {
					inverse[b] = make(map[string][]string)
				}
				inverse[b][a] = append(inverse[b][a], result)
			}
		}
	}
	
	return inverse
}

// ElementInfo menyimpan informasi tentang elemen untuk prioritisasi pencarian
type ElementInfo struct {
	Name        string
	Complexity  int     // Kompleksitas elemen (berapa banyak langkah untuk membuatnya)
	UsageCount  int     // Berapa kali elemen ini digunakan dalam kombinasi
	ResultCount int     // Berapa banyak hasil yang bisa dibuat dengan elemen ini
	Weight      float64 // Bobot prioritas gabungan
}

// CalculateElementInfo menghitung informasi untuk setiap elemen
func CalculateElementInfo(graph Graph, inverse InverseGraph, baseElements []string) map[string]ElementInfo {
	info := make(map[string]ElementInfo)
	
	// Inisialisasi elemen dasar dengan kompleksitas 0
	for _, elem := range baseElements {
		info[elem] = ElementInfo{
			Name:       elem,
			Complexity: 0,
		}
	}
	
	// Hitung berapa kali setiap elemen digunakan
	usageCounts := make(map[string]int)
	for _, combos := range graph {
		for _, combo := range combos {
			usageCounts[combo[0]]++
			usageCounts[combo[1]]++
		}
	}
	
	// Hitung jumlah hasil yang dapat dibuat dengan setiap elemen
	resultCounts := make(map[string]int)
	for elem, pairs := range inverse {
		count := 0
		for _, results := range pairs {
			count += len(results)
		}
		resultCounts[elem] = count
	}
	
	// Tentukan kompleksitas elemen secara topological sort
	// Semua elemen yang tidak ada di info adalah elemen yang dibuat
	toProcess := make([]string, 0)
	for result := range graph {
		if _, exists := info[result]; !exists {
			info[result] = ElementInfo{Name: result, Complexity: -1} // Tandai belum dihitung
			toProcess = append(toProcess, result)
		}
	}
	
	// Urutkan berdasarkan ketergantungan
	for len(toProcess) > 0 {
		progress := false
		for i := 0; i < len(toProcess); i++ {
			elem := toProcess[i]
			combos := graph[elem]
			
			allDependenciesProcessed := true
			maxComplexity := 0
			
			for _, combo := range combos {
				for _, ingredient := range combo {
					if info[ingredient].Complexity == -1 {
						allDependenciesProcessed = false
						break
					}
					if info[ingredient].Complexity > maxComplexity {
						maxComplexity = info[ingredient].Complexity
					}
				}
				if !allDependenciesProcessed {
					break
				}
			}
			
			if allDependenciesProcessed {
				// Update kompleksitas
				elemInfo := info[elem]
				elemInfo.Complexity = maxComplexity + 1
				info[elem] = elemInfo
				
				// Hapus dari daftar yang perlu diproses
				toProcess = append(toProcess[:i], toProcess[i+1:]...)
				i--
				progress = true
			}
		}
		
		if !progress && len(toProcess) > 0 {
			// Ada siklus, tetapkan kompleksitas default
			for _, elem := range toProcess {
				elemInfo := info[elem]
				elemInfo.Complexity = 99 // Nilai tinggi untuk elemen yang sulit dihitung
				info[elem] = elemInfo
			}
			break
		}
	}
	
	// Gabungkan semua informasi dan hitung bobot
	for elem, elemInfo := range info {
		elemInfo.UsageCount = usageCounts[elem]
		elemInfo.ResultCount = resultCounts[elem]
		
		// Rumus bobot: kombinasi kompleksitas, penggunaan, dan hasil
		// Elemen dengan kompleksitas rendah, penggunaan tinggi, dan hasil tinggi lebih diutamakan
		if elemInfo.Complexity > 0 {
			elemInfo.Weight = float64(elemInfo.ResultCount+1) * float64(elemInfo.UsageCount+1) / float64(elemInfo.Complexity*elemInfo.Complexity+1)
		} else {
			elemInfo.Weight = float64(elemInfo.ResultCount+1) * float64(elemInfo.UsageCount+1)
		}
		
		info[elem] = elemInfo
	}
	
	return info
}

// BFSRecipe menemukan jalur kombinasi dari elemen awal ke target menggunakan BFS yang dioptimasi
func BFSRecipe(graph Graph, start []string, target string, timeout time.Duration) ([]Step, bool, int) {
	// Buat inverse graph untuk pencarian lebih cepat
	inverse := CreateInverseGraph(graph)
	
	// Hitung informasi elemen untuk prioritisasi
	elementInfo := CalculateElementInfo(graph, inverse, start)
	
	type State struct {
		Available map[string]bool
		Path      []Step
	}
	
	// Inisialisasi queue dengan elemen awal
	queue := []State{{Available: sliceToSet(start), Path: []Step{}}}
	
	// Gunakan hash string dari elemen yang tersedia sebagai kunci state yang sudah dikunjungi
	visitedStates := make(map[string]bool, 1000)
	visitedStates[stateHash(sliceToSet(start))] = true
	
	nodesVisited := 0
	startTime := time.Now()
	
	// Cek awal apakah target sudah ada di elemen awal
	if queue[0].Available[target] {
		return queue[0].Path, true, nodesVisited
	}
	
	// Bidirectional search - jika target diketahui, cari jalur terpendek dari target juga
	targetPath := make(map[string][]Step)
	targetReachable := make(map[string]bool)
	
	// Tandai elemen yang dapat mencapai target (backward search)
	if _, exists := graph[target]; exists {
		for _, combo := range graph[target] {
			a, b := combo[0], combo[1]
			targetReachable[a] = true
			targetReachable[b] = true
			
			targetPath[a] = []Step{{
				Ingredients: [2]string{a, b},
				Result:      target,
			}}
			targetPath[b] = []Step{{
				Ingredients: [2]string{a, b},
				Result:      target,
			}}
		}
	}
	
	// Tandai timeout
	timeoutTime := startTime.Add(timeout)
	
	// Menentukan jika elemen tertentu adalah milestone
	isMilestone := func(element string) bool {
		// Elemen dengan kompleksitas tinggi adalah milestone
		info, exists := elementInfo[element]
		if !exists {
			return false
		}
		return info.Complexity >= 5 || info.Weight > 50
	}
	
	// Prioritized queue processing
	for len(queue) > 0 {
		// Cek timeout
		if time.Now().After(timeoutTime) {
			fmt.Println("\nTimeout reached! Stopping search.")
			break
		}
		
		curr := queue[0]
		queue = queue[1:]
		nodesVisited++
		
		// Logging progress
		if nodesVisited%10000 == 0 {
			fmt.Printf("Nodes visited: %d, Queue size: %d, Time elapsed: %v\r", 
				nodesVisited, len(queue), time.Since(startTime))
		}
		
		elements := sortElementsByPriority(keys(curr.Available), elementInfo)
		
		// Untuk elemen yang kita ketahui dapat mencapai target langsung
		for _, elem := range elements {
			if targetReachable[elem] {
				// Ada jalur langsung ke target, tambahkan langkah-langkahnya
				directPath := targetPath[elem]
				
				finalPath := make([]Step, len(curr.Path)+len(directPath))
				copy(finalPath, curr.Path)
				copy(finalPath[len(curr.Path):], directPath)
				
				return finalPath, true, nodesVisited
			}
		}
		
		// Terapkan strategi pemangkasan untuk mengurangi cabang
		// Jika kita memiliki lebih dari 20 elemen, fokus pada 20 elemen dengan prioritas tertinggi
		maxElementsToConsider := len(elements)
		if maxElementsToConsider > 20 {
			maxElementsToConsider = 20
		}
		
		for i := 0; i < maxElementsToConsider; i++ {
			a := elements[i]
			
			// Jika a tidak memiliki pasangan di inverse graph, lewati
			if _, exists := inverse[a]; !exists {
				continue
			}
			
			// Mengurangi jumlah pasangan yang diperiksa
			maxPairsToConsider := maxElementsToConsider
			if len(elements) < maxPairsToConsider {
				maxPairsToConsider = len(elements)
			}
			
			for j := i; j < maxPairsToConsider; j++ {
				b := elements[j]
				
				// Cek apakah kombinasi a dan b dapat menghasilkan sesuatu
				var possibleResults []string
				if pairs, exists := inverse[a]; exists {
					if results, hasPair := pairs[b]; hasPair {
						possibleResults = results
					}
				}
				
				for _, result := range possibleResults {
					if curr.Available[result] {
						continue
					}
					
					newAvail := copyMap(curr.Available)
					newAvail[result] = true
					
					isTarget := result == target
					
					step := Step{
						Ingredients: [2]string{a, b},
						Result:      result,
					}
					
					newPath := make([]Step, len(curr.Path)+1)
					copy(newPath, curr.Path)
					newPath[len(curr.Path)] = step
					
					if isTarget {
						return newPath, true, nodesVisited
					}
					
					stateKey := stateHash(newAvail)
					if !visitedStates[stateKey] {
						visitedStates[stateKey] = true
						
						// Prioritaskan state baru jika elemen baru adalah milestone
						if isMilestone(result) {
							// Masukkan di awal queue untuk prioritas tinggi
							newQueue := make([]State, len(queue)+1)
							newQueue[0] = State{
								Available: newAvail,
								Path:      newPath,
							}
							copy(newQueue[1:], queue)
							queue = newQueue
						} else {
							queue = append(queue, State{
								Available: newAvail,
								Path:      newPath,
							})
						}
					}
				}
			}
		}
	}
	
	fmt.Println("\nSearch completed without finding target.")
	return nil, false, nodesVisited
}

// sortElementsByPriority mengurutkan elemen berdasarkan prioritas
func sortElementsByPriority(elements []string, info map[string]ElementInfo) []string {
	sort.Slice(elements, func(i, j int) bool {
		infoI, existsI := info[elements[i]]
		infoJ, existsJ := info[elements[j]]
		
		// Jika informasi tidak tersedia, prioritaskan yang ada informasinya
		if !existsI && existsJ {
			return false
		}
		if existsI && !existsJ {
			return true
		}
		if !existsI && !existsJ {
			return elements[i] < elements[j] // Alphabetic as fallback
		}
		
		// Prioritaskan berdasarkan bobot
		return infoI.Weight > infoJ.Weight
	})
	return elements
}

// Fungsi utilitas

func stateHash(available map[string]bool) string {
	keys := make([]string, 0, len(available))
	for k := range available {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func sliceToSet(slice []string) map[string]bool {
	set := make(map[string]bool, len(slice))
	for _, v := range slice {
		set[v] = true
	}
	return set
}

func keys(m map[string]bool) []string {
	k := make([]string, 0, len(m))
	for key := range m {
		k = append(k, key)
	}
	return k
}

func copyMap(m map[string]bool) map[string]bool {
	newMap := make(map[string]bool, len(m))
	for k, v := range m {
		newMap[k] = v
	}
	return newMap
}

// WriteStepsToJSON menulis langkah-langkah ke file JSON
func WriteStepsToJSON(steps []Step, path string) error {
	data, err := json.MarshalIndent(steps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// PrintSteps mencetak langkah-langkah secara terformat
func PrintSteps(steps []Step) {
	fmt.Println("\nLangkah-langkah untuk membuat target:")
	fmt.Println("-----------------------------------")
	for i, step := range steps {
		fmt.Printf("%3d. %s + %s = %s\n", i+1, step.Ingredients[0], step.Ingredients[1], step.Result)
	}
	fmt.Printf("Total %d langkah\n", len(steps))
}

// GetIntermediateElementStats menghitung statistik elemen perantara
func GetIntermediateElementStats(steps []Step) map[string]int {
	stats := make(map[string]int)
	
	// Hitung berapa kali setiap elemen digunakan
	for _, step := range steps {
		stats[step.Ingredients[0]]++
		stats[step.Ingredients[1]]++
	}
	
	return stats
}

