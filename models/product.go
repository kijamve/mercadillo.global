package models

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	H "mercadillo-global/helpers"
)

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
	SearchContent  string    `json:"search_content" gorm:"type:text;comment:'AI-generated optimized search content: title + category + key specs'"`
	SearchKeywords string    `json:"search_keywords" gorm:"type:varchar(500);comment:'AI-generated comma-separated keywords for enhanced search'"`
	Status         string    `json:"status" gorm:"type:enum('active','wait_for_ia','wait_for_human_review','pause','draft');default:'draft'"`
	KYC            bool      `json:"kyc" gorm:"default:false"`
	FromCompany    bool      `json:"from_company" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relations
	User              User               `json:"user" gorm:"foreignKey:UserID"`
	Questions         []Question         `json:"questions" gorm:"foreignKey:ProductID"`
	Reviews           []Review           `json:"reviews" gorm:"foreignKey:ProductID"`
	Attributes        []ProductAttribute `json:"attributes" gorm:"foreignKey:ProductID"`
	Warehouses        []ProductWarehouse `json:"warehouses" gorm:"foreignKey:ProductID"`
	Categories        []Category         `json:"categories" gorm:"many2many:product_categories;"`
	ProductCategories []ProductCategory  `json:"product_categories" gorm:"foreignKey:ProductID"`
}

type ProductCategory struct {
	ID         string    `json:"id" gorm:"type:char(36);primaryKey"`
	ProductID  string    `json:"product_id" gorm:"type:char(36);not null;index"`
	CategoryID string    `json:"category_id" gorm:"type:varchar(36);not null;index"`
	IsPrimary  bool      `json:"is_primary" gorm:"default:false;comment:'Indicates if this is the primary category for the product'"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	Product  Product  `json:"product" gorm:"foreignKey:ProductID"`
	Category Category `json:"category" gorm:"foreignKey:CategoryID"`
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

// EnrichedProduct estructura para productos enriquecidos con datos calculados
type EnrichedProduct struct {
	Product
	FormattedPrice         string               `json:"formatted_price"`
	FormattedOriginalPrice string               `json:"formatted_original_price"`
	Discount               int                  `json:"discount"`
	Stars                  []int                `json:"stars"`
	RatingInt              int                  `json:"rating_int"`
	AvailableWarehouses    []ProductWarehouse   `json:"available_warehouses"`
	TotalStock             int                  `json:"total_stock"`
	TotalWeight            int                  `json:"total_weight"` // en gramos
	GlobalAttributes       []ProductAttribute   `json:"global_attributes"`
	ShippingOptions        []ShippingCost       `json:"shipping_options"`
	PrimaryCategory        *Category            `json:"primary_category"` // La categoría principal del producto
	AllCategories          map[string]*Category `json:"all_categories"`   // Todas las categorías del producto
}

// Specification struct for JSON serialization
type Specification struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ProductVariation estructura para variaciones de producto
type ProductVariation struct {
	AttributeSlug string                 `json:"attribute_slug"`
	Value         map[string]interface{} `json:"value"`
	WarehouseID   *string                `json:"warehouse_id,omitempty"`
	IsGlobal      bool                   `json:"is_global"`
}

// CategoryFilters estructura para filtros de categoría
type CategoryFilters struct {
	PriceMin     *int   `json:"price_min,omitempty"`
	PriceMax     *int   `json:"price_max,omitempty"`
	Rating       *int   `json:"rating,omitempty"`
	Reviews      *int   `json:"reviews,omitempty"`
	Sales        *int   `json:"sales,omitempty"`
	FreeShipping *bool  `json:"free_shipping,omitempty"`
	SortBy       string `json:"sort_by,omitempty"`
}

// GORM Hooks
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(p.ID) {
		p.ID = H.NewUUID()
	}
	return nil
}

func (pa *ProductAttribute) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(pa.ID) {
		pa.ID = H.NewUUID()
	}
	return nil
}

func (pc *ProductCategory) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(pc.ID) {
		pc.ID = H.NewUUID()
	}
	return nil
}

// GenerateSearchContent genera contenido optimizado para búsqueda usando IA
func (p *Product) GenerateSearchContent() {
	// Esta función debería integrarse con tu servicio de IA
	// Por ahora genero un contenido básico como ejemplo

	// Obtener nombres de todas las categorías asociadas
	categoryNames := ""
	for _, category := range p.Categories {
		if categoryNames != "" {
			categoryNames += " "
		}
		categoryNames += category.Name
	}

	// Si no hay categorías cargadas pero hay ProductCategories, usar esos IDs
	if categoryNames == "" && len(p.ProductCategories) > 0 {
		for _, pc := range p.ProductCategories {
			category := GetCategoryByID(pc.CategoryID)
			if category != nil {
				if categoryNames != "" {
					categoryNames += " "
				}
				categoryNames += category.Name
			}
		}
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
		// keywords += spec.Value + ", "
	}

	// Generar search_content optimizado
	p.SearchContent = fmt.Sprintf("%s %s %s %s",
		p.Title,
		categoryNames,
		p.Description,
		specText)

	// Generar keywords
	p.SearchKeywords = fmt.Sprintf("%s, %s, %s",
		p.Title,
		categoryNames,
		keywords)
}

// GetProductWithWarehouses obtiene un producto con todos sus almacenes y atributos
func GetProductWithWarehouses(db *gorm.DB, productID string) (*EnrichedProduct, error) {
	var product Product

	// Obtener producto con relaciones incluyendo las categorías
	err := db.Preload("User").
		Preload("Warehouses").
		Preload("Warehouses.Warehouse").
		Preload("Warehouses.ShippingCosts").
		Preload("Warehouses.Attributes").
		Preload("Warehouses.Attributes.ProductWarehouse").
		Preload("Attributes").
		Preload("Attributes.ProductWarehouse").
		Preload("Categories").
		Preload("ProductCategories").
		Preload("Questions").
		Preload("Questions.QuestionVotes").
		Preload("Questions.QuestionVotes.User").
		Preload("Reviews").
		Preload("Reviews.ReviewVotes").
		Preload("Reviews.ReviewVotes.User").
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

	// Cargar las categorías desde el sistema de categorías global
	var allCategories map[string]*Category
	var primaryCategory *Category

	for _, pc := range product.ProductCategories {
		category := GetCategoryByID(pc.CategoryID)
		if category != nil {
			allCategories[pc.CategoryID] = category

			// Si es la categoría primaria, guardarla por separado
			if pc.IsPrimary {
				primaryCategory = category
			}
		}
	}

	// Si no hay categoría primaria definida pero sí hay categorías, usar la primera como primaria
	if primaryCategory == nil && len(allCategories) > 0 {
		primaryCategory = allCategories[product.ProductCategories[0].CategoryID]
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
		PrimaryCategory:        primaryCategory,
		AllCategories:          allCategories,
	}

	return enrichedProduct, nil
}

// calculateDiscount calculates discount percentage
func calculateDiscount(originalPrice, price int) int {
	if originalPrice == 0 {
		return 0
	}
	return int(float64(originalPrice-price) / float64(originalPrice) * 100)
}

// GetProductsByCategoryCursor versión ultra-optimizada usando cursor pagination encriptado
// Más eficiente para millones de registros que OFFSET/LIMIT
func GetProductsByCategoryCursor(db *gorm.DB, categoryID string, encryptedCursor string, limit int, filters CategoryFilters) ([]Product, string, bool, error) {
	var products []Product

	query := db.Table("products p").
		Select("p.*").
		Joins("INNER JOIN product_categories pc ON p.id = pc.product_id").
		Where("pc.category_id = ? AND p.status = ?", categoryID, "active")

	// Desencriptar cursor si existe para obtener datos de paginación
	var cursorData H.CursorData
	if encryptedCursor != "" {
		if decoded, err := H.DecryptCursor(encryptedCursor); err == nil {
			cursorData = decoded
		}
		// Si hay error al desencriptar, cursorData permanece vacío y se inicia desde el principio
	}

	// Aplicar filtros combinando filtros de usuario y condiciones de cursor
	query = applyCategoryFiltersWithCursor(query, filters, cursorData)

	// Determinar orden basado en filtros
	orderBy := "p.created_at DESC"
	if filters.SortBy != "" {
		switch filters.SortBy {
		case "price_asc":
			orderBy = "p.price ASC, p.created_at DESC"
		case "price_desc":
			orderBy = "p.price DESC, p.created_at DESC"
		case "rating":
			orderBy = "p.rating DESC, p.created_at DESC"
		case "sales":
			orderBy = "p.sold DESC, p.created_at DESC"
		case "newest":
			orderBy = "p.created_at DESC"
		}
	}

	// Obtener productos con orden
	err := query.Order(orderBy).
		Limit(limit + 1). // +1 para saber si hay más páginas
		Find(&products).Error

	if err != nil {
		return nil, "", false, err
	}

	// Determinar si hay más páginas
	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit] // Remover el elemento extra
	}

	// Generar cursor encriptado para la siguiente página
	var nextEncryptedCursor string
	if hasMore && len(products) > 0 {
		lastProduct := products[len(products)-1]

		// Crear datos del cursor basado en el tipo de ordenamiento
		cursorData := H.CursorData{
			Timestamp: lastProduct.CreatedAt.Format("2006-01-02T15:04:05Z"),
			SortBy:    filters.SortBy,
		}

		// Agregar campo específico según el ordenamiento
		switch filters.SortBy {
		case "price_asc", "price_desc":
			cursorData.Price = &lastProduct.Price
		case "rating":
			cursorData.Rating = &lastProduct.Rating
		case "sales":
			cursorData.Sold = &lastProduct.Sold
		}

		// Encriptar cursor
		nextEncryptedCursor, err = H.EncryptCursor(cursorData)
		if err != nil {
			// En caso de error al encriptar, continuar sin cursor
			// En producción deberías loggear este error
			nextEncryptedCursor = ""
		}
	}

	return products, nextEncryptedCursor, hasMore, nil
}

// applyCategoryFiltersWithCursor aplica filtros de categoría junto con condiciones de cursor de forma integrada
func applyCategoryFiltersWithCursor(query *gorm.DB, filters CategoryFilters, cursorData H.CursorData) *gorm.DB {
	// PASO 1: Aplicar SIEMPRE todos los filtros del usuario (sin importar el cursor)
	if filters.PriceMin != nil {
		query = query.Where("p.price >= ?", *filters.PriceMin)
	}
	if filters.PriceMax != nil {
		query = query.Where("p.price <= ?", *filters.PriceMax)
	}
	if filters.Rating != nil {
		query = query.Where("p.rating >= ?", *filters.Rating)
	}
	if filters.Reviews != nil {
		if *filters.Reviews == 0 {
			query = query.Where("p.review_count = 0")
		} else {
			query = query.Where("p.review_count >= ?", *filters.Reviews)
		}
	}
	if filters.Sales != nil {
		if *filters.Sales == 0 {
			query = query.Where("p.sold = 0")
		} else {
			query = query.Where("p.sold >= ?", *filters.Sales)
		}
	}
	if filters.FreeShipping != nil {
		query = query.Where("p.free_shipping = ?", *filters.FreeShipping)
	}

	// PASO 2: Aplicar condición de cursor SOLO para paginación (AND con los filtros de arriba)
	if cursorData.Timestamp != "" {
		switch filters.SortBy {
		case "price_asc":
			if cursorData.Price != nil {
				// Continuar desde donde quedamos: price > cursor_price OR (price = cursor_price AND created_at < cursor_timestamp)
				query = query.Where("(p.price > ? OR (p.price = ? AND p.created_at < ?))",
					*cursorData.Price, *cursorData.Price, cursorData.Timestamp)
			} else {
				query = query.Where("p.created_at < ?", cursorData.Timestamp)
			}
		case "price_desc":
			if cursorData.Price != nil {
				// Continuar desde donde quedamos: price < cursor_price OR (price = cursor_price AND created_at < cursor_timestamp)
				query = query.Where("(p.price < ? OR (p.price = ? AND p.created_at < ?))",
					*cursorData.Price, *cursorData.Price, cursorData.Timestamp)
			} else {
				query = query.Where("p.created_at < ?", cursorData.Timestamp)
			}
		case "rating":
			if cursorData.Rating != nil {
				// Continuar desde donde quedamos: rating < cursor_rating OR (rating = cursor_rating AND created_at < cursor_timestamp)
				query = query.Where("(p.rating < ? OR (p.rating = ? AND p.created_at < ?))",
					*cursorData.Rating, *cursorData.Rating, cursorData.Timestamp)
			} else {
				query = query.Where("p.created_at < ?", cursorData.Timestamp)
			}
		case "sales":
			if cursorData.Sold != nil {
				// Continuar desde donde quedamos: sold < cursor_sold OR (sold = cursor_sold AND created_at < cursor_timestamp)
				query = query.Where("(p.sold < ? OR (p.sold = ? AND p.created_at < ?))",
					*cursorData.Sold, *cursorData.Sold, cursorData.Timestamp)
			} else {
				query = query.Where("p.created_at < ?", cursorData.Timestamp)
			}
		default:
			// Para ordenamiento por fecha o sin ordenamiento específico
			query = query.Where("p.created_at < ?", cursorData.Timestamp)
		}
	}

	return query
}
