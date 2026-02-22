package controller

import (
	"net/http"
	"strconv"

	"go-wordpress/internal/config"
	"go-wordpress/internal/http/response"
	"go-wordpress/internal/middleware"
	"go-wordpress/internal/product/service"
	"go-wordpress/internal/shared"
	"go-wordpress/internal/storage/sql/sqlc"

	"github.com/gin-gonic/gin"
)

type AdminProduct struct {
	Service *service.Product
}

func NewAdminProduct(s *service.Product) *AdminProduct {
	return &AdminProduct{Service: s}
}

func (c *AdminProduct) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	auth := middleware.JWTAuth(cfg)

	rg.POST("/", auth, c.CreateProduct)
	rg.PUT("/:id", auth, c.UpdateProduct)
	rg.DELETE("/:id", auth, c.DeleteProduct)
	rg.GET("/:id", auth, c.GetProductByID)
	rg.GET("/", auth, c.ListProducts)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with the provided details
// @Tags Admin Products
// @Accept json
// @Produce json
// @Param product body sqlc.CreateProductParams true "Product to create"
// @Success 201 {object} sqlc.Product
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/products [post]
func (c *AdminProduct) CreateProduct(ctx *gin.Context) {
	var req sqlc.CreateProductParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	product, err := c.Service.Create(ctx, req)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusCreated, product)
}

// UpdateProduct godoc
// @Summary Update an existing product
// @Description Update product details by ID
// @Tags Admin Products
// @Accept json
// @Produce json
// @Param id path int true "Product ID"
// @Param product body sqlc.UpdateProductParams true "Updated product details"
// @Success 200 {object} sqlc.Product
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/products/{id} [put]
func (c *AdminProduct) UpdateProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, response.ErrInvalidID)
		return
	}
	var req sqlc.UpdateProductParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	req.ID = int32(id)
	product, err := c.Service.Update(ctx, req)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, product)
}

// DeleteProduct godoc
// @Summary Delete a product by ID
// @Description Delete a product by its ID
// @Tags Admin Products
// @Param id path int true "Product ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/products/{id} [delete]
func (c *AdminProduct) DeleteProduct(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, response.ErrInvalidID)
		return
	}
	if err := c.Service.Delete(ctx, int32(id)); err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

// GetProductByID godoc
// @Summary Get a product by ID
// @Description Get a product by its ID
// @Tags Admin Products
// @Param id path int true "Product ID"
// @Success 200 {object} sqlc.Product
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/products/{id} [get]
func (c *AdminProduct) GetProductByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, response.ErrInvalidID)
		return
	}
	product, err := c.Service.GetProductByID(ctx, int32(id))
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, product)
}

// ListProducts godoc
// @Summary List all products
// @Description Get a list of all products
// @Tags Admin Products
// @Param product body shared.Pagination false "Pagination parameters"
// @Produce json
// @Success 200 {array} sqlc.Product
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/products [get]
func (c *AdminProduct) ListProducts(ctx *gin.Context) {
	pagination := shared.Pagination{
		Limit:  10,
		Offset: 0,
	}
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	products, err := c.Service.ListProducts(ctx, pagination)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, products)
}
