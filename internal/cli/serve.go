package cli

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/danielsantello/go-cnpj-api/internal/config"
	"github.com/danielsantello/go-cnpj-api/internal/httpapi"
)

func serve() error {
	configuration := config.Load()

	if err := config.Validate(configuration); err != nil {
		return fmt.Errorf("validate configuration: %w", err)
	}

	server := httpapi.NewServer(configuration.HTTPAddress)

	fmt.Printf("server listening on %s\n", server.Addr)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start HTTP server: %w", err)
	}

	return nil
}
