package keycloakopenid

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	KeycloakURL      string `json:"url"`
	ClientID         string `json:"client_id"`
	ClientSecret     string `json:"client_secret"`
	KeycloakRealm    string `json:"keycloak_realm"`
	ClientIDFile     string `json:"client_id_file"`
	ClientSecretFile string `json:"client_secret_file"`
}

type keycloakAuth struct {
	next          http.Handler
	KeycloakURL   *url.URL
	ClientID      string
	ClientSecret  string
	KeycloakRealm string
}

type KeycloakTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type state struct {
	RedirectURL string `json:"redirect_url"`
}

func CreateConfig() *Config {
	return &Config{}
}

func parseUrl(rawUrl string) (*url.URL, error) {
	if rawUrl == "" {
		return nil, errors.New("invalid empty url")
	}
	if !strings.Contains(rawUrl, "://") {
		rawUrl = "https://" + rawUrl
	}
	u, err := url.Parse(rawUrl)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(u.Scheme, "http") {
		return nil, fmt.Errorf("%v is not a valid scheme", u.Scheme)
	}
	return u, nil
}

func readSecretFiles(config *Config) error {
	if config.ClientIDFile != "" {
		clientId, err := os.ReadFile(config.ClientIDFile)
		if err != nil {
			return err
		}
		config.ClientID = string(clientId)
	}
	if config.ClientSecretFile != "" {
		clientSecret, err := os.ReadFile(config.ClientSecretFile)
		if err != nil {
			return err
		}
		config.ClientSecret = string(clientSecret)
	}
	return nil
}

func New(uctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	err := readSecretFiles(config)
	if err != nil {
		return nil, err
	}
	if config.ClientID == "" || config.KeycloakRealm == "" {
		return nil, errors.New("invalid configuration")
	}

	parsedURL, err := parseUrl(config.KeycloakURL)
	if err != nil {
		return nil, err
	}

	return &keycloakAuth{
		next:          next,
		KeycloakURL:   parsedURL,
		ClientID:      config.ClientID,
		ClientSecret:  config.ClientSecret,
		KeycloakRealm: config.KeycloakRealm,
	}, nil
}
