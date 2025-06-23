package main

import (
	"html/template"
	"io"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	H "mercadillo-global/helpers"
	"mercadillo-global/models"
)

// Template renderer
type TemplateRenderer struct {
	templates *template.Template
}

func (t *TemplateRenderer) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	// Initialize categories at startup
	if err := models.InitializeCategories(); err != nil {
		panic("Failed to initialize categories: " + err.Error())
	}

	e := echo.New()

	// Load templates with helper functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"formatNumber": func(number float64) string {
			return H.MaybeFormatNumber(number, true)
		},
		"isEmpty":       H.IsEmpty,
		"jsonDecode":    H.JSONDecode,
		"jsonDecodeMap": H.JSONDecodeMap,
	}

	templates := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/**/*.html"))
	e.Renderer = &TemplateRenderer{templates: templates}

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Static files (for CSS, JS, images)
	e.Static("/static", "static")

	// Routes
	e.GET("/", homePage)
	e.GET("/category/:categoryId", categoryPage)
	e.GET("/product/:productId", productPage)
	e.GET("/checkout/:productId", checkoutPage)

	// Start server
	e.Logger.Fatal(e.Start(":8080"))
}

func homePage(c echo.Context) error {
	// Log client IP using helper function
	clientIP := H.GetIP(c)
	c.Logger().Info("Home page accessed from IP: ", clientIP)

	data := models.HomePageData{
		Title:            "Mercadillo Global - Compra y Vende Online",
		FeaturedProducts: getEnrichedProducts(),
		Categories:       models.GetCategories(),
		PageTemplate:     "home-content",
	}
	c.Logger().Info("PageTemplate: ", data.PageTemplate)
	return c.Render(http.StatusOK, "base.html", data)
}

func categoryPage(c echo.Context) error {
	categoryId := c.Param("categoryId")
	clientIP := H.GetIP(c)
	c.Logger().Info("Category page accessed from IP: ", clientIP, " for category: ", categoryId)

	// Obtener parámetros para cursor pagination encriptado
	encryptedCursor := c.QueryParam("cursor") // Cursor encriptado
	limit := 12                               // Productos por página

	// Capturar filtros de la URL
	filters := models.CategoryFilters{}

	// Filtros de precio
	if priceMin := c.QueryParam("price_min"); priceMin != "" {
		if price, err := strconv.Atoi(priceMin); err == nil {
			filters.PriceMin = &price
		}
	}
	if priceMax := c.QueryParam("price_max"); priceMax != "" {
		if price, err := strconv.Atoi(priceMax); err == nil {
			filters.PriceMax = &price
		}
	}

	// Filtro de rating
	if rating := c.QueryParam("rating"); rating != "" {
		if ratingInt, err := strconv.Atoi(rating); err == nil {
			filters.Rating = &ratingInt
		}
	}

	// Filtro de reviews
	if reviews := c.QueryParam("reviews"); reviews != "" {
		if reviewsInt, err := strconv.Atoi(reviews); err == nil {
			filters.Reviews = &reviewsInt
		}
	}

	// Filtro de ventas
	if sales := c.QueryParam("sales"); sales != "" {
		if salesInt, err := strconv.Atoi(sales); err == nil {
			filters.Sales = &salesInt
		}
	}

	// Filtro de envío gratis
	if shipping := c.QueryParam("shipping"); shipping != "" {
		if shipping == "free" {
			freeShipping := true
			filters.FreeShipping = &freeShipping
		} else if shipping == "nonfree" {
			freeShipping := false
			filters.FreeShipping = &freeShipping
		}
	}

	// Filtro de ordenamiento
	if sortBy := c.QueryParam("sort"); sortBy != "" {
		filters.SortBy = sortBy
	}

	// Usar únicamente cursor pagination encriptado basado en timestamp (más eficiente para millones de registros)
	products, pagination, err := getCategoryProductsWithCursor(categoryId, encryptedCursor, limit, filters)
	if err != nil {
		c.Logger().Error("Error fetching category products: ", err)
		products = []models.EnrichedProduct{}
		pagination = models.Pagination{
			ItemsPerPage: limit,
			HasNext:      false,
			HasPrev:      false,
		}
	}

	data := models.CategoryPageData{
		Title:        getCategoryName(categoryId) + " - Mercadillo Global",
		CategoryId:   categoryId,
		CategoryName: getCategoryName(categoryId),
		Products:     products,
		Filters:      getFilters(),
		Pagination:   pagination,
		PageTemplate: "category-content",
	}
	return c.Render(http.StatusOK, "base.html", data)
}

func productPage(c echo.Context) error {
	productId := c.Param("productId")
	clientIP := H.GetIP(c)
	c.Logger().Info("Product page accessed from IP: ", clientIP, " for product: ", productId)

	product := getEnrichedProduct(c, productId)
	data := models.ProductPageData{
		Title:        product.Title + " - Mercadillo Global",
		Product:      product,
		Questions:    []models.Question{},
		Reviews:      []models.Review{},
		PageTemplate: "product-content",
	}
	return c.Render(http.StatusOK, "base.html", data)
}

func checkoutPage(c echo.Context) error {
	productId := c.Param("productId")
	clientIP := H.GetIP(c)
	c.Logger().Info("Checkout page accessed from IP: ", clientIP, " for product: ", productId)

	product := getEnrichedProduct(c, productId)
	data := models.CheckoutPageData{
		Title:        "Checkout - " + product.Title,
		Product:      product,
		PageTemplate: "checkout-content",
	}
	return c.Render(http.StatusOK, "base.html", data)
}

// Helper functions that need to be implemented
func getEnrichedProducts() []models.EnrichedProduct {
	// Obtener solo los IDs de los 100 mejores productos por rating y reviews
	var productIDs []string

	err := H.DB().Model(&models.Product{}).
		Select("id").
		Where("status = ? AND stock > 0", "active").
		Order("rating DESC, review_count DESC, created_at DESC").
		Limit(100).
		Pluck("id", &productIDs).Error

	if err != nil || len(productIDs) == 0 {
		return []models.EnrichedProduct{}
	}

	// Seleccionar 10 IDs aleatoriamente
	selectedIDs := selectRandomIDs(productIDs, 10)

	// Ahora sí obtener los datos completos solo de los 10 productos seleccionados
	var selectedProducts []models.Product
	err = H.DB().Where("id IN ?", selectedIDs).Find(&selectedProducts).Error

	if err != nil {
		return []models.EnrichedProduct{}
	}

	// Convertir a productos enriquecidos
	enrichedProducts := make([]models.EnrichedProduct, len(selectedProducts))
	for i, product := range selectedProducts {
		// Obtener la categoría primaria del producto
		var primaryCategory *models.Category
		var productCategories []models.ProductCategory

		H.DB().Where("product_id = ?", product.ID).Find(&productCategories)

		for _, pc := range productCategories {
			category := models.GetCategoryByID(pc.CategoryID)
			if category != nil && pc.IsPrimary {
				primaryCategory = category
				break
			}
		}

		// Si no hay categoría primaria, usar la primera disponible
		if primaryCategory == nil && len(productCategories) > 0 {
			category := models.GetCategoryByID(productCategories[0].CategoryID)
			if category != nil {
				primaryCategory = category
			}
		}

		enrichedProducts[i] = models.EnrichedProduct{
			Product:                product,
			FormattedPrice:         H.MaybeFormatNumber(float64(product.Price), true),
			FormattedOriginalPrice: H.MaybeFormatNumber(float64(product.OriginalPrice), true),
			Discount:               calculateDiscount(product.OriginalPrice, product.Price),
			Stars:                  []int{0, 1, 2, 3, 4},
			RatingInt:              int(product.Rating),
			PrimaryCategory:        primaryCategory,
		}
	}

	return enrichedProducts
}

// selectRandomIDs selecciona n IDs aleatoriamente de una lista
func selectRandomIDs(ids []string, n int) []string {
	if len(ids) <= n {
		return ids
	}

	// Crear una lista de índices
	indices := make([]int, len(ids))
	for i := range indices {
		indices[i] = i
	}

	// Mezclar los índices usando Fisher-Yates shuffle
	for i := len(indices) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		indices[i], indices[j] = indices[j], indices[i]
	}

	// Seleccionar los primeros n IDs
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = ids[indices[i]]
	}

	return result
}

func getEnrichedProduct(c echo.Context, productId string) models.EnrichedProduct {
	// Obtener el producto por ID
	var product models.Product
	err := H.DB().Preload("User").
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
		Where("id = ? AND status = ?", productId, "active").First(&product).Error
	if err != nil {
		c.Logger().Error("Error fetching product: ", err)
		return models.EnrichedProduct{}
	}

	// Obtener la categoría primaria del producto
	var primaryCategory *models.Category

	for _, pc := range product.ProductCategories {
		category := models.GetCategoryByID(pc.CategoryID)
		if category != nil && pc.IsPrimary {
			primaryCategory = category
			break
		}
	}

	var allCategories map[string]*models.Category

	// Si no hay categoría primaria, usar la primera disponible
	if primaryCategory == nil && len(product.ProductCategories) > 0 {
		category := models.GetCategoryByID(product.ProductCategories[0].CategoryID)
		if category != nil {
			primaryCategory = category
		}
		for _, category := range product.ProductCategories {
			allCategories[category.CategoryID] = models.GetCategoryByID(category.CategoryID)
		}
	}

	return models.EnrichedProduct{
		Product:                product,
		FormattedPrice:         H.MaybeFormatNumber(float64(product.Price), true),
		FormattedOriginalPrice: H.MaybeFormatNumber(float64(product.OriginalPrice), true),
		Discount:               calculateDiscount(product.OriginalPrice, product.Price),
		Stars:                  []int{0, 1, 2, 3, 4},
		RatingInt:              int(product.Rating),
		PrimaryCategory:        primaryCategory,
		AllCategories:          allCategories,
	}
}

func getCategoryName(categoryId string) string {
	category := models.GetCategoryByID(categoryId)
	if category != nil {
		return category.Name
	}
	return "Categoría Desconocida"
}

func getFilters() []models.Filter {
	return []models.Filter{
		{
			ID:   "price",
			Name: "Precio",
			Options: map[string]string{
				"fixed": "fixed",
			},
		},
		{
			ID:   "rating",
			Name: "Calificación",
			Options: map[string]string{
				"4": "4 estrellas o más",
				"3": "3 estrellas o más",
				"2": "2 estrellas o más",
			},
		},
		{
			ID:   "reviews",
			Name: "Cantidad de Reviews",
			Options: map[string]string{
				"3": "3 o más",
				"1": "1 o más",
				"0": "Ninguna",
			},
		},
		{
			ID:   "sales",
			Name: "Cantidad de Ventas",
			Options: map[string]string{
				"3": "3 o más",
				"1": "1 o más",
				"0": "Ninguna",
			},
		},
		{
			ID:   "shipping",
			Name: "Envío",
			Options: map[string]string{
				"free":    "Envío gratis",
				"nonfree": "Sin envío gratis",
			},
		},
	}
}

// getCategoryProductsWithCursor usa cursor pagination encriptado para mejor rendimiento
func getCategoryProductsWithCursor(categoryId, encryptedCursor string, limit int, filters models.CategoryFilters) ([]models.EnrichedProduct, models.Pagination, error) {
	products, nextEncryptedCursor, hasMore, err := models.GetProductsByCategoryCursor(H.DB(), categoryId, encryptedCursor, limit, filters)
	if err != nil {
		return nil, models.Pagination{}, err
	}

	// Convertir a productos enriquecidos
	enrichedProducts := make([]models.EnrichedProduct, len(products))
	for i, product := range products {
		enrichedProducts[i] = models.EnrichedProduct{
			Product:                product,
			FormattedPrice:         H.MaybeFormatNumber(float64(product.Price), true),
			FormattedOriginalPrice: H.MaybeFormatNumber(float64(product.OriginalPrice), true),
			Discount:               calculateDiscount(product.OriginalPrice, product.Price),
			Stars:                  []int{0, 1, 2, 3, 4},
			RatingInt:              int(product.Rating),
		}
	}

	// Paginación optimizada con cursors encriptados
	pagination := models.Pagination{
		ItemsPerPage: limit,
		HasNext:      hasMore,
		HasPrev:      encryptedCursor != "", // Si hay cursor, significa que hay página anterior
		NextCursor:   nextEncryptedCursor,
		PrevCursor:   encryptedCursor, // El cursor actual se convierte en el cursor anterior
	}

	return enrichedProducts, pagination, nil
}

// calculateDiscount helper function
func calculateDiscount(originalPrice, price int) int {
	if originalPrice == 0 {
		return 0
	}
	return int(float64(originalPrice-price) / float64(originalPrice) * 100)
}
