package router

import (
	"math-top/internal/handler"
	"math-top/internal/middleware"
	"math-top/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterAdminRoutes(rg *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	adminService := service.NewAdminService(db, rdb)
	adminHandler := handler.NewAdminHandler(adminService)

	rg.POST("/auth/login", adminHandler.Login)

	auth := rg.Group("")
	auth.Use(middleware.AdminJWTAuthMiddleware(rdb))
	auth.Use(middleware.AdminAuthMiddleware())
	{
		auth.GET("/dashboard", adminHandler.Dashboard)

		eventService := service.NewAdminEventService(db)
		eventHandler := handler.NewAdminEventHandler(eventService)
		auth.GET("/events", eventHandler.List)
		auth.GET("/events/:id", eventHandler.Detail)
		auth.POST("/events", eventHandler.Create)
		auth.PUT("/events/:id", eventHandler.Update)
		auth.DELETE("/events/:id", eventHandler.Delete)
		auth.PATCH("/events/:id/feature", eventHandler.ToggleFeature)

		newsService := service.NewAdminNewsService(db)
		newsHandler := handler.NewAdminNewsHandler(newsService)
		auth.GET("/news", newsHandler.List)
		auth.GET("/news/:id", newsHandler.Detail)
		auth.POST("/news", newsHandler.Create)
		auth.PUT("/news/:id", newsHandler.Update)
		auth.DELETE("/news/:id", newsHandler.Delete)

		resourceService := service.NewAdminResourceService(db)
		resourceHandler := handler.NewAdminResourceHandler(resourceService)
		auth.GET("/resources", resourceHandler.List)
		auth.GET("/resources/:id", resourceHandler.Detail)
		auth.PUT("/resources/:id", resourceHandler.Update)
		auth.DELETE("/resources/:id", resourceHandler.Delete)

		showcaseService := service.NewAdminShowcaseService(db)
		showcaseHandler := handler.NewAdminShowcaseHandler(showcaseService)
		auth.GET("/showcases", showcaseHandler.List)
		auth.GET("/showcases/:id", showcaseHandler.Detail)
		auth.POST("/showcases", showcaseHandler.Create)
		auth.PUT("/showcases/:id", showcaseHandler.Update)
		auth.DELETE("/showcases/:id", showcaseHandler.Delete)
	}
}
