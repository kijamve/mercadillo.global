// pkg.pagination
package H

import (
	"math"
	"strconv"
	"strings"
	"sync"

	"gorm.io/gorm/schema"

	"github.com/labstack/echo/v4"

	"gorm.io/gorm"
)

type Pagination[T any] struct {
	Limit      int                    `json:"limit"`
	Page       int                    `json:"page"`
	Sort       string                 `json:"sort"`
	TotalRows  int64                  `json:"total_rows"`
	TotalPages int                    `json:"total_pages"`
	Rows       []T                    `json:"rows"`
	Filters    map[string]interface{} `json:"filters"`
}

func (p *Pagination[any]) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination[any]) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}
	if p.Limit > 100 {
		p.Limit = 500
	}
	return p.Limit
}

func (p *Pagination[any]) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}
	return p.Page
}

func (p *Pagination[any]) GetSort() string {
	if IsEmpty(p.Sort) {
		p.Sort = "created_at desc"
	}
	return p.Sort
}

func (p *Pagination[any]) FromContext(value interface{}, c echo.Context) {
	limit, _ := strconv.Atoi(c.Request().URL.Query().Get("limit"))
	page, _ := strconv.Atoi(c.Request().URL.Query().Get("page"))
	col_sort := c.Request().URL.Query().Get("col_sort")
	dir_sort := c.Request().URL.Query().Get("dir_sort")
	p.Rows = make([]any, 0)
	if limit > 0 {
		p.Limit = limit
	}
	if limit > 500 {
		p.Limit = 500
	}
	if page >= 0 {
		p.Page = page
	}
	if dir_sort == "" {
		dir_sort = "desc"
	}
	if col_sort == "" {
		col_sort = "created_at"
	}
	p.Sort = col_sort + " " + dir_sort
	s, err := schema.Parse(value, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		panic("failed to parse schema")
	}
	p.Filters = make(map[string]interface{})
	for _, field := range s.Fields {
		v := c.Request().URL.Query().Get("filter_like[" + field.DBName + "]")
		if !IsEmpty(v) {
			if field.GORMDataType == "int" {
				p.Filters[field.DBName+" = ?"], _ = strconv.Atoi(v)
			} else if field.GORMDataType == "string" {
				if len(v) > 3 {
					p.Filters[field.DBName+" LIKE ?"] = "%" + v + "%"
				} else {
					p.Filters[field.DBName+" LIKE ?"] = v
				}
			} else {
				p.Filters[field.DBName+" = ?"] = v
			}
		}
		v = c.Request().URL.Query().Get("filter_left_like[" + field.DBName + "]")
		if !IsEmpty(v) {
			if field.GORMDataType == "int" {
				p.Filters[field.DBName+" = ?"], _ = strconv.Atoi(v)
			} else if field.GORMDataType == "string" {
				p.Filters[field.DBName+" LIKE ?"] = v + "%"
			} else {
				p.Filters[field.DBName+" = ?"] = v
			}
		}
		v = c.Request().URL.Query().Get("filter[" + field.DBName + "]")
		if !IsEmpty(v) {
			if field.GORMDataType == "int" {
				p.Filters[field.DBName+" = ?"], _ = strconv.Atoi(v)
			} else if field.GORMDataType == "string" {
				p.Filters[field.DBName+" LIKE ?"] = v
			} else {
				p.Filters[field.DBName+" = ?"] = v
			}
		}
		v = c.Request().URL.Query().Get("filter_is_null[" + field.DBName + "]")
		if !IsEmpty(v) {
			if v == "true" {
				p.Filters[field.DBName+" IS NULL"] = true
			} else {
				p.Filters[field.DBName+" IS NOT NULL"] = true
			}
		}
	}
}

func Paginate(value interface{}, pagination *Pagination[interface{}], db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	currentDb := db.Debug().Model(value)
	if len(pagination.Filters) > 0 {
		for field, value := range pagination.Filters {
			switch v := value.(type) {
			case int:
				currentDb = currentDb.Where(field, v)
			case uint:
				currentDb = currentDb.Where(field, v)
			case int64:
				currentDb = currentDb.Where(field, v)
			case uint64:
				currentDb = currentDb.Where(field, v)
			case float64:
				currentDb = currentDb.Where(field, v)
			case bool:
				currentDb = currentDb.Where(field)
			default:
				currentDb = currentDb.Where(field, v)
			}
		}
		currentDb.Count(&totalRows)
	} else {
		currentDb.Count(&totalRows)
	}

	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.GetLimit())))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		currentDb := db
		for field, value := range pagination.Filters {
			if strings.Contains(field, " IS NULL") || strings.Contains(field, " IS NOT NULL") {
				currentDb = currentDb.Where(field)
			} else {
				currentDb = currentDb.Where(field, value)
			}
		}
		return currentDb.Offset(pagination.GetOffset()).Limit(pagination.GetLimit()).Order(pagination.GetSort())
	}
}

func FilterInvoicesByQueryParams(user_uuid string, db *gorm.DB, c echo.Context) (*gorm.DB, *GenericError) {
	var listCustomersUUID []string
	customerEmail := Trim(c.QueryParam("email"))
	customerDni := Trim(c.QueryParam("dni"))
	InvoiceNumber := Trim(c.QueryParam("number"))
	customerName := Trim(c.QueryParam("name"))

	if !IsEmpty(InvoiceNumber) {
		db = db.Where("invoice_number LIKE ?", InvoiceNumber+"%")
	}
	if !IsEmpty(customerDni) {
		dniOnlyWithDigits := RemoveNonNumeric(customerDni)
		dniOnlyDigitsAndLetters := strings.ToUpper(RemoveNonNumericAndLetters(customerDni))
		if len(dniOnlyDigitsAndLetters) < 3 {
			return nil, &GenericError{Message: TranslateText("The client's DNI is invalid, must be at least 3 characters", c)}
		}
		DB().Raw(
			"SELECT uuid FROM customers WHERE user_uuid = ? AND (dni LIKE ? OR dni_normalized1 LIKE ? OR dni_normalized2 LIKE ? OR company_vat LIKE ? OR company_vat_normalized1 LIKE ? OR company_vat_normalized2 LIKE ?) LIMIT 20",
			user_uuid,
			customerDni+"%",
			dniOnlyDigitsAndLetters+"%",
			dniOnlyWithDigits+"%",
			customerDni+"%",
			dniOnlyDigitsAndLetters+"%",
			dniOnlyWithDigits+"%",
		).Scan(&listCustomersUUID)
		if len(listCustomersUUID) == 0 {
			return db.Where("1=0"), nil
		}
		db = db.Where("customer_uuid IN (?)", listCustomersUUID)
	} else if !IsEmpty(customerName) {
		if len(customerName) < 3 {
			return nil, &GenericError{Message: TranslateText("The client's Name is invalid, must be at least 3 characters", c)}
		}
		DB().Raw("SELECT uuid FROM customers WHERE user_uuid = ? AND (name LIKE ? OR company LIKE ?) LIMIT 20",
			user_uuid,
			customerName+"%",
			customerName+"%",
		).Scan(&listCustomersUUID)
		if len(listCustomersUUID) == 0 {
			return db.Where("1=0"), nil
		}
		db = db.Where("customer_uuid IN (?)", listCustomersUUID)
	} else if !IsEmpty(customerEmail) {
		if len(customerEmail) < 3 {
			return nil, &GenericError{Message: TranslateText("The client's e-mail is invalid, must be at least 3 characters", c)}
		}
		DB().Raw("SELECT uuid FROM customers WHERE user_uuid = ? AND email LIKE ? LIMIT 20",
			user_uuid,
			customerEmail+"%",
		).Scan(&listCustomersUUID)
		if len(listCustomersUUID) == 0 {
			return db.Where("1=0"), nil
		}
		db = db.Where("customer_uuid IN (?)", listCustomersUUID)
	}
	return db, nil

}
