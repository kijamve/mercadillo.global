package models

import (
	"gorm.io/gorm"
)

// SearchFilters estructura para filtros de búsqueda
type SearchFilters struct {
	Categories   []string `json:"categories"`
	MinPrice     *int     `json:"min_price"`
	MaxPrice     *int     `json:"max_price"`
	MinRating    *float64 `json:"min_rating"`
	IsService    *bool    `json:"is_service"`
	FreeShipping *bool    `json:"free_shipping"`
	Limit        int      `json:"limit"`
	Offset       int      `json:"offset"`
}

// SearchResult estructura para resultados de búsqueda
type SearchResult struct {
	Products   []Product `json:"products"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PerPage    int       `json:"per_page"`
	TotalPages int       `json:"total_pages"`
}

// SearchProducts búsqueda principal usando search_content y search_keywords optimizados
func SearchProducts(db *gorm.DB, query string, filters SearchFilters) (*SearchResult, error) {
	var products []Product
	var total int64

	// Configurar paginación por defecto
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	// Query principal con full-text search optimizado
	baseQuery := db.Model(&Product{}).
		Select("*, MATCH(search_content, search_keywords) AGAINST(? IN NATURAL LANGUAGE MODE) AS relevance", query).
		Where("MATCH(search_content, search_keywords) AGAINST(? IN NATURAL LANGUAGE MODE)", query)

	// Aplicar filtros
	baseQuery = applySearchFilters(baseQuery, filters)

	// Obtener total de resultados
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Si no hay resultados, intentar búsqueda de respaldo
	if total == 0 {
		return SearchProductsBasic(db, query, filters)
	}

	// Obtener productos ordenados por relevancia
	err := baseQuery.
		Order("relevance DESC, rating DESC, sold DESC").
		Limit(filters.Limit).
		Offset(filters.Offset).
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	// Calcular paginación
	page := (filters.Offset / filters.Limit) + 1
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	return &SearchResult{
		Products:   products,
		Total:      total,
		Page:       page,
		PerPage:    filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// SearchProductsBasic búsqueda de respaldo usando title y description
func SearchProductsBasic(db *gorm.DB, query string, filters SearchFilters) (*SearchResult, error) {
	var products []Product
	var total int64

	// Configurar paginación por defecto
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	// Query de respaldo con full-text search básico
	baseQuery := db.Model(&Product{}).
		Select("*, MATCH(title, description) AGAINST(? IN NATURAL LANGUAGE MODE) AS relevance", query).
		Where("MATCH(title, description) AGAINST(? IN NATURAL LANGUAGE MODE)", query)

	// Aplicar filtros
	baseQuery = applySearchFilters(baseQuery, filters)

	// Obtener total de resultados
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Si tampoco hay resultados, buscar por coincidencias parciales en el título
	if total == 0 {
		baseQuery = db.Model(&Product{}).
			Where("title LIKE ?", "%"+query+"%")
		baseQuery = applySearchFilters(baseQuery, filters)
		baseQuery.Count(&total)
	}

	// Obtener productos
	err := baseQuery.
		Order("relevance DESC, rating DESC, sold DESC").
		Limit(filters.Limit).
		Offset(filters.Offset).
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	// Calcular paginación
	page := (filters.Offset / filters.Limit) + 1
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	return &SearchResult{
		Products:   products,
		Total:      total,
		Page:       page,
		PerPage:    filters.Limit,
		TotalPages: totalPages,
	}, nil
}

// applySearchFilters aplica los filtros comunes a las consultas de búsqueda
func applySearchFilters(query *gorm.DB, filters SearchFilters) *gorm.DB {
	// Filtrar por categorías
	if len(filters.Categories) > 0 {
		query = query.Where("category_id IN ?", filters.Categories)
	}

	// Filtrar por rango de precios
	if filters.MinPrice != nil {
		query = query.Where("price >= ?", *filters.MinPrice)
	}
	if filters.MaxPrice != nil {
		query = query.Where("price <= ?", *filters.MaxPrice)
	}

	// Filtrar por rating mínimo
	if filters.MinRating != nil {
		query = query.Where("rating >= ?", *filters.MinRating)
	}

	// Filtrar por envío gratis
	if filters.FreeShipping != nil {
		query = query.Where("free_shipping = ?", *filters.FreeShipping)
	}

	// Filtrar por tipo de servicio
	if filters.IsService != nil {
		query = query.Where("is_service = ?", *filters.IsService)
	}

	query = query.Where("status = ?", "active")

	return query
}

// SearchProductsByCategory busca productos en categorías específicas
func SearchProductsByCategory(db *gorm.DB, categories []string, filters SearchFilters) (*SearchResult, error) {
	var products []Product
	var total int64

	// Configurar paginación por defecto
	if filters.Limit <= 0 {
		filters.Limit = 20
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}

	// Forzar las categorías en los filtros
	filters.Categories = categories

	// Query para productos en categorías específicas
	baseQuery := db.Model(&Product{}).Where("category_id IN ?", categories)

	// Aplicar otros filtros
	baseQuery = applySearchFilters(baseQuery, filters)

	// Obtener total
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Obtener productos ordenados por relevancia (rating, ventas, fecha)
	err := baseQuery.
		Order("rating DESC, sold DESC, created_at DESC").
		Limit(filters.Limit).
		Offset(filters.Offset).
		Find(&products).Error

	if err != nil {
		return nil, err
	}

	// Calcular paginación
	page := (filters.Offset / filters.Limit) + 1
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))

	return &SearchResult{
		Products:   products,
		Total:      total,
		Page:       page,
		PerPage:    filters.Limit,
		TotalPages: totalPages,
	}, nil
}
