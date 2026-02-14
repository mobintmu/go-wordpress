package domain

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

type Websites struct {
	Websites []Website `json:"websites"`
}

type Website struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
}
