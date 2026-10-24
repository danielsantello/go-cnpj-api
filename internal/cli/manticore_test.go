package cli

import (
	"net"
	"testing"
	"time"

	"github.com/danielsantello/go-cnpj-api/internal/config"
)

func TestManticoreStartupPolicies(t *testing.T) {
	configuration := config.Default()
	pool, err := initializeManticoreForServer(configuration)
	if pool != nil || err != nil {
		t.Fatalf("disabled server integration must not open a pool: %v", err)
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := uint16(listener.Addr().(*net.TCPAddr).Port)
	listener.Close()
	configuration.ManticoreEnabled = true
	configuration.ManticorePort = port
	configuration.ManticoreConnectTimeout = 100 * time.Millisecond

	pool, err = initializeManticoreForServer(configuration)
	if err != nil || pool == nil {
		t.Fatalf("server must retain its pool when unreachable: %v", err)
	}
	defer pool.Close()

	strictPool, err := initializeManticore(configuration)
	if err == nil || strictPool != nil {
		t.Fatal("indexing commands must reject an unreachable Manticore")
	}
}
