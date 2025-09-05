package server

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/sirrobot01/protodex/internal/client"
	"github.com/sirrobot01/protodex/internal/manager"
	"github.com/sirrobot01/protodex/internal/server/auth"
)

const AuthContextKey = "auth_context"

type APIHandler struct {
}

func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		token := auth.ExtractTokenFromHeader(authHeader)

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		authCtx, err := s.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		c.Set(AuthContextKey, authCtx)
		c.Next()
	}
}

func (s *Server) getAuthContext(c *gin.Context) (*auth.Context, error) {
	if ctx, exists := c.Get(AuthContextKey); exists {
		if authCtx, ok := ctx.(*auth.Context); ok {
			return authCtx, nil
		}
	}
	return nil, fmt.Errorf("no auth context")
}

// Auth handlers
func (s *Server) loginHandler(c *gin.Context) {
	var req client.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userAgent := c.GetHeader("User-Agent")

	loginResp, err := s.authService.Login(userAgent, req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	response := client.LoginResponse{
		Token: loginResp.Token,
		User: client.User{
			ID:       loginResp.User.ID,
			Username: loginResp.User.Username,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (s *Server) registerHandler(c *gin.Context) {
	var req client.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	_, err := s.authService.GetUserByUsername(req.Username)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// Create user
	user, err := s.authService.CreateUser(req.Username, req.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	userAgent := c.GetHeader("User-Agent")

	// Create login session
	loginResp, err := s.authService.Login(userAgent, req.Username, req.Password)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to login after registration")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Registration succeeded but login failed"})
		return
	}

	response := client.RegisterResponse{
		Token: loginResp.Token,
		User: client.User{
			ID:       user.ID,
			Username: user.Username,
		},
	}

	c.JSON(http.StatusCreated, response)
}

func (s *Server) getCurrentUserHandler(c *gin.Context) {
	authCtx, err := s.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user := client.User{
		ID:       authCtx.User.ID,
		Username: authCtx.User.Username,
	}

	c.JSON(http.StatusOK, user)
}

// Package handlers
func (s *Server) listPackagesHandler(c *gin.Context) {
	packages, err := s.packageStore.ListPackages()
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to list packages")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var clientPackages []*client.Package
	for _, pkg := range packages {
		clpkg := &client.Package{
			ID:          pkg.ID,
			Name:        pkg.Name,
			Description: pkg.Description,
			Tags:        pkg.Tags,
			CreatedAt:   pkg.CreatedAt,
			OwnerID:     pkg.OwnerID,
		}
		clientPackages = append(clientPackages, clpkg)
	}

	c.JSON(http.StatusOK, clientPackages)
}

func (s *Server) getPackageHandler(c *gin.Context) {
	packageName := c.Param("package")
	pkg, err := s.packageStore.GetPackage(packageName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}
	clientPackage := &client.Package{
		ID:          pkg.ID,
		Name:        pkg.Name,
		Description: pkg.Description,
		Tags:        pkg.Tags,
		CreatedAt:   pkg.CreatedAt,
		OwnerID:     pkg.OwnerID,
	}

	c.JSON(http.StatusOK, clientPackage)
}

func (s *Server) createPackageHandler(c *gin.Context) {
	authCtx, err := s.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pkg, err := s.packageStore.CreatePackage(req.Name, req.Description, authCtx.UserID, req.Tags)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create package")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clientPackage := &client.Package{
		ID:          pkg.ID,
		Name:        pkg.Name,
		Description: pkg.Description,
		Tags:        pkg.Tags,
		CreatedAt:   pkg.CreatedAt,
	}

	c.JSON(http.StatusCreated, clientPackage)
}

func (s *Server) searchPackagesHandler(c *gin.Context) {
	query := c.Query("q")
	tags := c.QueryArray("tags")

	packages, err := s.packageStore.SearchPackages(query, tags)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to search packages")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var clientPackages []*client.Package
	for _, pkg := range packages {
		clientPackages = append(clientPackages, &client.Package{
			ID:          pkg.ID,
			Name:        pkg.Name,
			Description: pkg.Description,
			Tags:        pkg.Tags,
			CreatedAt:   pkg.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, clientPackages)
}

// pushVersionZipHandler handles uploading a ZIP file containing schema files for a specific package version
func (s *Server) pushVersionHandler(c *gin.Context) {
	authCtx, err := s.getAuthContext(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	packageName := c.Param("package")
	version := c.PostForm("version")

	if version == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "version is required"})
		return
	}

	// Get uploaded zip file
	fileHeader, err := c.FormFile("zip")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "zip file is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open zip file"})
		return
	}
	defer file.Close()

	zipData, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read zip file"})
		return
	}

	// Get or create package
	pkg, err := s.packageStore.GetPackage(packageName)
	if err != nil {
		pkg, err = s.packageStore.CreatePackage(packageName, "", authCtx.UserID, []string{})
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to create package")
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Extract and validate zip contents
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid zip file"})
		return
	}

	// Create schema directory
	schemaDir := s.packageStore.GetSchemaPath(packageName, version)
	if err := os.MkdirAll(schemaDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create schema directory"})
		return
	}

	var filePaths []string
	var allContent []byte
	var hasProtodexYaml, hasProtoFiles bool

	// Extract files from zip
	for _, zipFile := range zipReader.File {
		// Skip directories
		if zipFile.FileInfo().IsDir() {
			continue
		}

		rc, err := zipFile.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to open zip entry %s", zipFile.Name)})
			return
		}

		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to read zip entry %s", zipFile.Name)})
			return
		}

		// Track file types for validation
		if zipFile.Name == "protodex.yaml" {
			hasProtodexYaml = true
		}
		if strings.HasSuffix(zipFile.Name, ".proto") {
			hasProtoFiles = true
		}

		allContent = append(allContent, content...)

		// Write file to schema directory maintaining structure
		filePath := filepath.Join(schemaDir, zipFile.Name)
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file directory"})
			return
		}

		if err := os.WriteFile(filePath, content, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to save file %s", zipFile.Name)})
			return
		}

		filePaths = append(filePaths, filePath)
	}

	// Validate package structure
	if !hasProtodexYaml {
		// Clean up created files
		if err := os.RemoveAll(schemaDir); err != nil {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "zip must contain protodex.yaml file"})
		return
	}

	if !hasProtoFiles {
		// Clean up created files

		if err := os.RemoveAll(schemaDir); err != nil {
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "zip must contain at least one .proto file"})
		return
	}

	if len(filePaths) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no files provided in zip"})
		return
	}

	// Calculate checksum
	hasher := sha256.New()
	hasher.Write(allContent)
	checksum := hex.EncodeToString(hasher.Sum(nil))

	// Don't save ZIP archive - generate on-demand for pulls

	// Store schema files in database
	schemaVersion, err := s.packageStore.SaveSchemaFiles(pkg.ID, version, filePaths, authCtx.UserID)
	if err != nil {
		// Clean up created files
		os.RemoveAll(schemaDir)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update version with checksum
	schemaVersion.Checksum = checksum

	clientVersion := &client.Version{
		ID:        schemaVersion.ID,
		Version:   schemaVersion.Version,
		CreatedAt: schemaVersion.CreatedAt,
		CreatedBy: schemaVersion.CreatedBy,
		Checksum:  checksum,
	}

	c.JSON(http.StatusCreated, clientVersion)
}

func (s *Server) listVersionsHandler(c *gin.Context) {
	packageName := c.Param("package")

	pkg, err := s.packageStore.GetPackage(packageName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	versions, err := s.packageStore.ListVersions(pkg.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clientVersions := make([]client.Version, 0, len(versions))
	for _, ver := range versions {
		clientVersions = append(clientVersions, client.Version{
			ID:        ver.ID,
			Version:   ver.Version,
			CreatedAt: ver.CreatedAt,
			CreatedBy: ver.CreatedBy,
		})
	}

	c.JSON(http.StatusOK, clientVersions)
}

func (s *Server) pullVersionHandler(c *gin.Context) {
	packageName := c.Param("package")
	version := c.Param("version")

	pkg, err := s.packageStore.GetPackage(packageName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	schemaVersion, err := s.packageStore.GetSchemaVersion(pkg.ID, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	if len(schemaVersion.Files) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no files found in schema version"})
		return
	}

	// Create ZIP archive on-demand
	zipBuffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(zipBuffer)

	for _, file := range schemaVersion.Files {
		// Handle both relative and absolute paths
		filePath := file.FilePath
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(s.packageStore.GetDataDir(), filePath)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			s.logger.Error().Err(err).Str("file", filePath).Msg("Failed to read file")
			continue
		}

		// Use the original zip structure (relative to schema root)
		zipEntryPath := file.Filename
		if strings.Contains(file.FilePath, "schemas/") {
			parts := strings.Split(file.FilePath, "schemas/")
			if len(parts) > 1 {
				schemaParts := strings.SplitN(parts[1], "/", 3)
				if len(schemaParts) >= 3 {
					zipEntryPath = schemaParts[2]
				}
			}
		}

		f, err := zipWriter.Create(zipEntryPath)
		if err != nil {
			s.logger.Error().Err(err).Str("filename", zipEntryPath).Msg("Failed to create zip entry")
			continue
		}

		if _, err := f.Write(content); err != nil {
			s.logger.Error().Err(err).Str("filename", file.Filename).Msg("Failed to write zip entry")
			continue
		}
	}

	if err := zipWriter.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create zip archive"})
		return
	}

	filename := fmt.Sprintf("%s-%s.zip", packageName, version)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/zip", zipBuffer.Bytes())
}

// viewSchemaHandler returns schema information and file contents for viewing
func (s *Server) viewSchemaHandler(c *gin.Context) {
	packageName := c.Param("package")
	version := c.Param("version")

	pkg, err := s.packageStore.GetPackage(packageName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	schemaVersion, err := s.packageStore.GetSchemaVersion(pkg.ID, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	if len(schemaVersion.Files) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no files found in schema version"})
		return
	}

	type FileContent struct {
		Name    string `json:"name"`
		Path    string `json:"path"`
		Content string `json:"content"`
		Size    int64  `json:"size"`
	}

	type SchemaView struct {
		Package     string        `json:"package"`
		Version     string        `json:"version"`
		Description string        `json:"description"`
		Checksum    string        `json:"checksum"`
		CreatedAt   string        `json:"created_at"`
		CreatedBy   string        `json:"created_by"`
		Files       []FileContent `json:"files"`
	}

	var files []FileContent
	for _, file := range schemaVersion.Files {
		// Handle both relative and absolute paths
		filePath := file.FilePath
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(s.packageStore.GetDataDir(), filePath)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			s.logger.Error().Err(err).Str("file", filePath).Msg("Failed to read file")
			continue
		}

		// Extract file path relative to schema directory for hierarchy display
		displayPath := file.Filename // default to filename

		// If FilePath is stored as relative from data dir, extract the schema-relative part
		if strings.Contains(file.FilePath, "schemas/") {
			parts := strings.Split(file.FilePath, "schemas/")
			if len(parts) > 1 {
				// Further split to get just the file path within the schema
				schemaParts := strings.SplitN(parts[1], "/", 3) // package/version/filepath
				if len(schemaParts) >= 3 {
					displayPath = schemaParts[2]
				}
			}
		} else if !filepath.IsAbs(file.FilePath) {
			// If already relative, use as-is
			displayPath = file.FilePath
		}

		files = append(files, FileContent{
			Name:    file.Filename,
			Path:    displayPath,
			Content: string(content),
			Size:    file.SizeBytes,
		})
	}

	schema := SchemaView{
		Package:     packageName,
		Version:     version,
		Description: pkg.Description,
		Checksum:    schemaVersion.Checksum,
		CreatedAt:   schemaVersion.CreatedAt.Format("2006-01-02 15:04:05"),
		CreatedBy:   schemaVersion.CreatedBy,
		Files:       files,
	}

	c.JSON(http.StatusOK, schema)
}

func (s *Server) generateCodeHandler(c *gin.Context) {
	packageName := c.Param("package")
	version := c.Param("version")

	var req client.GenerateOptions
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	language := c.Query("language")
	if language == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "language is required"})
		return
	}

	// Get package and version from registry
	pkg, err := s.packageStore.GetPackage(packageName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "package not found"})
		return
	}

	schemaVersion, err := s.packageStore.GetSchemaVersion(pkg.ID, version)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "version not found"})
		return
	}

	if len(schemaVersion.Files) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no files found in schema version"})
		return
	}

	// Create temporary directory for generation
	tempDir, err := os.MkdirTemp("", "protodex-generate-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create temporary directory"})
		return
	}
	defer os.RemoveAll(tempDir)

	// Copy proto files to temp directory
	var protoFiles []string
	for _, file := range schemaVersion.Files {
		if !strings.HasSuffix(file.Filename, ".proto") {
			continue // Skip non-proto files
		}

		// Handle both relative and absolute paths
		filePath := file.FilePath
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(s.packageStore.GetDataDir(), filePath)
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			s.logger.Error().Err(err).Str("file", filePath).Msg("Failed to read proto file")
			continue
		}

		tempFilePath := filepath.Join(tempDir, file.Filename)
		if err := os.MkdirAll(filepath.Dir(tempFilePath), 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file directory"})
			return
		}

		if err := os.WriteFile(tempFilePath, content, 0644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write proto file"})
			return
		}

		protoFiles = append(protoFiles, tempFilePath)
	}

	if len(protoFiles) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no proto files found in package"})
		return
	}

	// Initialize manager for generation
	pm, err := manager.NewManager(tempDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize manager"})
		return
	}

	// Create output directory within temp dir
	outputDir := filepath.Join(tempDir, "generated")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create output directory"})
		return
	}

	// Setup language configuration
	options := manager.LanguageConfig{
		Name:      language,
		OutputDir: outputDir,
		Options:   make(map[string]string),
	}

	// Add package name if provided
	if req.PackageName != "" {
		options.Options["package"] = req.PackageName
	}

	// Generate code
	if err := pm.Generate(protoFiles, options); err != nil {
		s.logger.Error().Err(err).Msg("Code generation failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("code generation failed: %v", err)})
		return
	}

	var generatedFiles []string
	err = filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(outputDir, path)
			generatedFiles = append(generatedFiles, relPath)
		}
		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to collect generated files"})
		return
	}
	zipBuffer := &bytes.Buffer{}
	zipWriter := zip.NewWriter(zipBuffer)

	for _, file := range generatedFiles {
		fullPath := filepath.Join(outputDir, file)
		content, err := os.ReadFile(fullPath)
		if err != nil {
			continue
		}

		f, err := zipWriter.Create(file)
		if err != nil {
			continue
		}

		if _, err := f.Write(content); err != nil {
			continue
		}
	}

	if err := zipWriter.Close(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create zip archive"})
		return
	}

	// Return ZIP file
	filename := fmt.Sprintf("%s-%s-%s-generated.zip", packageName, version, language)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/zip", zipBuffer.Bytes())
}
