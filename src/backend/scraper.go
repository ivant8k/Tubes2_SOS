package backend

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type CombinationMap map[string]map[string][]string

// Base elements (Tier 0)
var baseElements = []string{"air", "earth", "fire", "water"}

func fetchDocument(url string) *goquery.Document {
	time.Sleep(100 * time.Millisecond)
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("Error fetching %s: %v", url, err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("Status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalf("Error reading body from %s: %v", url, err)
	}
	return doc
}

func ScrapeAll() {
	baseURL := "https://littlealchemy2.gambledude.com"
	mainDoc := fetchDocument(baseURL)

	combinations := make(CombinationMap)
	tierMap := make(map[string]int)
	for _, elem := range baseElements {
		tierMap[elem] = 0
	}

	mainDoc.Find(`li.c-element-list__item[data-val="1"]`).Each(func(i int, s *goquery.Selection) {
		firstIngredient := strings.ToLower(strings.TrimSpace(s.Find("a").Text()))
		link, exists := s.Find("a").Attr("href")
		if !exists || firstIngredient == "" {
			return
		}

		fmt.Printf("[%d] Scraping %s\n", i+1, firstIngredient)
		elementDoc := fetchDocument(link)

		recipes := make(map[string][]string)

		elementDoc.Find("table.o-table--tiny").First().Find("tbody tr").Each(func(j int, row *goquery.Selection) {
			cols := row.Find("a")
			if cols.Length() < 2 {
				return
			}

			secondIngredient := strings.ToLower(strings.TrimSpace(cols.Eq(0).Text()))
			results := []string{}
			cols.Each(func(k int, a *goquery.Selection) {
				if k == 0 {
					return
				}
				result := strings.ToLower(strings.TrimSpace(a.Text()))
				if result != "" {
					results = append(results, result)
				}
			})

			if secondIngredient != "" && len(results) > 0 {
				recipes[secondIngredient] = results
			}
		})

		for _, resultList := range recipes {
			for _, result := range resultList {
				if _, exists := tierMap[result]; !exists {
					tierMap[result] = tierMap[firstIngredient] + 1
				}
			}
		}

		combinations[firstIngredient] = recipes
	})

	file, err := os.Create("combinations.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(combinations)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Selesai! Total elemen discrape: %d\n", len(combinations))
}
