package models

import (
	"encoding/json"
	"fmt"
	"os"
)

type Category struct {
	ID          string               `json:"id,omitempty"`
	Name        string               `json:"name"`
	Children    map[string]*Category `json:"children,omitempty"`
	Attributes  []string             `json:"attributes,omitempty"`
	IsService   bool                 `json:"isService,omitempty"`
	Only18      bool                 `json:"only18,omitempty"`
	KYC         bool                 `json:"kyc,omitempty"`
	OnlyCompany bool                 `json:"onlyCompany,omitempty"`
}

// CategoryFlat struct for displaying flat category lists
type CategoryFlat struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Image    string `json:"image"`
	ParentID string `json:"parent_id,omitempty"`
	Level    int    `json:"level"`
}

// Global category system - loaded once at startup
var (
	categoriesMap    map[string]*Category
	categoriesList   []Category
	categoriesFlat   []CategoryFlat
	categoriesLoaded bool
)

// InitializeCategories loads categories once at startup
func InitializeCategories() error {
	if categoriesLoaded {
		return nil
	}

	file, err := os.Open("categories.json")
	if err != nil {
		return fmt.Errorf("error opening categories.json: %v", err)
	}
	defer file.Close()

	var rawCategories map[string]*Category
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&rawCategories); err != nil {
		return fmt.Errorf("error decoding categories.json: %v", err)
	}

	// Initialize the global map
	categoriesMap = make(map[string]*Category)
	categoriesList = make([]Category, 0)
	categoriesFlat = make([]CategoryFlat, 0)

	// Process categories and build maps
	for id, category := range rawCategories {
		category.ID = id
		processCategoryTree(category, "", 0)
		categoriesList = append(categoriesList, *category)
	}

	categoriesLoaded = true
	return nil
}

// processCategoryTree recursively processes the category tree and builds the flat map
func processCategoryTree(category *Category, parentID string, level int) {
	// Add to global map (O(1) lookup)
	categoriesMap[category.ID] = category

	// Add to flat list for UI purposes
	categoriesFlat = append(categoriesFlat, CategoryFlat{
		ID:       category.ID,
		Name:     category.Name,
		ParentID: parentID,
		Level:    level,
	})

	// Process children
	if category.Children != nil {
		for id, child := range category.Children {
			child.ID = id
			processCategoryTree(child, category.ID, level+1)
		}
	}
}

// GetCategoryByID optimized O(1) lookup using the global map
func GetCategoryByID(categoryID string) *Category {
	return categoriesMap[categoryID]
}

// GetCategories returns the loaded categories list
func GetCategories() []Category {
	return categoriesList
}

// GetFlatCategories returns the pre-built flat categories list
func GetFlatCategories() []CategoryFlat {
	return categoriesFlat
}

// GetCategoryAttributes returns the attributes for a specific category - O(1) lookup
func GetCategoryAttributes(categoryID string) []string {
	category := GetCategoryByID(categoryID)
	if category != nil {
		return category.Attributes
	}
	return []string{}
}

// getCategoryName returns the name of a category by ID
func getCategoryName(categoryID string) string {
	category := GetCategoryByID(categoryID)
	if category != nil {
		return category.Name
	}
	return "Categoría Desconocida"
}
