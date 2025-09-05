package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/sirrobot01/protodex/internal/config"
	"github.com/sirrobot01/protodex/internal/logger"
)

type Client interface {
	Login(username, password string) (*LoginResponse, error)
	Register(username, password string) (*RegisterResponse, error)
	GetCurrentUser() (*User, error)
	Logout() error

	ListPackages() ([]*Package, error)
	GetPackage(name string) (*Package, error)
	CreatePackage(name, description string, tags []string) (*Package, error)
	SearchPackages(query string, tags []string) ([]*Package, error)

	PushVersion(packageName, version string, zipData []byte) (*Version, error)
	PullVersion(packageName, version, outputDir string) error
	ListVersions(packageName string) ([]*Version, error)
	ViewSchema(packageName, version string) (*SchemaView, error)

	GenerateCode(packageName, version, language, outputDir string, options GenerateOptions) (*GenerateResult, error)
}

type HTTPClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
	config     *config.Config
	logger     zerolog.Logger
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type Package struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	OwnerID     string    `json:"owner_id"`
}

type Version struct {
	ID        string    `json:"id"`
	Version   string    `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	CreatedBy string    `json:"created_by"`
	Metadata  string    `json:"metadata,omitempty"`
	Checksum  string    `json:"checksum,omitempty"`
}

type GenerateOptions struct {
	PackageName string `json:"package_name,omitempty"`
	ModulePath  string `json:"module_path,omitempty"`
}

type GenerateResult struct {
	OutputDir      string   `json:"output_dir"`
	GeneratedFiles []string `json:"generated_files"`
}

type FileContent struct {
	Name    string `json:"name"`
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

func New() (Client, error) {
	cfg := config.Get()
	return &HTTPClient{
		baseURL:    cfg.Registry,
		token:      cfg.HashedToken,
		httpClient: &http.Client{},
		logger:     logger.Get(),
		config:     cfg,
	}, nil
}

func NewWithToken(token string) (Client, error) {
	cfg := config.Get()
	return &HTTPClient{
		baseURL:    cfg.Registry,
		token:      token,
		httpClient: &http.Client{},
		logger:     logger.Get(),
		config:     cfg,
	}, nil
}

func (c *HTTPClient) do(req *http.Request) (*http.Response, error) {
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	return c.httpClient.Do(req)
}

func (c *HTTPClient) Login(username, password string) (*LoginResponse, error) {
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/auth/login", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to close response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed: %s - %s", resp.Status, string(body))
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &loginResp, nil
}

func (c *HTTPClient) Register(username, password string) (*RegisterResponse, error) {
	registerReq := RegisterRequest{
		Username: username,
		Password: password,
	}

	jsonData, err := json.Marshal(registerReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/auth/register", c.baseURL)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to close response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("register failed: %s - %s", resp.Status, string(body))
	}

	var registerResp RegisterResponse
	if err := json.NewDecoder(resp.Body).Decode(&registerResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Save token to config
	c.token = registerResp.Token
	c.config.HashedToken = registerResp.Token
	if err := c.config.Save(); err != nil {
		return nil, fmt.Errorf("failed to save token: %w", err)
	}

	return &registerResp, nil
}

func (c *HTTPClient) GetCurrentUser() (*User, error) {
	url := fmt.Sprintf("%s/api/auth/me", c.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			c.logger.Error().Err(err).Msg("failed to close response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: %s", resp.Status)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &user, nil
}

func (c *HTTPClient) Logout() error {
	c.token = ""
	c.config.HashedToken = ""
	return c.config.Save()
}
