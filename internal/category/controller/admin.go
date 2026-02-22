package controller

import (
	"net/http"
	"strconv"

	"go-wordpress/internal/category/service"
	"go-wordpress/internal/config"
	"go-wordpress/internal/http/response"
	"go-wordpress/internal/middleware"
	"go-wordpress/internal/shared"
	"go-wordpress/internal/storage/sql/sqlc"

	"github.com/gin-gonic/gin"
)

type AdminCategory struct {
	Service *service.Category
}

func NewAdminCategory(s *service.Category) *AdminCategory {
	return &AdminCategory{Service: s}
}

func (c *AdminCategory) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	auth := middleware.JWTAuth(cfg)

	rg.POST("/", auth, c.CreateCategory)
	rg.PUT("/:id", auth, c.UpdateCategory)
	rg.DELETE("/:id", auth, c.DeleteCategory)
	rg.GET("/:id", auth, c.GetCategoryByID)
	rg.GET("/", auth, c.ListCategories)
}

// CreateCategory godoc
// @Summary Create a new category
// @Description Create a new category with the provided details
// @Tags Admin Categories
// @Accept json
// @Produce json
// @Param category body sqlc.CreateCategoryParams true "Category to create"
// @Success 201 {object} sqlc.Category
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/categories [post]
func (c *AdminCategory) CreateCategory(ctx *gin.Context) {
	var req sqlc.CreateCategoryParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	category, err := c.Service.Create(ctx, req)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusCreated, category)
}

// UpdateCategory godoc
// @Summary Update an existing category
// @Description Update category details by ID
// @Tags Admin Categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param category body sqlc.UpdateCategoryParams true "Updated category details"
// @Success 200 {object} sqlc.Category
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/categories/{id} [put]
func (c *AdminCategory) UpdateCategory(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, response.ErrInvalidID)
		return
	}
	var req sqlc.UpdateCategoryParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	req.ID = int32(id)
	category, err := c.Service.Update(ctx, req)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, category)
}

// DeleteCategory godoc
// @Summary Delete a category by ID
// @Description Delete a category by its ID
// @Tags Admin Categories
// @Param id path int true "Category ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/categories/{id} [delete]
func (c *AdminCategory) DeleteCategory(ctx *gin.Context) {
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

// GetCategoryByID godoc
// @Summary Get a category by ID
// @Description Get a category by its ID
// @Tags Admin Categories
// @Param id path int true "Category ID"
// @Success 200 {object} sqlc.Category
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/categories/{id} [get]
func (c *AdminCategory) GetCategoryByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, response.ErrInvalidID)
		return
	}
	category, err := c.Service.GetCategoryByID(ctx, int32(id))
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, category)
}

// ListCategories godoc
// @Summary List all categories
// @Description Get a list of all categories
// @Tags Admin Categories
// @Param pagination query shared.Pagination false "Pagination parameters"
// @Produce json
// @Success 200 {array} sqlc.Category
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/categories [get]
func (c *AdminCategory) ListCategories(ctx *gin.Context) {
	var pagination shared.Pagination = shared.Pagination{
		Limit:  10,
		Offset: 0,
	}
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	categories, err := c.Service.ListCategories(ctx, pagination)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, categories)
}
