package website

import (
	"encoding/json"
	"go-wordpress/internal/crawler/category"
	"go-wordpress/internal/crawler/domain"
	"os"

	"go.uber.org/zap"
)

type Service struct {
	Domain        string
	Categories    []domain.Category
	StartURL      string                     `json:"start_url"`
	ProductList   domain.ProductListConfig   `json:"product_list"`
	Pagination    *domain.PaginationConfig   `json:"pagination,omitempty"`
	ProductDetail domain.ProductDetailConfig `json:"product_detail"`
}

func New(logger *zap.Logger) []Service {
	webs, err := LoadWebSite("website.json")
	if err != nil {
		logger.Error("failed to load website configuration", zap.Error(err))
		return nil
	}
	var results []Service
	for _, w := range webs.Websites {
		path := w.Domain + ".json"
		w, err := LoadConfig(path)
		if err != nil {
			logger.Error("failed to load website config", zap.String("path", path), zap.Error(err))
			return nil
		}
		cat := category.New(logger, w.Domain, w.StartURL)
		err = cat.Discover()
		if err != nil {
			logger.Error("failed to discover categories", zap.String("domain", w.Domain), zap.Error(err))
			return nil
		}
		cat.SaveToFile()
		for _, dc := range cat.Categories {
			w.Categories = append(w.Categories, domain.Category{
				Name:            dc.Name,
				Link:            dc.Link,
				Products:        []domain.Product{},
				CleanedProducts: []domain.ProductClean{},
			})
		}
		results = append(results, *w)
	}

	return results
}

func NewWithCategories(logger *zap.Logger) []Service {
	webs, err := LoadWebSite("website.json")
	if err != nil {
		logger.Error("failed to load website configuration", zap.Error(err))
		return nil
	}
	var results []Service
	for _, w := range webs.Websites {
		path := w.Domain + ".json"
		w, err := LoadConfig(path)
		if err != nil {
			logger.Error("failed to load website config", zap.String("path", path), zap.Error(err))
			return nil
		}
		cat := category.New(logger, w.Domain, w.StartURL)
		cat.LoadFromFile()
		for _, dc := range cat.Categories {
			w.Categories = append(w.Categories, domain.Category{
				Name:            dc.Name,
				Link:            dc.Link,
				Products:        []domain.Product{},
				CleanedProducts: []domain.ProductClean{},
			})
		}
		results = append(results, *w)
	}
	return results
}

// LoadWebSite loads the configuration from a JSON file
func LoadWebSite(configPath string) (*domain.Websites, error) {
	w := &domain.Websites{}

	// os.ReadFile is the direct replacement for ioutil.ReadFile
	data, err := os.ReadFile("./config/" + configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

// LoadConfig loads the configuration from a JSON file
func LoadConfig(configPath string) (*Service, error) {
	w := &Service{}

	// os.ReadFile is the direct replacement for ioutil.ReadFile
	data, err := os.ReadFile("./config/" + configPath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, w)
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Service) LoadCategoriesFromFile(catPath string) error {
	data, err := os.ReadFile(catPath)
	if err != nil {
		return err
	}
	w.Categories = []domain.Category{}
	err = json.Unmarshal(data, &w.Categories)
	if err != nil {
		return err
	}
	return nil
}
