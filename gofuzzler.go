package main

import (
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"sync"

	"github.com/bbalet/stopwords"
	"github.com/fluhus/gostuff/nlp/wordnet"
)

var (
	wg sync.WaitGroup
)

func checkCewl() string {
	path, err := exec.LookPath("ls")
	if err != nil {
		fmt.Println("[-] cewl is not installed. Please run:")
		fmt.Println("sudo apt install cewl")
		panic(err)
	}
	return path
}

// Goroutine task for obtaining synonmys
func routineGetSyns(wn *wordnet.WordNet, tasks <-chan string, results chan<- string) {
	defer wg.Done()
	for word := range tasks {
		wordsSlice := []string{}
		wordsSliceUnique := []string{}
		synsForWord := wn.SearchRanked(word)
		for _, syn := range synsForWord {
			for i := range syn {
				if len(syn[i].Word) != 0 {
					wordsSlice = append(wordsSlice, syn[i].Word...)
				}
			}

		}
		for _, word := range wordsSlice {
			if !slices.Contains(wordsSliceUnique, word) {
				wordsSliceUnique = append(wordsSliceUnique, word)
			}
		}

		if len(wordsSliceUnique) > 10 {
			for i := range 10 {
				results <- wordsSliceUnique[i]
			}
		} else {
			for i := range wordsSliceUnique {
				results <- wordsSliceUnique[i]
			}
		}
	}
}

// Get the synonyms for each word using wordnet
func GetSyns(allWords []string) []string {
	var synonyms []string
	tasks := make(chan string, len(allWords))
	results := make(chan string)

	wn, err := wordnet.Parse("dict/")
	if err != nil {
		panic(err)
	}

	workers := 50
	for range workers {
		wg.Add(1)
		go routineGetSyns(wn, tasks, results)
	}

	for _, word := range allWords {
		tasks <- word
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		if r != "" {
			synonyms = append(synonyms, r)
		}
	}
	return synonyms
}

// Get words from the site using cewl command-line tool
// TODO: implement custom cewl function in Go
func GetWords(cewlPath string) []string {
	var site string
	var allWords []string

	// Get target site from stdin
	if l := len(os.Args); l == 1 || l > 2 {
		fmt.Println("[-] Please specify a target")
		fmt.Println("--> ./go-fuzzler https://example.com")
		os.Exit(1)
	}

	site = os.Args[1]

	// Run CeWL on the target site with depth=0. Otherwise the output file will be ginormous
	cmd := exec.Command(cewlPath, "-d", "0", "--lowercase", "-w", "cewl.txt", site)
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error running cewl:", err)
		panic(err)
	}

	data, err := os.ReadFile("cewl.txt")
	if err != nil {
		fmt.Println("Error reading file:", err)
		panic(err)
	}

	// Remove stopwords
	data = stopwords.Clean(data, "en", false)

	fileContent := string(data)
	fileWordsPre := strings.Split(fileContent, " ")
	// Only take the first 300 words to keep the size down
	if len(fileWordsPre) > 300 {
		allWords = fileWordsPre[0:300]
	} else {
		allWords = fileWordsPre
	}
	// Return the 300 words/synonyms
	return allWords
}

func GetUniqueWords(oldSlice []string) []string {
	var newSlice []string
	for _, word := range oldSlice {
		if !slices.Contains(newSlice, word) {
			newSlice = append(newSlice, word)
		}
	}
	// Return unique string slice
	return newSlice
}

// Goroutine to manipulate each word
func routineFuzzWords(tasks <-chan string, results chan<- string) {
	defer wg.Done()
	for word := range tasks {
		results <- strings.ToLower(word)                                                                                                                                      // Lowercase
		results <- strings.ToUpper(word)                                                                                                                                      // Uppercase
		results <- strings.ToUpper(strings.Split(word, "")[0]) + strings.Join(strings.Split(word, "")[1:], "")                                                                // Capitalize first letter
		results <- strings.Join(strings.Split(word, "")[0:len(word)-1], "") + strings.ToUpper(strings.Split(word, "")[len(word)-1])                                           // Capitalize last letter
		results <- strings.Split(word, "")[0] + strings.ToUpper(strings.Join(strings.Split(word, "")[1:], ""))                                                                // Capitalize all but first letter
		results <- strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(word, "a", "4"), "e", "3"), "l", "1"), "t", "7"), "o", "0") // 1337 speak

		reverseWord := []string{}
		for l := len(strings.Split(word, "")) - 1; l >= 0; l-- {
			reverseWord = append(reverseWord, strings.Split(word, "")[l])
		}
		results <- strings.Join(reverseWord, "")                                                         // Reversed lowercase
		results <- strings.ToUpper(strings.Join(reverseWord, ""))                                        // Reversed uppercase
		results <- strings.ToUpper(reverseWord[0]) + strings.Join(reverseWord[1:], "")                   // Reversed first letter capital
		results <- reverseWord[0] + strings.ToUpper(strings.Join(reverseWord[1:len(reverseWord)-1], "")) // Reversed all but first letter capital

		// Append & prepend digits 0-2050
		for i := range 2051 {
			results <- word + strconv.Itoa(i)
			results <- strconv.Itoa(i) + word
		}

		// Append & prepend special characters
		specs := []string{"!", "@", "#", "$"}
		for i := range len(specs) {
			results <- word + strings.Join(specs[0:i+1], "")
			results <- strings.Join(specs[0:i+1], "") + word
		}
	}
}

// Create worker pool and start fuzzing words
func FuzzWords(words []string) []string {
	var fuzzedWords []string
	tasks := make(chan string, len(words))
	results := make(chan string)

	// Worker pool
	workers := 50
	for range workers {
		wg.Add(1)
		go routineFuzzWords(tasks, results)
	}

	// Add words to the tasks channel to be picked up by workers
	for _, word := range words {
		tasks <- word
	}
	close(tasks)

	go func() {
		wg.Wait()
		close(results)
	}()

	// Get the fuzzed results from the results channel
	for r := range results {
		if r != "" {
			fuzzedWords = append(fuzzedWords, r)
		}
	}
	return fuzzedWords
}

// Write the final list to file
func WriteFile(words []string) {
	outputStr := strings.Join(words, "\n")
	outputBytes := []byte(outputStr)
	err := os.WriteFile("gofuzzler.txt", outputBytes, 0644)
	if err != nil {
		panic(err)
	}
	err = os.Remove("./cewl.txt")
	if err != nil {
		fmt.Println("Failed to delete cewl.txt")
		panic(err)
	}
}

func main() {
	// Check for cewl command
	cewlPath := checkCewl()
	allWords := GetWords(cewlPath)
	fmt.Printf("[+] Words from website: %d\n", len(allWords))
	syns := GetSyns(allWords)
	uniqueSyns := GetUniqueWords(syns)
	fmt.Printf("[+] After adding synonyms : %d\n", len(uniqueSyns))
	fuzzedWords := FuzzWords(uniqueSyns)
	fmt.Printf("[+] After fuzzing : %d\n", len(fuzzedWords))
	WriteFile(fuzzedWords)
	fmt.Printf("[>] Wordlist saved to ./gofuzzler.txt\n")
}
