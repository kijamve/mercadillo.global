package main

import (
	"html/template"
	"io"
	"net/http"

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
		FeaturedProducts: getEnrichedProducts([]models.Product{}),
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

	data := models.CategoryPageData{
		Title:        getCategoryName(categoryId) + " - Mercadillo Global",
		CategoryId:   categoryId,
		CategoryName: getCategoryName(categoryId),
		Products:     getEnrichedProducts(getCategoryProducts(categoryId)),
		Filters:      getFilters(),
		PageTemplate: "category-content",
	}
	return c.Render(http.StatusOK, "base.html", data)
}

func productPage(c echo.Context) error {
	productId := c.Param("productId")
	clientIP := H.GetIP(c)
	c.Logger().Info("Product page accessed from IP: ", clientIP, " for product: ", productId)

	product := getEnrichedProduct(models.Product{})
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

	product := getEnrichedProduct(models.Product{})
	data := models.CheckoutPageData{
		Title:        "Checkout - " + product.Title,
		Product:      product,
		PageTemplate: "checkout-content",
	}
	return c.Render(http.StatusOK, "base.html", data)
}

// Helper functions that need to be implemented
func getEnrichedProducts(products []models.Product) []models.EnrichedProduct {
	// Placeholder implementation
	return []models.EnrichedProduct{}
}

func getEnrichedProduct(product models.Product) models.EnrichedProduct {
	// Placeholder implementation
	return models.EnrichedProduct{}
}

func getCategoryName(categoryId string) string {
	category := models.GetCategoryByID(categoryId)
	if category != nil {
		return category.Name
	}
	return "Categoría Desconocida"
}

func getCategoryProducts(categoryId string) []models.Product {
	// Placeholder implementation
	return []models.Product{}
}

func getFilters() []models.Filter {
	return []models.Filter{
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
