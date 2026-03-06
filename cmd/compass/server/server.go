package server

import (
	"crypto/tls"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
	middleware "github.com/oapi-codegen/gin-middleware"

	"github.com/complytime/gemara-content-service/api"
	httpmw "github.com/complytime/gemara-content-service/internal/middleware"
	"github.com/complytime/gemara-content-service/internal/oci"
	compass "github.com/complytime/gemara-content-service/service"
)

func NewGinServer(service *compass.Service, registry *oci.Registry, port string) *http.Server {
	swagger, err := api.GetSwagger()
	if err != nil {
		slog.Error("Error loading swagger spec", "err", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestid.New(), httpmw.AccessLogger())

	// OCI Distribution routes are registered directly on the router without
	// OpenAPI validation. The catch-all path parameter used for repository
	// names containing slashes cannot be validated by the OpenAPI spec.
	if registry != nil {
		oci.RegisterRoutes(r, registry)
	}

	// Enrichment routes use OpenAPI request validation.
	enrichGroup := r.Group("")
	enrichGroup.Use(middleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(enrichGroup, service)

	s := &http.Server{
		Handler:           r,
		Addr:              net.JoinHostPort("0.0.0.0", port),
		ReadHeaderTimeout: 10 * time.Second,
	}

	return s
}

func SetupTLS(server *http.Server, config Config) (string, string) {
	// TODO: Allow loosening here through configuration
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS13}
	server.TLSConfig = tlsConfig

	if config.Certificate.PublicKey == "" {
		slog.Error("Invalid certification configuration. Please add certConfig.cert to the configuration.")
		os.Exit(1)
	}

	if config.Certificate.PrivateKey == "" {
		slog.Error("Invalid certification configuration. Please add certConfig.key to the configuration.")
		os.Exit(1)
	}

	return config.Certificate.PublicKey, config.Certificate.PrivateKey
}
