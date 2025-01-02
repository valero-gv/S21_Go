package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Recipes struct {
	Cakes []Cake `json:"cake" xml:"cake"`
}

type Cake struct {
	Name        string       `json:"name" xml:"name"`
	StoveTime   string       `json:"time" xml:"stovetime"`
	Ingredients []Ingredient `json:"ingredients" xml:"ingredients>item"`
}

type Ingredient struct {
	Name  string `json:"ingredient_name" xml:"itemname"`
	Count string `json:"ingredient_count" xml:"itemcount"`
	Unit  string `json:"ingredient_unit" xml:"itemunit"`
}

type DBReader interface {
	Read(filename string) (interface{}, error)
}

type JSONReader struct{}

func (j *JSONReader) Read(filename string) (interface{}, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data interface{}
	err = json.NewDecoder(file).Decode(&data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

type XMLReader struct{}

func (x *XMLReader) Read(filename string) (interface{}, error) {
	content, _ := os.ReadFile(filename)

	var data Recipes
	err := xml.Unmarshal(content, &data)
	return data, err
}

func main() {
	// Parse command-line flags
	filename := flag.String("f", "", "Filename of the database")
	flag.Parse()

	if *filename == "" {
		fmt.Println("Please provide the filename using -f flag")
		os.Exit(1)
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(*filename), "."))

	var reader DBReader
	switch ext {
	case "json":
		reader = &JSONReader{}
	case "xml":
		reader = &XMLReader{}
	default:
		fmt.Println("Unsupported file format")
		os.Exit(1)
	}

	data, err := reader.Read(*filename)
	if err != nil {
		fmt.Printf("Error reading database: %v\n", err)
		os.Exit(1)
	}

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		fmt.Printf("Error formatting data as JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonData))
}
