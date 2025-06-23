package models

type Filter struct {
	Name    string
	Options []string
}

type HomePageData struct {
	Title            string
	FeaturedProducts []EnrichedProduct
	Categories       []Category
	PageTemplate     string
}

type CategoryPageData struct {
	Title        string
	CategoryId   string
	CategoryName string
	Products     []EnrichedProduct
	Filters      []Filter
	PageTemplate string
}

type ProductPageData struct {
	Title        string
	Product      EnrichedProduct
	Questions    []Question
	Reviews      []Review
	PageTemplate string
}

type CheckoutPageData struct {
	Title        string
	Product      EnrichedProduct
	PageTemplate string
}
