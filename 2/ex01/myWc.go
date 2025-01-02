package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"unicode"
)

func countLines(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return lines, nil
}

func countChars(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	chars := 0
	for scanner.Scan() {
		chars += len(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return chars, nil
}

func countWords(filename string) (int, error) {
	file, err := os.Open(filename)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	words := 0
	for scanner.Scan() {
		line := scanner.Text()
		wordCount := 0
		inWord := false
		for _, r := range line {
			if unicode.IsLetter(r) || unicode.IsNumber(r) {
				if !inWord {
					wordCount++
					inWord = true
				}
			} else {
				inWord = false
			}
		}
		words += wordCount
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return words, nil
}

func main() {
	lineFlag := flag.Bool("l", false, "Count lines")
	charFlag := flag.Bool("m", false, "Count characters")
	wordFlag := flag.Bool("w", false, "Count words")
	flag.Parse()

	files := flag.Args()
	if len(files) == 0 {
		fmt.Println("Please provide file names")
		return
	}

	if *lineFlag && *charFlag || *lineFlag && *wordFlag || *charFlag && *wordFlag {
		fmt.Println("Please specify only one flag")
		return
	}

	ch := make(chan string, len(files))
	for _, file := range files {
		go func(file string) {
			var result int
			var err error
			if *lineFlag {
				result, err = countLines(file)
			} else if *charFlag {
				result, err = countChars(file)
			} else {
				result, err = countWords(file)
			}
			if err != nil {
				ch <- fmt.Sprintf("Error: %v", err)
				return
			}
			ch <- fmt.Sprintf("%d\t%s", result, file)
		}(file)
	}

	for i := 0; i < len(files); i++ {
		fmt.Println(<-ch)
	}
}
