package models

type Filter struct {
	ID      string
	Name    string
	Options map[string]string
}

type Pagination struct {
	ItemsPerPage int    `json:"items_per_page"`
	HasNext      bool   `json:"has_next"`
	HasPrev      bool   `json:"has_prev"`
	NextCursor   string `json:"next_cursor,omitempty"`
	PrevCursor   string `json:"prev_cursor,omitempty"`
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
	Pagination   Pagination
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
