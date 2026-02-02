package crawler

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

type Category struct {
	Name            string
	Link            string
	Products        []Product
	CleanedProducts []ProductClean
}

type Website struct {
	Domain        string
	Categories    []Category
	StartURLs     []string            `json:"start_urls"`
	ProductList   ProductListConfig   `json:"product_list"`
	Pagination    *PaginationConfig   `json:"pagination,omitempty"`
	ProductDetail ProductDetailConfig `json:"product_detail"`
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
