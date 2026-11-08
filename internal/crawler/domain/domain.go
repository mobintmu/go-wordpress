package domain

const (
	DefaultLimit  = 1000
	DefaultOffset = 0

	// ScraperConfigKey is the configs.key of the row holding a website's CFG.
	ScraperConfigKey = "scraper_config"
)

type Website struct {
	ID         int32      `json:"id"`
	Name       string     `json:"name"`
	Domain     string     `json:"domain"`
	Categories []Category `json:"categories"`
	CFG        CFG        `json:"cfg"`
}

type Category struct {
	Name            string `json:"name"`
	Link            string `json:"link"`
	Products        []Product
	CleanedProducts []ProductClean
}

type Product struct {
	Title       string
	Price       string
	Link        string
	Image       string
	Description string
}

type ProductClean struct {
	Title       string
	Price       int64
	Link        string
	Image       string
	Description string
}

type CFG struct {
	StartURL      string               `json:"start_url"`
	ProductList   ProductListConfig    `json:"product_list"`
	Pagination    *PaginationConfig    `json:"pagination,omitempty"`
	ProductDetail *ProductDetailConfig `json:"product_detail"`
}

type FieldConfig struct {
	Selector string `json:"selector"`
	Attr     string `json:"attr"`
}

type ProductListConfig struct {
	Item   string                 `json:"item"`
	Fields map[string]FieldConfig `json:"fields"`
}

type PaginationConfig struct {
	Next string `json:"next"`
}

type ProductDetailConfig struct {
	Description FieldConfig `json:"description"`
}
