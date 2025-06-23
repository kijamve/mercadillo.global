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
	FormattedPrice         string             `json:"formatted_price"`
	FormattedOriginalPrice string             `json:"formatted_original_price"`
	Discount               int                `json:"discount"`
	Stars                  []int              `json:"stars"`
	RatingInt              int                `json:"rating_int"`
	AvailableWarehouses    []ProductWarehouse `json:"available_warehouses"`
	TotalStock             int                `json:"total_stock"`
	TotalWeight            int                `json:"total_weight"` // en gramos
	GlobalAttributes       []ProductAttribute `json:"global_attributes"`
	ShippingOptions        []ShippingCost     `json:"shipping_options"`
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

// calculateDiscount calculates discount percentage
func calculateDiscount(originalPrice, price int) int {
	if originalPrice == 0 {
		return 0
	}
	return int(float64(originalPrice-price) / float64(originalPrice) * 100)
}
