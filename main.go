package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	H "mercadillo-global/helpers"
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
	if err := InitializeCategories(); err != nil {
		panic("Failed to initialize categories: " + err.Error())
	}

	e := echo.New()

	// Load templates with helper functions
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
		"formatNumber": func(number float64) string {
			return H.MaybeFormatNumber(number, true)
		},
		"isEmpty": H.IsEmpty,
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

	data := HomePageData{
		Title:            "Mercadillo Global - Compra y Vende Online",
		FeaturedProducts: getEnrichedProducts([]Product{}),
		Categories:       GetCategories(),
		PageTemplate:     "home-content",
	}
	c.Logger().Info("PageTemplate: ", data.PageTemplate)
	return c.Render(http.StatusOK, "base.html", data)
}

func categoryPage(c echo.Context) error {
	categoryId := c.Param("categoryId")
	clientIP := H.GetIP(c)
	c.Logger().Info("Category page accessed from IP: ", clientIP, " for category: ", categoryId)

	data := CategoryPageData{
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

	product := getEnrichedProduct(Product{})
	data := ProductPageData{
		Title:        product.Title + " - Mercadillo Global",
		Product:      product,
		Questions:    []Question{},
		Reviews:      []Review{},
		PageTemplate: "product-content",
	}
	return c.Render(http.StatusOK, "base.html", data)
}

func checkoutPage(c echo.Context) error {
	productId := c.Param("productId")
	clientIP := H.GetIP(c)
	c.Logger().Info("Checkout page accessed from IP: ", clientIP, " for product: ", productId)

	product := getEnrichedProduct(Product{})
	data := CheckoutPageData{
		Title:        "Checkout - " + product.Title,
		Product:      product,
		PageTemplate: "checkout-content",
	}
	return c.Render(http.StatusOK, "base.html", data)
}
