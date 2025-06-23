package main

import (
	H "mercadillo-global/helpers"
)

// EnrichedProduct extends Product with template-friendly fields
type EnrichedProduct struct {
	Product
	FormattedPrice         string
	FormattedOriginalPrice string
	Discount               int
	Stars                  []int
	RatingInt              int
	// Warehouse-related fields
	AvailableWarehouses []ProductWarehouse `json:"available_warehouses"`
	TotalStock          int                `json:"total_stock"`
	TotalWeight         int                `json:"total_weight"` // peso total en gramos
	GlobalAttributes    []ProductAttribute `json:"global_attributes"`
	ShippingOptions     []ShippingCost     `json:"shipping_options"`
}

// getEnrichedProduct enriches a single product with template data
func getEnrichedProduct(product Product) EnrichedProduct {
	return EnrichedProduct{
		Product:                product,
		FormattedPrice:         H.MaybeFormatNumber(float64(product.Price), true),
		FormattedOriginalPrice: H.MaybeFormatNumber(float64(product.OriginalPrice), true),
		Discount:               calculateDiscount(product.OriginalPrice, product.Price),
		Stars:                  []int{0, 1, 2, 3, 4},
		RatingInt:              int(product.Rating),
	}
}

// getEnrichedProducts enriches a slice of products
func getEnrichedProducts(products []Product) []EnrichedProduct {
	enriched := make([]EnrichedProduct, len(products))
	for i, product := range products {
		enriched[i] = getEnrichedProduct(product)
	}
	return enriched
}

// Note: calculateDiscount function is now in models.go to avoid duplication
