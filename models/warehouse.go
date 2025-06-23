package models

import (
	"time"

	"gorm.io/gorm"

	H "mercadillo-global/helpers"
)

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

// GORM Hooks
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
