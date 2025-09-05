package server

import (
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/sirrobot01/protodex/internal/server/web"
	pkgstore "github.com/sirrobot01/protodex/internal/store/pkg"

	"github.com/sirrobot01/protodex/internal/logger"
	"github.com/sirrobot01/protodex/internal/server/auth"
	"github.com/sirrobot01/protodex/internal/store"
)

type Server struct {
	packageStore pkgstore.Store
	authService  *auth.Service
	logger       zerolog.Logger
	router       *gin.Engine
}

func New(dataDir string, port int) *Server {
	// Get or create folder
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create data directory: %v", err))
	}
	// Initialize storage
	_store, err := store.New(dataDir)
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize store: %v", err))
	}

	if err := _store.Init(); err != nil {
		panic(fmt.Sprintf("Failed to initialize store: %v", err))
	}

	authService := auth.NewAuthService(_store.Auth())
	gin.SetMode(gin.ReleaseMode)

	server := &Server{
		packageStore: _store.Package(),
		authService:  authService,
		router:       gin.Default(),
		logger:       logger.Get(),
	}

	// Setup routes

	server.setupWebRoutes()

	// Log server start
	server.logger.Info().Msgf("Starting server on port %d", port)
	return server
}

func (s *Server) Start(port int) error {
	return s.router.Run(fmt.Sprintf(":%d", port))
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

func (s *Server) setupAPIRoutes() {
	api := s.router.Group("/api")

	api.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Auth routes
	authGroup := api.Group("/auth")
	{
		authGroup.POST("/login", s.loginHandler)
		authGroup.POST("/register", s.registerHandler)
		authGroup.GET("/me", s.authMiddleware(), s.getCurrentUserHandler)
	}

	// Protected package routes
	packages := api.Group("/packages")
	packages.GET("/search", s.searchPackagesHandler)
	packages.Use(s.authMiddleware())
	{
		packages.GET("", s.listPackagesHandler)
		packages.POST("", s.createPackageHandler)
		packages.GET("/:package", s.getPackageHandler)

		// Version routes
		packages.POST("/:package/versions", s.pushVersionHandler)
		packages.GET("/:package/versions", s.listVersionsHandler)
		packages.GET("/:package/versions/:version/files", s.pullVersionHandler)
		packages.GET("/:package/versions/:version/schema", s.viewSchemaHandler)
		packages.POST("/:package/versions/:version/generate", s.generateCodeHandler)
	}
}

func (s *Server) setupWebRoutes() {
	// Create a file server for the dist/assets directory specifically
	assetsFS, err := fs.Sub(web.StaticFS, "dist/assets")
	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed to set up static file server")
	}

	// Now we can use StaticFS directly since we're serving from the assets directory
	s.router.StaticFS("/assets", http.FS(assetsFS))

	s.setupAPIRoutes()
	s.serveReactApp()
}

// serveReactApp serves the React application index.html for all app routes
func (s *Server) serveReactApp() {
	s.router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Return 404 JSON for API routes
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "API endpoint not found",
				"path":  path,
			})
			return
		}

		// Return 404 for asset requests that don't exist
		if strings.HasPrefix(path, "/assets/") {
			c.String(http.StatusNotFound, "Asset not found")
			return
		}

		// Don't serve index.html for static asset requests
		if strings.HasPrefix(path, "/assets/") {
			c.String(http.StatusNotFound, "Asset not found")
			return
		}

		data, err := web.StaticFS.ReadFile("dist/index.html")
		if err != nil {
			c.String(http.StatusNotFound, "Frontend not found")
			return
		}
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, string(data))
	})
}
