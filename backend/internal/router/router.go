package router

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/haojia/commute/internal/config"
	"github.com/haojia/commute/internal/handler"
	"github.com/haojia/commute/internal/middleware"
	"github.com/haojia/commute/internal/pkg/amap"
	"github.com/haojia/commute/internal/pkg/doubao"
	"github.com/haojia/commute/internal/repository"
	"github.com/haojia/commute/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg *config.Config, db *pgxpool.Pool, startedAt time.Time) *gin.Engine {
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Logger(), middleware.Recovery(), middleware.RequestID())

	addressRepo := repository.NewAddressRepo(db)
	companyRepo := repository.NewCompanyRepo(db)
	commuteRepo := repository.NewCommuteRepo(db)
	aiRepo := repository.NewAIRepo(db)

	amapClient := amap.New(amap.Config{
		Key:     cfg.Amap.WebServiceKey,
		BaseURL: cfg.Amap.BaseURL,
		Timeout: cfg.Amap.Timeout(),
	})
	doubaoClient := doubao.New(doubao.Config{
		APIKey:  cfg.Doubao.APIKey,
		BaseURL: cfg.Doubao.BaseURL,
		Model:   cfg.Doubao.Model,
		Timeout: cfg.Doubao.Timeout(),
	})

	profileSvc := service.NewProfileService(repository.NewProfileRepo(db))
	addressSvc := service.NewAddressService(addressRepo)
	companySvc := service.NewCompanyService(companyRepo)
	commuteSvc := service.NewCommuteService(commuteRepo, addressRepo, companyRepo, amapClient)
	aiSvc := service.NewAIService(aiRepo, doubaoClient, amapClient)

	health := handler.NewHealthHandler(db, cfg, startedAt)
	meta := handler.NewMetaHandler()
	profile := handler.NewProfileHandler(profileSvc)
	address := handler.NewAddressHandler(addressSvc)
	company := handler.NewCompanyHandler(companySvc)
	commute := handler.NewCommuteHandler(commuteSvc)
	mapH := handler.NewMapHandler(amapClient)
	aiH := handler.NewAIHandler(aiSvc)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", health.Get)
		v1.GET("/meta/enums", meta.Enums)

		v1.GET("/profile", profile.Get)
		v1.PUT("/profile", profile.Upsert)

		addresses := v1.Group("/addresses")
		{
			addresses.GET("", address.List)
			addresses.POST("", address.Create)
			addresses.GET("/:id", address.Get)
			addresses.PUT("/:id", address.Update)
			addresses.DELETE("/:id", address.Delete)
			addresses.POST("/:id/set-default", address.SetDefault)
		}

		companies := v1.Group("/companies")
		{
			companies.GET("", company.List)
			companies.POST("", company.Create)
			companies.POST("/batch", company.Batch)
			companies.GET("/:id", company.Get)
			companies.PUT("/:id", company.Update)
			companies.PATCH("/:id/status", company.UpdateStatus)
			companies.DELETE("/:id", company.Delete)
		}

		commuteGroup := v1.Group("/commute")
		{
			commuteGroup.POST("/calculate", commute.Calculate)
			commuteGroup.GET("/results/:id", commute.GetResult)
			commuteGroup.GET("/queries", commute.ListQueries)
			commuteGroup.GET("/queries/:id", commute.GetQuery)
			commuteGroup.GET("/queries/:id/results", commute.ListByQuery)
			commuteGroup.DELETE("/queries/:id", commute.DeleteQuery)
		}

		mapGroup := v1.Group("/map")
		{
			mapGroup.GET("/geocode", mapH.Geocode)
			mapGroup.GET("/regeocode", mapH.Regeocode)
			mapGroup.GET("/poi/search", mapH.POISearch)
		}

		aiGroup := v1.Group("/ai")
		{
			aiGroup.POST("/recommend/companies", aiH.RecommendCompanies)
		}
	}

	return r
}
