package controller

import (
	"net/http"
	"strconv"

	"go-wordpress/internal/config"
	"go-wordpress/internal/http/response"
	"go-wordpress/internal/middleware"
	"go-wordpress/internal/shared"
	"go-wordpress/internal/storage/sql/sqlc"
	"go-wordpress/internal/website/service"

	"github.com/gin-gonic/gin"
)

type AdminWebsite struct {
	Service *service.Website
}

func NewAdminWebsite(s *service.Website) *AdminWebsite {
	return &AdminWebsite{Service: s}
}

func (c *AdminWebsite) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	auth := middleware.JWTAuth(cfg)

	rg.POST("/", auth, c.CreateWebsite)
	rg.PUT("/:id", auth, c.UpdateWebsite)
	rg.DELETE("/:id", auth, c.DeleteWebsite)
	rg.GET("/:id", auth, c.GetWebsiteByID)
	rg.GET("/", auth, c.ListWebsites)
}

// CreateWebsite godoc
// @Summary Create a new website
// @Description Create a new website with the provided details
// @Tags Admin Websites
// @Accept json
// @Produce json
// @Param website body sqlc.CreateWebsiteParams true "Website to create"
// @Success 201 {object} sqlc.CreateWebsiteRow
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/websites [post]
func (c *AdminWebsite) CreateWebsite(ctx *gin.Context) {
	var req sqlc.CreateWebsiteParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	website, err := c.Service.Create(ctx, req)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusCreated, website)
}

// UpdateWebsite godoc
// @Summary Update an existing website
// @Description Update website details by ID
// @Tags Admin Websites
// @Accept json
// @Produce json
// @Param id path int true "Website ID"
// @Param website body sqlc.UpdateWebsiteParams true "Updated website details"
// @Success 200 {object} sqlc.Website
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/websites/{id} [put]
func (c *AdminWebsite) UpdateWebsite(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, response.ErrInvalidID)
		return
	}
	var req sqlc.UpdateWebsiteParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	req.ID = int32(id)
	website, err := c.Service.Update(ctx, req)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, website)
}

// DeleteWebsite godoc
// @Summary Delete a website by ID
// @Description Delete a website by its ID
// @Tags Admin Websites
// @Param id path int true "Website ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/websites/{id} [delete]
func (c *AdminWebsite) DeleteWebsite(ctx *gin.Context) {
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

// GetWebsiteByID godoc
// @Summary Get a website by ID
// @Description Get a website by its ID
// @Tags Admin Websites
// @Param id path int true "Website ID"
// @Success 200 {object} sqlc.Website
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/websites/{id} [get]
func (c *AdminWebsite) GetWebsiteByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, response.ErrInvalidID)
		return
	}
	website, err := c.Service.GetWebsiteByID(ctx, int32(id))
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, website)
}

// ListWebsites godoc
// @Summary List all websites
// @Description Get a list of all websites
// @Tags Admin Websites
// @Param pagination query shared.Pagination false "Pagination parameters"
// @Produce json
// @Success 200 {array} sqlc.Website
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/websites [get]
func (c *AdminWebsite) ListWebsites(ctx *gin.Context) {
	var pagination shared.Pagination = shared.Pagination{
		Limit:  10,
		Offset: 0,
	}
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	websites, err := c.Service.ListWebsites(ctx, pagination)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, websites)
}
