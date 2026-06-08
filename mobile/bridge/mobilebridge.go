package mobilebridge

import (
	"fmt"
	"sync"

	"passworder/internal/config"
	"passworder/internal/embedded"
)

var (
	mu            sync.Mutex
	runningServer *embedded.EmbeddedServer
)

func StartServer(host string, port int, dbPath, storagePath string) string {
	mu.Lock()
	defer mu.Unlock()

	if runningServer != nil {
		return ""
	}

	server, err := embedded.NewEmbeddedServer(config.CLIOverrides{
		Host:       host,
		Port:       port,
		DBPath:     dbPath,
		StorageDir: storagePath,
	})
	if err != nil {
		return err.Error()
	}
	if err := server.Start(); err != nil {
		_ = server.Stop()
		return err.Error()
	}
	runningServer = server
	go func(s *embedded.EmbeddedServer) {
		if err := s.Wait(); err != nil {
			fmt.Printf("embedded server stopped with error: %v\n", err)
		}
	}(server)
	return ""
}

func StopServer() string {
	mu.Lock()
	defer mu.Unlock()
	if runningServer == nil {
		return ""
	}
	err := runningServer.Stop()
	runningServer = nil
	if err != nil {
		return err.Error()
	}
	return ""
}
