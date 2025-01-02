package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	Read(fileContent []byte) (Recipes, error)
}

type JSONReader struct{}

func (r JSONReader) Read(fileContent []byte) (Recipes, error) {
	var recipes Recipes
	err := json.Unmarshal(fileContent, &recipes)
	return recipes, err
}

type XMLReader struct{}

func (r XMLReader) Read(fileContent []byte) (Recipes, error) {
	var recipes Recipes
	err := xml.Unmarshal(fileContent, &recipes)
	return recipes, err
}

func (r *Recipes) CompareDatabases(newData Recipes) {
	var addedCakes, removedCakes, changedCookingTime, ingredientChanges []string

	cakesInOldData := map[string]Cake{}
	for _, oldCake := range r.Cakes {
		cakesInOldData[oldCake.Name] = oldCake
	}

	for _, newCake := range newData.Cakes {
		if _, exists := cakesInOldData[newCake.Name]; !exists {
			addedCakes = append(addedCakes, fmt.Sprintf("ADDED cake \"%s\"\n", newCake.Name))
		}
	}

	for oldName, oldCake := range cakesInOldData {
		newCake, exists := newData.getCakeByName(oldName)
		if !exists {
			removedCakes = append(removedCakes, fmt.Sprintf("REMOVED cake \"%s\"\n", oldName))
			continue
		}

		if oldCake.StoveTime != newCake.StoveTime {
			changedCookingTime = append(changedCookingTime, fmt.Sprintf("CHANGED cooking time for cake \"%s\" - \"%s\" instead of \"%s\"\n", oldName, newCake.StoveTime, oldCake.StoveTime))
		}

		ingredientChanges = oldCake.compareIngredients(newCake, ingredientChanges)
	}

	allChanges := append(addedCakes, removedCakes...)
	allChanges = append(allChanges, changedCookingTime...)
	allChanges = append(allChanges, ingredientChanges...)

	for _, change := range allChanges {
		fmt.Printf("%s", change)
	}
}

func (r *Recipes) getCakeByName(name string) (Cake, bool) {
	for _, cake := range r.Cakes {
		if cake.Name == name {
			return cake, true
		}
	}
	return Cake{}, false
}

func (c Cake) compareIngredients(newCake Cake, changes []string) []string {
	var addedIngredients, removedIngredients, changedUnits, changedCounts, removedUnits []string

	oldIngredients := make(map[string]Ingredient)
	for _, ingredient := range c.Ingredients {
		oldIngredients[ingredient.Name] = ingredient
	}

	for _, ingredient := range newCake.Ingredients {
		oldIngredient, exists := oldIngredients[ingredient.Name]

		if !exists {
			addedIngredients = append(addedIngredients, fmt.Sprintf("ADDED ingredient \"%s\" for cake \"%s\"\n", ingredient.Name, c.Name))
		} else {
			if oldIngredient.Unit != ingredient.Unit {
				if ingredient.Unit != "" {
					changedUnits = append(changedUnits, fmt.Sprintf("CHANGED unit for ingredient \"%s\" for cake \"%s\" - \"%s\" instead of \"%s\"\n", ingredient.Name, c.Name, ingredient.Unit, oldIngredient.Unit))
				} else {
					removedUnits = append(removedUnits, fmt.Sprintf("REMOVED unit \"%s\" for ingredient \"%s\" for cake \"%s\"\n", oldIngredient.Unit, ingredient.Name, c.Name))
				}
			}
			if oldIngredient.Count != ingredient.Count {
				changedCounts = append(changedCounts, fmt.Sprintf("CHANGED unit count for ingredient \"%s\" for cake \"%s\" - \"%s\" instead of \"%s\"\n", ingredient.Name, c.Name, ingredient.Count, oldIngredient.Count))
			}
		}

		delete(oldIngredients, ingredient.Name)
	}

	for name := range oldIngredients {
		removedIngredients = append(removedIngredients, fmt.Sprintf("REMOVED ingredient \"%s\" for cake \"%s\"\n", name, c.Name))
	}

	ingredientChanges := append(addedIngredients, removedIngredients...)
	ingredientChanges = append(ingredientChanges, changedUnits...)
	ingredientChanges = append(ingredientChanges, changedCounts...)
	ingredientChanges = append(ingredientChanges, removedUnits...)

	return append(changes, ingredientChanges...)
}

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	oldFilename := flag.String("old", "", "Read old file to compare with the new one")
	newFilename := flag.String("new", "", "Read new file to compare with the old one")
	flag.Parse()

	if *oldFilename != "" && *newFilename != "" {
		if filepath.Ext(*oldFilename) != ".xml" {
			log.Fatal("Wrong file extension for old file")
		}
		if filepath.Ext(*newFilename) != ".json" {
			log.Fatal("Wrong file extension for new file")
		}

		oldFileContent, err := os.ReadFile(*oldFilename)
		handleError(err)

		newFileContent, err := os.ReadFile(*newFilename)
		handleError(err)

		var xmlReader DBReader
		var jsonReader DBReader

		xmlReader = XMLReader{}
		jsonReader = JSONReader{}

		xmlRecipes, err := xmlReader.Read(oldFileContent)
		handleError(err)

		jsonRecipes, err := jsonReader.Read(newFileContent)
		handleError(err)

		xmlRecipes.CompareDatabases(jsonRecipes)
	}
}
