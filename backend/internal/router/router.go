package router

import (
	"log/slog"
	"math-top/internal/cache"
	"math-top/internal/config"
	"math-top/internal/handler"
	"math-top/internal/middleware"
	"math-top/internal/model"
	"math-top/internal/response"
	"math-top/internal/service"
	"math-top/internal/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func NewEngine() *gin.Engine {
	if config.GlobalConfig == nil {
		config.LoadConfig()
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CorsMiddleware())

	uploadBase := config.GlobalConfig.Storage.UploadDir
	if uploadBase == "" {
		uploadBase = "./storage/uploads"
	}
	r.Static("/uploads", uploadBase)

	db := config.ConnectMysql()
	rdb := config.InitRedis()
	c := cache.New(rdb)

	if err := db.AutoMigrate(
		&model.User{},
		&model.Event{},
		&model.News{},
		&model.Resource{},
		&model.Favorite{},
		&model.Showcase{},
		&model.Admin{},
		&model.DownloadLog{},
		&model.Comment{},
		&model.CommentLike{},
		&model.ChatMessage{},
		&model.ChatPresence{},
		&model.DailyMetric{},
	); err != nil {
		slog.Error("数据库迁移失败", "err", err)
		panic(err)
	}

	adminInitSvc := service.NewAdminService(db, rdb)
	if err := adminInitSvc.EnsureDefaultAdmin(); err != nil {
		slog.Warn("创建默认管理员失败", "err", err)
	}

	// 注入初始测试数据
	seedSvc := service.NewSeedService(db)
	seedSvc.AutoSeed()

	registerPublicRoutes(r, db, rdb, c)
	registerAuthRoutes(r, db, rdb, c)
	registerAdminRoutes(r, db, rdb, c)

	return r
}

func Addr() string {
	return ":" + config.GlobalConfig.App.Port
}

func registerPublicRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, c *cache.Cache) {
	r.GET("/", func(c *gin.Context) {
		response.Success(c, gin.H{"message": "pong"})
	})
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	userService := service.NewUserService(db, rdb)
	homepageService := service.NewHomepageService(db, c)
	eventService := service.NewEventService(db)
	newsService := service.NewNewsService(db)
	resourceService := service.NewResourceService(db)
	showcaseService := service.NewShowcaseService(db)
	commentService := service.NewCommentService(db)

	userHandler := handler.NewUserHandler(userService)
	homepageHandler := handler.NewHomepageHandler(homepageService)
	eventHandler := handler.NewEventHandler(eventService)
	newsHandler := handler.NewNewsHandler(newsService)
	resourceHandler := handler.NewResourceHandler(resourceService)
	showcaseHandler := handler.NewShowcaseHandler(showcaseService)
	commentHandler := handler.NewCommentHandler(commentService)

	public := r.Group("/api/v1")
	{
		authLimit := middleware.RateLimitMiddleware(rdb, 10, time.Minute)
		public.POST("/auth/register", authLimit, userHandler.Register)
		public.POST("/auth/login", authLimit, userHandler.Login)
		public.POST("/auth/logout", userHandler.Logout)

		public.POST("/active/track", func(c *gin.Context) {
			today := time.Now().Format("2006-01-02")
			ctx := c.Request.Context()

			// 1. 统计 PV (Page View)
			pvKey := "dau:pv:" + today
			rdb.Incr(ctx, pvKey)
			rdb.Expire(ctx, pvKey, 48*time.Hour)
			rdb.Incr(ctx, "dau:pv:all")

			// 2. 统计 UV (IP 独立访客)
			clientIP := c.ClientIP()
			if clientIP != "" {
				uvKey := "dau:ip:" + today
				rdb.PFAdd(ctx, uvKey, clientIP)
				rdb.Expire(ctx, uvKey, 48*time.Hour)
				rdb.PFAdd(ctx, "dau:ip:all", clientIP)
			}

			// 3. 统计 DAU (会员独立活跃数)
			var userID uint
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString := parts[1]
					redisKey := middleware.UserTokenPrefix + tokenString
					if rdb.Exists(ctx, redisKey).Val() > 0 {
						if claims, err := utils.ParseToken(tokenString); err == nil {
							userID = claims.UserID
						}
					}
				}
			}

			if userID > 0 {
				dauKey := "dau:user:" + today
				rdb.PFAdd(ctx, dauKey, userID)
				rdb.Expire(ctx, dauKey, 48*time.Hour)
				rdb.PFAdd(ctx, "dau:user:all", userID)
			}

			response.Success(c, nil)
		})

		public.GET("/home", homepageHandler.Get)

		public.GET("/events", eventHandler.List)
		public.GET("/events/:id", eventHandler.Detail)

		public.GET("/news", newsHandler.List)
		public.GET("/news/:id", newsHandler.Detail)

		public.GET("/resources", resourceHandler.List)
		public.GET("/resources/:id", resourceHandler.Detail)
		public.GET("/resources/download/:id", resourceHandler.Download)

		public.GET("/showcases", showcaseHandler.List)
		public.GET("/showcases/:id", showcaseHandler.Detail)

		public.GET("/comments", commentHandler.ListByTarget)
	}
}

func registerAuthRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, c *cache.Cache) {
	userService := service.NewUserService(db, rdb)
	memberService := service.NewMemberService(db)
	favoriteService := service.NewFavoriteService(db)
	resourceService := service.NewResourceService(db)
	commentService := service.NewCommentService(db)
	avatarService := service.NewAvatarService(db)
	chatService := service.NewChatService(db, c)

	userHandler := handler.NewUserHandler(userService)
	memberHandler := handler.NewMemberHandler(memberService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)
	resourceHandler := handler.NewResourceHandler(resourceService)
	commentHandler := handler.NewCommentHandler(commentService)
	avatarHandler := handler.NewAvatarHandler(avatarService)
	chatHandler := handler.NewChatHandler(chatService)

	auth := r.Group("/api/v1")
	auth.Use(middleware.JWTAuthMiddleware(rdb))
	{
		auth.POST("/auth/change-password", userHandler.ChangePassword)

		auth.GET("/profile", memberHandler.GetProfile)
		auth.PUT("/profile", memberHandler.UpdateProfile)
		auth.POST("/user/avatar", avatarHandler.Upload)

		auth.POST("/member/favorites", favoriteHandler.AddFavorite)
		auth.DELETE("/member/favorites/:id", favoriteHandler.Remove)
		auth.GET("/member/favorites", favoriteHandler.ListFavorites)

		auth.GET("/member/downloads", resourceHandler.MyDownloads)

		auth.POST("/comments", commentHandler.Create)
		auth.DELETE("/comments/:id", commentHandler.Delete)
		auth.POST("/comments/:id/like", commentHandler.ToggleLike)

		auth.POST("/chat/join", chatHandler.Join)
		auth.GET("/chat/messages", chatHandler.ListMessages)
		auth.POST("/chat/messages", chatHandler.SendText)
		auth.POST("/chat/messages/file", chatHandler.SendFile)
		auth.POST("/chat/leave", chatHandler.Leave)
		auth.DELETE("/chat/messages/:id", chatHandler.Delete)
	}
}

func registerAdminRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, c *cache.Cache) {
	adminService := service.NewAdminService(db, rdb)
	adminEventService := service.NewAdminEventService(db)
	adminNewsService := service.NewAdminNewsService(db)
	adminResourceService := service.NewAdminResourceService(db)
	adminShowcaseService := service.NewAdminShowcaseService(db)
	commentService := service.NewCommentService(db)
	adminUserService := service.NewAdminUserService(db)
	homepageService := service.NewHomepageService(db, c)
	resourceService := service.NewResourceService(db)
	chatService := service.NewChatService(db, c)

	adminHandler := handler.NewAdminHandler(adminService)
	adminEventHandler := handler.NewAdminEventHandler(adminEventService)
	adminNewsHandler := handler.NewAdminNewsHandler(adminNewsService)
	adminResourceHandler := handler.NewAdminResourceHandler(adminResourceService)
	adminShowcaseHandler := handler.NewAdminShowcaseHandler(adminShowcaseService)
	adminUserHandler := handler.NewAdminUserHandler(adminUserService)
	adminCommentHandler := handler.NewAdminCommentHandler(commentService)
	adminChatHandler := handler.NewAdminChatHandler(chatService)
	resourceHandler := handler.NewResourceHandler(resourceService)

	admin := r.Group("/api/v1/admin")
	adminLimit := middleware.RateLimitMiddleware(rdb, 20, time.Minute)
	admin.POST("/auth/login", adminLimit, adminHandler.Login)
	admin.POST("/auth/logout", adminHandler.Logout)

	auth := admin.Group("")
	auth.Use(middleware.AdminJWTAuthMiddleware(rdb))
	auth.Use(middleware.AdminAuthMiddleware())
	{
		auth.GET("/dashboard", adminHandler.Dashboard)
		auth.POST("/homepage/invalidate", func(c *gin.Context) {
			homepageService.Invalidate()
			response.Success(c, nil)
		})

		auth.GET("/events", adminEventHandler.List)
		auth.GET("/events/:id", adminEventHandler.Detail)
		auth.POST("/events", adminEventHandler.Create)
		auth.PUT("/events/:id", adminEventHandler.Update)
		auth.DELETE("/events/:id", adminEventHandler.Delete)
		auth.PATCH("/events/:id/feature", adminEventHandler.ToggleFeature)

		auth.GET("/news", adminNewsHandler.List)
		auth.GET("/news/:id", adminNewsHandler.Detail)
		auth.POST("/news", adminNewsHandler.Create)
		auth.PUT("/news/:id", adminNewsHandler.Update)
		auth.DELETE("/news/:id", adminNewsHandler.Delete)

		auth.GET("/resources", adminResourceHandler.List)
		auth.GET("/resources/:id", adminResourceHandler.Detail)
		auth.POST("/resources", resourceHandler.Upload)
		auth.PUT("/resources/:id", adminResourceHandler.Update)
		auth.DELETE("/resources/:id", adminResourceHandler.Delete)

		auth.GET("/showcases", adminShowcaseHandler.List)
		auth.GET("/showcases/:id", adminShowcaseHandler.Detail)
		auth.POST("/showcases", adminShowcaseHandler.Create)
		auth.PUT("/showcases/:id", adminShowcaseHandler.Update)
		auth.DELETE("/showcases/:id", adminShowcaseHandler.Delete)

		auth.GET("/users", adminUserHandler.List)
		auth.PATCH("/users/:id/status", adminUserHandler.SetStatus)
		auth.POST("/users/:id/reset-password", adminUserHandler.ResetPassword)
		auth.DELETE("/users/:id", adminUserHandler.Delete)

		auth.GET("/comments", adminCommentHandler.List)
		auth.DELETE("/comments/:id", adminCommentHandler.Delete)
		auth.PATCH("/comments/:id/status", adminCommentHandler.SetStatus)

		auth.GET("/chat/messages", adminChatHandler.List)
		auth.DELETE("/chat/messages/:id", adminChatHandler.Delete)
	}
}
