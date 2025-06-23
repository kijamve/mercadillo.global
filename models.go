package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"gorm.io/gorm"

	H "mercadillo-global/helpers"
)

// GORM Models for Database
type User struct {
	ID              string         `json:"id" gorm:"type:char(36);primaryKey"`
	Email           string         `json:"email" gorm:"type:varchar(255);uniqueIndex;not null"`
	Password        string         `json:"-" gorm:"type:varchar(255);not null"`
	KYCStatus       string         `json:"kyc_status" gorm:"type:enum('pending','approved','rejected');default:'pending'"`
	PlanSlug        string         `json:"plan_slug" gorm:"type:varchar(100);default:'free'"`
	Status          string         `json:"status" gorm:"type:enum('active','inactive','suspended');default:'active'"`
	EmailVerifiedAt *time.Time     `json:"email_verified_at" gorm:"type:timestamp null"`
	DeletedAt       gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Relations
	Products   []Product   `json:"products" gorm:"foreignKey:UserID"`
	Warehouses []Warehouse `json:"warehouses" gorm:"foreignKey:UserID"`
}

type Product struct {
	ID             string    `json:"id" gorm:"type:char(36);primaryKey"`
	ShortKey       string    `json:"short_key" gorm:"type:varchar(20);uniqueIndex;not null"`
	Slug           string    `json:"slug" gorm:"type:varchar(255);uniqueIndex;not null"`
	UserID         string    `json:"user_id" gorm:"type:char(36);not null;index"`
	Title          string    `json:"title" gorm:"type:varchar(500);not null"`
	Price          int       `json:"price" gorm:"not null"`
	OriginalPrice  int       `json:"original_price" gorm:"default:0"`
	CurrencyID     string    `json:"currency_id" gorm:"type:varchar(10);default:'USD'"`
	Images         string    `json:"images" gorm:"type:json"`
	Rating         float64   `json:"rating" gorm:"type:decimal(3,2);default:0"`
	ReviewCount    int       `json:"review_count" gorm:"default:0"`
	Sold           int       `json:"sold" gorm:"default:0"`
	Stock          int       `json:"stock" gorm:"default:0"`
	IsService      bool      `json:"is_service" gorm:"default:false"`
	FreeShipping   bool      `json:"free_shipping" gorm:"default:false"`
	Description    string    `json:"description" gorm:"type:text"`
	Specifications string    `json:"specifications" gorm:"type:json"`
	CategoryID     string    `json:"category_id" gorm:"type:varchar(50);not null;index"`
	SearchContent  string    `json:"search_content" gorm:"type:text;comment:'AI-generated optimized search content: title + category + key specs'"`
	SearchKeywords string    `json:"search_keywords" gorm:"type:varchar(500);comment:'AI-generated comma-separated keywords for enhanced search'"`
	Status         string    `json:"status" gorm:"type:enum('active','wait_for_ia','wait_for_human_review','pause','draft');default:'draft'"`
	KYC            bool      `json:"kyc" gorm:"default:false"`
	FromCompany    bool      `json:"from_company" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	User       User               `json:"user" gorm:"foreignKey:UserID"`
	Questions  []Question         `json:"questions" gorm:"foreignKey:ProductID"`
	Reviews    []Review           `json:"reviews" gorm:"foreignKey:ProductID"`
	Attributes []ProductAttribute `json:"attributes" gorm:"foreignKey:ProductID"`
	Warehouses []ProductWarehouse `json:"warehouses" gorm:"foreignKey:ProductID"`
}

type Warehouse struct {
	ID         string    `json:"id" gorm:"type:char(36);primaryKey"`
	UserID     string    `json:"user_id" gorm:"type:char(36);not null;index"`
	Name       string    `json:"name" gorm:"type:varchar(255);not null"`
	Country    string    `json:"country" gorm:"type:varchar(2);not null;index"`
	State      string    `json:"state" gorm:"type:varchar(100);not null;index"`
	City       string    `json:"city" gorm:"type:varchar(100);not null"`
	Address    string    `json:"address" gorm:"type:text;not null"`
	PostalCode string    `json:"postal_code" gorm:"type:varchar(20)"`
	Phone      string    `json:"phone" gorm:"type:varchar(50)"`
	Email      string    `json:"email" gorm:"type:varchar(255)"`
	IsActive   bool      `json:"is_active" gorm:"default:true;index"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	User              User               `json:"user" gorm:"foreignKey:UserID"`
	ProductWarehouses []ProductWarehouse `json:"product_warehouses" gorm:"foreignKey:WarehouseID"`
}

type ProductWarehouse struct {
	ID             string    `json:"id" gorm:"type:char(36);primaryKey"`
	ProductID      string    `json:"product_id" gorm:"type:char(36);not null;index"`
	WarehouseID    string    `json:"warehouse_id" gorm:"type:char(36);not null;index"`
	Quantity       int       `json:"quantity" gorm:"default:0;index"`
	Weight         float64   `json:"weight" gorm:"default:0;index;comment:'Weight in KG'"`
	Dimensions     string    `json:"dimensions" gorm:"type:json;comment:'JSON with length, width, height in cm'"`
	Specifications string    `json:"specifications" gorm:"type:json"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	Product       Product            `json:"product" gorm:"foreignKey:ProductID"`
	Warehouse     Warehouse          `json:"warehouse" gorm:"foreignKey:WarehouseID"`
	Attributes    []ProductAttribute `json:"attributes" gorm:"foreignKey:ProductWarehouseID"`
	ShippingCosts []ShippingCost     `json:"shipping_costs" gorm:"foreignKey:ProductWarehouseID"`
}

type ShippingCost struct {
	ID                 string    `json:"id" gorm:"type:char(36);primaryKey"`
	ProductWarehouseID string    `json:"product_warehouse_id" gorm:"type:char(36);not null;index"`
	Country            string    `json:"country" gorm:"type:varchar(2);not null;index"`
	Locations          string    `json:"locations" gorm:"type:json;comment:'Array of states/cities where this cost applies'"`
	Cost               float64   `json:"cost" gorm:"not null"`
	CurrencyID         string    `json:"currency_id" gorm:"type:varchar(3);default:'USD'"`
	PriceType          string    `json:"price_type" gorm:"type:enum('fixed','per_kg');default:'fixed';index"`
	MinWeight          *float64  `json:"min_weight" gorm:"comment:'Minimum weight for this cost (when price_type is per_kg)'"`
	MaxWeight          *float64  `json:"max_weight" gorm:"comment:'Maximum weight for this cost (when price_type is per_kg)'"`
	EstimatedDaysMin   *int      `json:"estimated_days_min"`
	EstimatedDaysMax   *int      `json:"estimated_days_max"`
	IsActive           bool      `json:"is_active" gorm:"default:true;index"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Relations
	ProductWarehouse ProductWarehouse `json:"product_warehouse" gorm:"foreignKey:ProductWarehouseID"`
}

type ProductAttribute struct {
	ID                 string    `json:"id" gorm:"type:char(36);primaryKey"`
	ProductID          string    `json:"product_id" gorm:"type:char(36);not null;index"`
	ProductWarehouseID *string   `json:"product_warehouse_id" gorm:"type:char(36);index"`
	AttributeSlug      string    `json:"attribute_slug" gorm:"type:varchar(100);not null;index"`
	Value              string    `json:"value" gorm:"type:json;not null"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`

	// Relations
	Product          Product           `json:"product" gorm:"foreignKey:ProductID"`
	ProductWarehouse *ProductWarehouse `json:"product_warehouse" gorm:"foreignKey:ProductWarehouseID"`
}

type Question struct {
	ID           string    `json:"id" gorm:"type:char(36);primaryKey"`
	ProductID    string    `json:"product_id" gorm:"type:char(36);not null;index"`
	Question     string    `json:"question" gorm:"type:text;not null"`
	Answer       string    `json:"answer" gorm:"type:text"`
	AnsweredByIA bool      `json:"answered_by_ia" gorm:"default:false"`
	Helpful      int       `json:"helpful" gorm:"default:0"`
	Status       string    `json:"status" gorm:"type:enum('wait_for_ia','wait_for_human_review','hidden','answered');default:'wait_for_ia'"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relations
	Product       Product        `json:"product" gorm:"foreignKey:ProductID"`
	QuestionVotes []QuestionVote `json:"question_votes" gorm:"foreignKey:QuestionID"`
}

type Review struct {
	ID        string    `json:"id" gorm:"type:char(36);primaryKey"`
	ProductID string    `json:"product_id" gorm:"type:char(36);not null;index"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	Rating    int       `json:"rating" gorm:"type:tinyint;not null"`
	Comment   string    `json:"comment" gorm:"type:text"`
	Helpful   int       `json:"helpful" gorm:"default:0"`
	Status    string    `json:"status" gorm:"type:enum('approved','wait_for_ia','wait_for_human_review','hidden');default:'wait_for_ia'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Product     Product      `json:"product" gorm:"foreignKey:ProductID"`
	ReviewVotes []ReviewVote `json:"review_votes" gorm:"foreignKey:ReviewID"`
}

type QuestionVote struct {
	ID         string    `json:"id" gorm:"type:char(36);primaryKey"`
	UserID     string    `json:"user_id" gorm:"type:char(36);not null;index"`
	QuestionID string    `json:"question_id" gorm:"type:char(36);not null;index"`
	Vote       int       `json:"vote" gorm:"type:tinyint;not null;comment:'1 for helpful, -1 for not helpful'"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	User     User     `json:"user" gorm:"foreignKey:UserID"`
	Question Question `json:"question" gorm:"foreignKey:QuestionID"`
}

type ReviewVote struct {
	ID        string    `json:"id" gorm:"type:char(36);primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:char(36);not null;index"`
	ReviewID  string    `json:"review_id" gorm:"type:char(36);not null;index"`
	Vote      int       `json:"vote" gorm:"type:tinyint;not null;comment:'1 for helpful, -1 for not helpful'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User   User   `json:"user" gorm:"foreignKey:UserID"`
	Review Review `json:"review" gorm:"foreignKey:ReviewID"`
}

// GORM Hooks
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(u.ID) {
		u.ID = H.NewUUID()
	}
	return nil
}

func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(p.ID) {
		p.ID = H.NewUUID()
	}
	return nil
}

func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(q.ID) {
		q.ID = H.NewUUID()
	}
	return nil
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(r.ID) {
		r.ID = H.NewUUID()
	}
	return nil
}

func (qv *QuestionVote) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(qv.ID) {
		qv.ID = H.NewUUID()
	}
	return nil
}

func (rv *ReviewVote) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(rv.ID) {
		rv.ID = H.NewUUID()
	}
	return nil
}

func (pa *ProductAttribute) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(pa.ID) {
		pa.ID = H.NewUUID()
	}
	return nil
}

func (w *Warehouse) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(w.ID) {
		w.ID = H.NewUUID()
	}
	return nil
}

func (pw *ProductWarehouse) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(pw.ID) {
		pw.ID = H.NewUUID()
	}
	return nil
}

func (sc *ShippingCost) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(sc.ID) {
		sc.ID = H.NewUUID()
	}
	return nil
}

// Specification struct for JSON serialization
type Specification struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

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

// WarehouseStock estructura para información de stock por almacén
type WarehouseStock struct {
	ProductWarehouseID string  `json:"product_warehouse_id"`
	WarehouseID        string  `json:"warehouse_id"`
	WarehouseName      string  `json:"warehouse_name"`
	Country            string  `json:"country"`
	State              string  `json:"state"`
	City               string  `json:"city"`
	Quantity           int     `json:"quantity"`
	Weight             float64 `json:"weight"`
}

// DimensionsCm estructura para dimensiones en centímetros
type DimensionsCm struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ShippingLocation estructura para ubicaciones de envío
type ShippingLocation struct {
	State  string   `json:"state,omitempty"`
	Cities []string `json:"cities,omitempty"`
}

// ProductVariation estructura para variaciones de producto
type ProductVariation struct {
	AttributeSlug string                 `json:"attribute_slug"`
	Value         map[string]interface{} `json:"value"`
	WarehouseID   *string                `json:"warehouse_id,omitempty"`
	IsGlobal      bool                   `json:"is_global"`
}

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

// GetProductWithWarehouses obtiene un producto con todos sus almacenes y atributos
func GetProductWithWarehouses(db *gorm.DB, productID string) (*EnrichedProduct, error) {
	var product Product

	// Obtener producto con relaciones
	err := db.Preload("Warehouses").
		Preload("Warehouses.Warehouse").
		Preload("Warehouses.ShippingCosts").
		Preload("Attributes").
		Preload("Attributes.ProductWarehouse").
		First(&product, "id = ?", productID).Error

	if err != nil {
		return nil, err
	}

	// Calcular stock total y peso total
	totalStock := 0
	totalWeight := 0.0
	var allShippingCosts []ShippingCost

	for _, warehouse := range product.Warehouses {
		totalStock += warehouse.Quantity
		totalWeight += warehouse.Weight * float64(warehouse.Quantity)
		allShippingCosts = append(allShippingCosts, warehouse.ShippingCosts...)
	}

	// Separar atributos globales de específicos por almacén
	var globalAttributes []ProductAttribute
	for _, attr := range product.Attributes {
		if attr.ProductWarehouseID == nil {
			globalAttributes = append(globalAttributes, attr)
		}
	}

	enrichedProduct := &EnrichedProduct{
		Product:                product,
		FormattedPrice:         H.MaybeFormatNumber(float64(product.Price), true),
		FormattedOriginalPrice: H.MaybeFormatNumber(float64(product.OriginalPrice), true),
		Discount:               calculateDiscount(product.OriginalPrice, product.Price),
		Stars:                  []int{0, 1, 2, 3, 4},
		RatingInt:              int(product.Rating),
		AvailableWarehouses:    product.Warehouses,
		TotalStock:             totalStock,
		TotalWeight:            int(totalWeight * 1000), // Convertir kg a gramos
		GlobalAttributes:       globalAttributes,
		ShippingOptions:        allShippingCosts,
	}

	return enrichedProduct, nil
}

// GetWarehousesByCountry obtiene almacenes por país
func GetWarehousesByCountry(db *gorm.DB, productID string, country string) ([]ProductWarehouse, error) {
	var productWarehouses []ProductWarehouse

	err := db.Preload("Warehouse").
		Joins("JOIN warehouses ON warehouses.id = product_warehouses.warehouse_id").
		Where("product_warehouses.product_id = ? AND warehouses.country = ?", productID, country).
		Find(&productWarehouses).Error

	return productWarehouses, err
}

// UpdateWarehouseStock actualiza el stock de un almacén específico de producto
func UpdateWarehouseStock(db *gorm.DB, productWarehouseID string, newQuantity int) error {
	return db.Model(&ProductWarehouse{}).
		Where("id = ?", productWarehouseID).
		Update("quantity", newQuantity).Error
}

// GetShippingCosts obtiene los costos de envío para un product_warehouse
func GetShippingCosts(db *gorm.DB, productWarehouseID string, country string) ([]ShippingCost, error) {
	var shippingCosts []ShippingCost

	query := db.Where("product_warehouse_id = ? AND is_active = ?", productWarehouseID, true)

	if country != "" {
		query = query.Where("country = ?", country)
	}

	err := query.Find(&shippingCosts).Error
	return shippingCosts, err
}

// CalculateShippingCost calcula el costo de envío para un peso específico
func CalculateShippingCost(shippingCost ShippingCost, weightKg float64) float64 {
	if shippingCost.PriceType == "fixed" {
		return shippingCost.Cost
	}

	return shippingCost.Cost * weightKg
}

// GetUserWarehouses obtiene todos los almacenes de un usuario
func GetUserWarehouses(db *gorm.DB, userID string) ([]Warehouse, error) {
	var warehouses []Warehouse

	err := db.Where("user_id = ? AND is_active = ?", userID, true).
		Find(&warehouses).Error

	return warehouses, err
}

// calculateDiscount calculates discount percentage
func calculateDiscount(originalPrice, price int) int {
	if originalPrice == 0 {
		return 0
	}
	return int(float64(originalPrice-price) / float64(originalPrice) * 100)
}

// GenerateSearchContent genera contenido optimizado para búsqueda usando IA
func (p *Product) GenerateSearchContent() {
	// Esta función debería integrarse con tu servicio de IA
	// Por ahora genero un contenido básico como ejemplo

	// Obtener el nombre de la categoría desde las categorías cargadas
	category := GetCategoryByID(p.CategoryID)
	categoryName := ""
	if category != nil {
		categoryName = category.Name
	}

	// Extraer especificaciones clave
	var specs []Specification
	if len(p.Specifications) > 0 {
		json.Unmarshal([]byte(p.Specifications), &specs)
	}

	specText := ""
	keywords := ""

	for _, spec := range specs {
		specText += spec.Name + " " + spec.Value + " "
		keywords += spec.Value + ", "
	}

	// Generar search_content optimizado
	p.SearchContent = fmt.Sprintf("%s %s %s %s",
		p.Title,
		categoryName,
		p.Description,
		specText)

	// Generar keywords
	p.SearchKeywords = fmt.Sprintf("%s, %s, %s",
		p.Title,
		categoryName,
		keywords)
}

// getCategoryName returns the name of a category by ID
func getCategoryName(categoryID string) string {
	category := GetCategoryByID(categoryID)
	if category != nil {
		return category.Name
	}
	return "Categoría"
}

func getCategoryProducts(categoryId string) []Product {
	return []Product{}
}

func getFilters() []Filter {
	return []Filter{
		{
			Name:    "Precio",
			Options: []string{"Menos de $50.000", "$50.000 - $200.000", "$200.000 - $500.000", "Más de $500.000"},
		},
		{
			Name:    "Marca",
			Options: []string{"Samsung", "Apple", "Nike", "Sony", "LG"},
		},
		{
			Name:    "Calificación",
			Options: []string{"4 estrellas o más", "3 estrellas o más", "2 estrellas o más"},
		},
		{
			Name:    "Envío",
			Options: []string{"Envío gratis", "Envío express"},
		},
	}
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
