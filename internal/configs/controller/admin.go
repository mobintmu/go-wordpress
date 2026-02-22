package controller

import (
	"net/http"
	"strconv"

	"go-wordpress/internal/config"
	configsService "go-wordpress/internal/configs/service"
	"go-wordpress/internal/http/response"
	"go-wordpress/internal/middleware"
	"go-wordpress/internal/shared"
	"go-wordpress/internal/storage/sql/sqlc"

	"github.com/gin-gonic/gin"
)

type AdminConfig struct {
	Service *configsService.Config
}

func NewAdminConfig(s *configsService.Config) *AdminConfig {
	return &AdminConfig{Service: s}
}

func (c *AdminConfig) RegisterRoutes(rg *gin.RouterGroup, cfg *config.Config) {
	auth := middleware.JWTAuth(cfg)

	rg.POST("/", auth, c.CreateConfig)
	rg.PUT("/:id", auth, c.UpdateConfig)
	rg.DELETE("/:id", auth, c.DeleteConfig)
	rg.GET("/:id", auth, c.GetConfigByID)
	rg.GET("/", auth, c.ListConfigs)
}

// CreateConfig godoc
// @Summary Create a new config
// @Description Create a new config with the provided details
// @Tags Admin Configs
// @Accept json
// @Produce json
// @Param config body sqlc.CreateConfigParams true "Config to create"
// @Success 201 {object} sqlc.CreateConfigRow
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/configs [post]
func (c *AdminConfig) CreateConfig(ctx *gin.Context) {
	var req sqlc.CreateConfigParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	cfg, err := c.Service.Create(ctx, req)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusCreated, cfg)
}

// UpdateConfig godoc
// @Summary Update an existing config
// @Description Update config details by ID
// @Tags Admin Configs
// @Accept json
// @Produce json
// @Param id path int true "Config ID"
// @Param config body sqlc.UpdateConfigParams true "Updated config details"
// @Success 200 {object} sqlc.UpdateConfigRow
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/configs/{id} [put]
func (c *AdminConfig) UpdateConfig(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, response.ErrInvalidID)
		return
	}
	var req sqlc.UpdateConfigParams
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	req.ID = int32(id)
	cfg, err := c.Service.Update(ctx, req)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, cfg)
}

// DeleteConfig godoc
// @Summary Delete a config by ID
// @Description Delete a config by its ID
// @Tags Admin Configs
// @Param id path int true "Config ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/configs/{id} [delete]
func (c *AdminConfig) DeleteConfig(ctx *gin.Context) {
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

// GetConfigByID godoc
// @Summary Get a config by ID
// @Description Get a config by its ID
// @Tags Admin Configs
// @Param id path int true "Config ID"
// @Success 200 {object} sqlc.GetConfigByIDRow
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/configs/{id} [get]
func (c *AdminConfig) GetConfigByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response.JSONError(ctx, http.StatusBadRequest, response.ErrInvalidID)
		return
	}
	cfg, err := c.Service.GetConfigByID(ctx, int32(id))
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, cfg)
}

// ListConfigs godoc
// @Summary List all configs
// @Description Get a list of all configs
// @Tags Admin Configs
// @Param pagination query shared.Pagination false "Pagination parameters"
// @Produce json
// @Success 200 {array} sqlc.ListAllConfigsRow
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Security BearerAuth
// @Router /api/v1/admin/configs [get]
func (c *AdminConfig) ListConfigs(ctx *gin.Context) {
	pagination := shared.Pagination{
		Limit:  10,
		Offset: 0,
	}
	if err := ctx.ShouldBindQuery(&pagination); err != nil {
		response.JSONError(ctx, http.StatusBadRequest, err)
		return
	}
	configs, err := c.Service.ListConfigs(ctx, pagination)
	if err != nil {
		response.JSONError(ctx, http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, configs)
}
