# Mercadillo Global - Go Version

E-commerce platform built with Go, Echo framework, and Tailwind CSS. This is a server-side rendered version of the original React application for maximum SEO performance.

## Features

- **Server-Side Rendering**: Full HTML generation on the server for optimal SEO
- **Go + Echo**: High-performance web framework
- **Tailwind CSS**: Modern, responsive design
- **Multiple Pages**: Home, Category listings, Product details, Checkout
- **Mobile Responsive**: Works perfectly on all devices

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Git

### Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd go-project
```

2. Install dependencies:
```bash
go mod tidy
```

3. Run the development server:
```bash
go run .
```

The application will be available at `http://localhost:8080`

### Available Scripts

- `go run .` - Start the development server
- `go build -o bin/mercadillo-global` - Build for production
- `go mod tidy` - Install/update dependencies

## Project Structure

```
go-project/
├── main.go              # Application entry point and routes
├── models.go            # Data structures and business logic
├── templates.go         # Base HTML templates and layout
├── page_templates.go    # Page-specific template rendering
├── other_pages.go       # Additional page templates
├── go.mod              # Go module definition
├── go.sum              # Dependency lock file
└── README.md           # This file
```

## Routes

- `/` - Home page with featured products and categories
- `/category/:categoryId` - Category page with product listings and filters
- `/product/:productId` - Product detail page with images, specs, and reviews
- `/checkout/:productId` - Checkout page with shipping and payment forms

## Features Implemented

### Home Page
- Hero section with call-to-action
- Feature highlights (free shipping, secure payment, etc.)
- Category grid with images
- Featured products carousel
- Newsletter signup section

### Category Page
- Breadcrumb navigation
- Product grid with filtering sidebar
- Sorting options
- Pagination
- Responsive design

### Product Page
- Image gallery with thumbnails
- Product information and specifications
- Rating and reviews system
- Questions and answers section
- Add to cart functionality
- Related products

### Checkout Page
- Shipping information form
- Payment method selection
- Order summary
- Responsive design

## Styling

The application uses Tailwind CSS loaded via CDN for rapid development. The color scheme uses:
- Primary: `#fc8b06` (orange)
- Primary hover: `#e67c05`
- Primary dark: `#cc6d04`

## Data

Currently uses mock data for demonstration. In a production environment, you would integrate with:
- Database (PostgreSQL, MySQL, etc.)
- Payment processing (Stripe, PayPal, etc.)
- Image storage (AWS S3, Cloudinary, etc.)
- Search engine (Elasticsearch, Algolia, etc.)

## Performance

- Server-side rendering for fast initial page loads
- Optimized HTML structure
- Responsive images with proper sizing
- Minimal JavaScript for maximum compatibility

## SEO Optimization

- Semantic HTML structure
- Proper meta tags and titles
- Structured data ready
- Fast loading times
- Mobile-first design

## Deployment

For production deployment:

1. Build the application:
```bash
go build -o bin/mercadillo-global
```

2. Run the binary:
```bash
./bin/mercadillo-global
```

Or deploy to platforms like:
- Heroku
- Google Cloud Platform
- AWS
- DigitalOcean
- Railway

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is licensed under the MIT License. 