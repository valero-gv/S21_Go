package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	oldFilename := flag.String("old", "", "Upload old snapshot with paths")
	newFilename := flag.String("new", "", "Upload new snapshot with paths")
	flag.Parse()

	if *oldFilename != "" && *newFilename != "" {
		if filepath.Ext(*oldFilename) != ".txt" {
			log.Fatal("Wrong file extension for old file")
		}
		if filepath.Ext(*newFilename) != ".txt" {
			log.Fatal("Wrong file extension for old file")
		}
		oldFiles := make(map[string]struct{})
		file, err := os.Open(*oldFilename)
		handleError(err)
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			oldFiles[scanner.Text()] = struct{}{}
		}
		file.Close()

		file, err = os.Open(*newFilename)
		handleError(err)
		scanner = bufio.NewScanner(file)
		for scanner.Scan() {
			path := scanner.Text()
			if _, found := oldFiles[path]; !found {
				fmt.Println("ADDED", path)
			} else {
				delete(oldFiles, path)
			}
		}
		file.Close()
		for path := range oldFiles {
			fmt.Println("REMOVED", path)
		}
	}
}
