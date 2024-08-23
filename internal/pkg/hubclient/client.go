package hubclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Auth struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token"`
}

type Client struct {
	BaseURL     string
	auth        Auth
	HTTPClient  *http.Client
	userAgent   string
	token       string
	tokenExpiry time.Time
	mu          sync.Mutex
}

type Config struct {
	Host             string
	Username         string
	Password         string
	UserAgentVersion string
}

func NewClient(config Config) *Client {
	version := config.UserAgentVersion
	if version == "" {
		version = "dev"
	}

	return &Client{
		BaseURL: config.Host,
		auth: Auth{
			Username: config.Username,
			Password: config.Password,
		},
		HTTPClient: &http.Client{
			Timeout: time.Minute,
		},
		userAgent: fmt.Sprintf("terraform-provider-docker/%s", version),
	}
}

// parseTokenExpiration parses the JWT token to get the exact expiration time.
func parseTokenExpiration(tokenString string) (time.Time, error) {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Time{}, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			return time.Unix(int64(exp), 0), nil
		}
	}

	return time.Time{}, fmt.Errorf("could not find expiration in token")
}

func (c *Client) ensureValidToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.token != "" && time.Now().Before(c.tokenExpiry) {
		return nil
	}

	authJSON, err := json.Marshal(c.auth)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/users/login/", c.BaseURL), bytes.NewBuffer(authJSON))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("HTTP error: %s", res.Status)
	}

	token := Token{}
	if err = json.NewDecoder(res.Body).Decode(&token); err != nil {
		return err
	}

	// Parse the exact expiration time from the token
	expirationTime, err := parseTokenExpiration(token.Token)
	if err != nil {
		return err
	}

	// Store the new token and its exact expiration time
	c.token = token.Token
	c.tokenExpiry = expirationTime

	return nil
}

func (c *Client) sendRequest(ctx context.Context, method string, url string, body []byte, result interface{}) error {
	if err := c.ensureValidToken(ctx); err != nil {
		return err
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s%s", c.BaseURL, url), bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	req = req.WithContext(ctx)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode < http.StatusOK || res.StatusCode >= http.StatusBadRequest {
		bodyBytes, readErr := io.ReadAll(res.Body)
		if readErr != nil {
			return readErr
		}
		return errors.New(string(bodyBytes))
	}

	if result != nil {
		if err = json.NewDecoder(res.Body).Decode(result); err != nil {
			return err
		}
	}

	return nil
}
