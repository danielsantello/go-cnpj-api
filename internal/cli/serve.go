package cli

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/danielsantello/go-cnpj-api/internal/httpapi"
)

func serve() error {
	server := httpapi.NewServer(":8080")

	fmt.Printf("server listening on %s\n", server.Addr)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("start HTTP server: %w", err)
	}

	return nil
}
