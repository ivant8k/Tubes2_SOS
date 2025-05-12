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

// Struktur Combination merepresentasikan satu kombinasi untuk membentuk elemen baru
// Root adalah elemen hasil, Left dan Right adalah bahan pembentuknya, Tier adalah tingkatannya
type Combination struct {
	Root  string `json:"root"`
	Left  string `json:"left"`
	Right string `json:"right"`
	Tier  int    `json:"tier,string"`
}

// Struktur Node adalah representasi dari node dalam pohon resep
// Element adalah nama elemen, Left dan Right adalah subtree bahan pembentuk
type Node struct {
	Element string
	Left    *Node
	Right   *Node
}

// Struktur RecipePath menyimpan jalur bahan untuk bidirectional search
type RecipePath struct {
	Left  string
	Right string
}

// [MAP] combinations menyimpan semua kombinasi elemen berdasarkan target
var combinations map[string][]Combination

// [MAP] tierMap menyimpan tingkatan (tier) dari setiap elemen
var tierMap map[string]int

// [COUNTERS] Variabel untuk menghitung jumlah node yang dikunjungi di setiap metode
var BFSVisitedCount int
var MultiVisitedCount int32  // Menggunakan atomic untuk memastikan thread-safe dalam multithreading
var DFSVisitedCount int
var BidirectionalVisitedCount int

// [MAP] reverseMap menyimpan hubungan kebalikan dari kombinasi untuk bidirectional search
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

// Fungsi isBasic memeriksa apakah elemen adalah elemen dasar (starter element)
// Elemen dasar adalah elemen yang tidak membutuhkan kombinasi untuk dibuat
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

// Fungsi FindRecipeBidirectional bertujuan mencari resep menggunakan pencarian dua arah (dari elemen dasar dan dari target)
// Metode ini memungkinkan pencarian lebih cepat dengan mempertemukan dua pencarian di tengah
func FindRecipeBidirectional(target string) *Node {
	// [PERCABANGAN] Jika target adalah elemen dasar, langsung kembalikan
	if isBasic(target) {
		return &Node{Element: target}
	}

	// [PERCABANGAN] Jika target tidak ada di kombinasi, tidak ada resep
	if _, exists := combinations[target]; !exists {
		return nil
	}

    // [INISIALISASI] Forward BFS dari semua elemen dasar
	forwardVisited := make(map[string]*Node)
	queue := []string{}
	visited := make(map[string]bool)
	BidirectionalVisitedCount = 0
	// [PERULANGAN] Memasukkan semua elemen dasar sebagai titik awal
	for elem := range tierMap {
		if isBasic(elem) {
			forwardVisited[elem] = &Node{Element: elem}
			queue = append(queue, elem)
			visited[elem] = true
			BidirectionalVisitedCount++
		}
	}
	// [FORWARD BFS] Melakukan pencarian dari elemen dasar ke target
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
	// [PERULANGAN] Telusuri semua kombinasi
	for _, comb := range combinations {
		for _, c := range comb {
			// [PERCABANGAN] Validasi bahan dan memastikan kedua bahan tersedia
			if (c.Left == current || c.Right == current) &&
				forwardVisited[c.Left] != nil && forwardVisited[c.Right] != nil &&
				tierMap[c.Left] < tierMap[c.Root] && tierMap[c.Right] < tierMap[c.Root] {
				
					// [PERCABANGAN] Jika node belum ada, buat baru
				if _, exists := forwardVisited[c.Root]; !exists {
					forwardVisited[c.Root] = &Node{
						Element: c.Root,
						Left:    forwardVisited[c.Left],
						Right:   forwardVisited[c.Right],
					}
					// Tambahkan ke antrian BFS
					queue = append(queue, c.Root)
					visited[c.Root] = true
					BidirectionalVisitedCount++

					// [PERCABANGAN] Jika sudah menemukan target, kembalikan hasil
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

// Fungsi FindRecipeBFS bertujuan mencari satu resep untuk membentuk elemen target menggunakan pencarian BFS.
// Pendekatan BFS akan menjelajahi elemen-elemen yang diperlukan untuk membentuk target secara bertahap dari atas ke bawah.
func FindRecipeBFS(target string) *Node {
	fmt.Printf("\n=== Starting BFS search for: %s ===\n", target)
	// [PERCABANGAN] Jika elemen target adalah elemen dasar, langsung kembalikan node terminal.
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
	// [FORWARD BFS PASS] Mengeksekusi BFS untuk menjelajahi semua komponen yang bisa digunakan membentuk elemen target
	for len(queue) > 0 {
		// Ambil elemen pertama dari queue
		current := queue[0]
		queue = queue[1:]
		
		// [PERCABANGAN] Jika elemen sudah dikunjungi, skip
		if visited[current] {
			continue
		}
		visited[current] = true
		BFSVisitedCount++
		fmt.Printf("Visiting: %s (visited count: %d)\n", current, BFSVisitedCount)
		// [PERCABANGAN] Jika elemen adalah elemen dasar, langsung masukkan ke recipeMap
		if isBasic(current) { 
			fmt.Printf("Found basic element: %s\n", current)
			recipeMap[current] = &Node{Element: current}
			continue
		}
		// [FOR] Telusuri semua kombinasi dari current
		for _, comb := range combinations[current] {
			// [PERCABANGAN] Hanya kombinasikan jika kedua bahan tier-nya lebih rendah
			if tierMap[comb.Left] < tierMap[current] && tierMap[comb.Right] < tierMap[current] {
				fmt.Printf("  Checking combination: %s (tier %d) + %s (tier %d) = %s (tier %d)\n", 
				comb.Left, tierMap[comb.Left], comb.Right, tierMap[comb.Right], comb.Root, comb.Tier)
			
					// Tambahkan ke antrian jika belum dikunjungi
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

	// [PERULANGAN] Bangun pohon resep dari bawah ke atas selama ada perubahan
	for changed {
		changed = false
		for elem := range visited {
			// [PERCABANGAN] Jika elemen sudah punya tree, skip
			if recipeMap[elem] != nil {
				continue
			}
			// [PERULANGAN] Coba semua kombinasi untuk membentuk elem
			for _, comb := range combinations[elem] {
				// [PERCABANGAN] Validasi bahan tier
				if tierMap[comb.Left] < tierMap[elem] && tierMap[comb.Right] < tierMap[elem] {
					leftRecipe := recipeMap[comb.Left]
					rightRecipe := recipeMap[comb.Right]
				// [PERCABANGAN] Jika kedua bahan sudah tersedia, buat node
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

// Fungsi FindRecipeDFS menggunakan pencarian DFS (Depth-First Search) untuk mencari satu resep
// DFS mengeksplorasi hingga kedalaman penuh untuk setiap cabang sebelum kembali ke cabang lain
func FindRecipeDFS(target string, visited map[string]bool) *Node {
	// [PERCABANGAN] Jika target bukan elemen dasar dan tidak memiliki kombinasi, tidak ada resep yang valid
	if _, exists := combinations[target]; !exists && !isBasic(target) {
		return nil
	}

	// [PERCABANGAN] Jika elemen dasar ditemukan, buat node untuk elemen tersebut
	if isBasic(target) {
		DFSVisitedCount++
		return &Node{Element: target}
	}

    // [INISIALISASI] Jika peta visited belum ada, buat baru
	if visited == nil {
		visited = make(map[string]bool)
		DFSVisitedCount = 0
	}
	// [PERCABANGAN] Jika sudah dikunjungi, hentikan pencarian
	if visited[target] {
		return nil
	}

    // Tandai target sebagai dikunjungi
	visited[target] = true
	DFSVisitedCount++

    // [DEFER] Bersihkan penandaan ketika rekursi kembali
	defer func() { visited[target] = false }()

    // [REKURSIF] Telusuri semua kombinasi untuk target
	for _, comb := range combinations[target] {
        // [PERCABANGAN] Validasi apakah kombinasi lebih rendah dari target
		if tierMap[comb.Left] < tierMap[target] && tierMap[comb.Right] < tierMap[target] {
			// [REKURSIF] Cari resep untuk komponen kiri
			left := FindRecipeDFS(comb.Left, visited)
			if left == nil {
				continue
			}
            // [REKURSIF] Cari resep untuk komponen kanan
			right := FindRecipeDFS(comb.Right, visited)
			if right != nil {
				return &Node{Element: target, Left: left, Right: right}
			}
		}
	}
	// [KELUARAN] Jika tidak ada kombinasi yang valid ditemukan
	return nil
}

// Fungsi FindMultipleRecipes bertujuan untuk menemukan beberapa resep unik untuk membentuk elemen target
// Menggunakan multithreading dengan context dan kontrol timeout untuk efisiensi
func FindMultipleRecipes(target string, maxCount int) []*Node {
    // [PERCABANGAN] Jika elemen target adalah elemen dasar, langsung kembalikan
	if isBasic(target) {
		atomic.StoreInt32(&MultiVisitedCount, 1)
		return []*Node{{Element: target}}
	}

	// [PERCABANGAN] Jika target tidak ada di kombinasi, tidak ada resep
	if _, exists := combinations[target]; !exists {
		return nil
	}
	atomic.StoreInt32(&MultiVisitedCount, 0)
	var results []*Node
	var mu sync.Mutex
	var wg sync.WaitGroup

	// [SYNC.MAP] Cache lokal untuk menyimpan hasil resep
	recipeCache := sync.Map{}
	seen := sync.Map{}

	// [CONTEXT] Membatasi waktu pencarian menjadi 5 detik
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Fungsi pembantu untuk memilih kombinasi yang paling optimal
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
	// Fungsi rekursif untuk membangun resep
	var findRecipe func(ctx context.Context, elem string, depth int, targetDepth int) []*Node
	findRecipe = func(ctx context.Context, elem string, depth int, targetDepth int) []*Node {
		// [CONTEXT] Cek timeout
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// [PERCABANGAN] Jika elemen dasar, langsung kembalikan
		if isBasic(elem) {
			atomic.AddInt32(&MultiVisitedCount, 1)
			return []*Node{{Element: elem}}
		}

		// [BATAS DEPTH] Jika kedalaman sudah melebihi target, hentikan
		if depth > targetDepth {
			return nil
		}

		// [CACHE] Jika sudah ada di cache, gunakan
		if cached, ok := recipeCache.Load(elem); ok {
			return cached.([]*Node)
		}
		// Kumpulkan kombinasi yang valid
		validCombos := []Combination{}
		for _, c := range combinations[elem] {
			if tierMap[c.Left] < tierMap[elem] && tierMap[c.Right] < tierMap[elem] {
				validCombos = append(validCombos, c)
			}
		}
		// [OPTIMASI] Hanya gunakan kombinasi terbaik
		if len(validCombos) > 3 && tierMap[elem] > 3 {
			validCombos = selectCombination(validCombos, tierMap[elem])
		}
		var localResults []*Node
		// [PERULANGAN] Telusuri semua kombinasi
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
				maxPairs = 3 }
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
					// [PERCABANGAN] Hanya simpan resep unik	
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
		// [CACHE] Simpan hasil ke cache untuk optimasi
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
	// [START] Mulai proses pencarian dengan depth terbatas
	initialDepth := 12
	if tierMap[target] > 6 {
		initialDepth = 8 } else if tierMap[target] > 8 {
		initialDepth = 6 }
		wg.Add(1)
	go func() {
		defer wg.Done()
		findRecipe(ctx, target, 0, initialDepth)
	}()
		wg.Wait()
	// [SORTING] Urutkan hasil berdasarkan kedalaman
	if len(results) > 0 {
		sort.Slice(results, func(i, j int) bool {
			return treeDepth(results[i]) < treeDepth(results[j])
		})
	}
	return results
}

// Fungsi treeDepth mengukur kedalaman maksimum dari pohon resep
func treeDepth(node *Node) int {
	// [PERCABANGAN] Jika node kosong, kedalaman adalah 0
	if node == nil {
		return 0
	}

	// [REKURSI] Hitung kedalaman subtree kiri dan kanan
	leftDepth := treeDepth(node.Left)
	rightDepth := treeDepth(node.Right)

	// [PERCABANGAN] Kembalikan kedalaman terbesar + 1 (untuk root)
	if leftDepth > rightDepth {
		return leftDepth + 1
	}
	return rightDepth + 1
}

// Fungsi serializeTree mengubah pohon menjadi string unik (untuk mendeteksi duplikasi)
func serializeTree(n *Node) string {
	// [PERCABANGAN] Jika node kosong, kembalikan string kosong
	if n == nil {
		return ""
	}

	// [PERCABANGAN] Jika node adalah elemen dasar (leaf), kembalikan nama elemen
	if n.Left == nil && n.Right == nil {
		return n.Element
	}

	// [REKURSI] Serialisasikan subtree kiri dan kanan
	leftStr := serializeTree(n.Left)
	rightStr := serializeTree(n.Right)

	// [PERCABANGAN] Urutkan secara konsisten untuk mencegah duplikat
	if leftStr > rightStr {
		return n.Element + "(" + rightStr + "," + leftStr + ")"
	}
	return n.Element + "(" + leftStr + "," + rightStr + ")"
}

// Fungsi IsBasic untuk memeriksa apakah elemen adalah elemen dasar
func IsBasic(element string) bool {
	return isBasic(element)
}

// Fungsi GetCombinations mengembalikan semua kombinasi dari elemen tertentu
func GetCombinations(element string) []Combination {
	return combinations[element]
}

// Fungsi IsLowerTier memeriksa apakah kombinasi terdiri dari bahan dengan tier lebih rendah
func IsLowerTier(c Combination) bool {
	return tierMap[c.Left] < tierMap[c.Root] && tierMap[c.Right] < tierMap[c.Root]
}

// Fungsi GetBFSVisited mengembalikan jumlah node yang dikunjungi dalam pencarian BFS
func GetBFSVisited() int {
	return BFSVisitedCount
}

// Fungsi GetDFSVisited mengembalikan jumlah node yang dikunjungi dalam pencarian DFS
func GetDFSVisited() int {
	return DFSVisitedCount
}

// Fungsi GetMultiVisited mengembalikan jumlah node yang dikunjungi dalam pencarian Multirecipe
// Atomic digunakan untuk memastikan bahwa pembacaan dan penulisan MultiVisitedCount aman dalam konteks multithreading
func GetMultiVisited() int {
	return int(atomic.LoadInt32(&MultiVisitedCount))
}

// Fungsi GetBidirectionalVisited mengembalikan jumlah node yang dikunjungi dalam pencarian Bidirectional
func GetBidirectionalVisited() int {
	return BidirectionalVisitedCount
}