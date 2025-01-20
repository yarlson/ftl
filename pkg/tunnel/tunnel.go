package tunnel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/yarlson/ftl/pkg/ssh"
)

// TunnelConfig describes which local port should forward to which remote address.
type TunnelConfig struct {
	LocalPort  string
	RemoteAddr string
}

// StartTunnels spawns one goroutine per tunnel, each calling ssh.CreateSSHTunnel.
func StartTunnels(
	ctx context.Context,
	host string,
	port int,
	user, sshKey string,
	tunnels []TunnelConfig,
) error {
	if len(tunnels) == 0 {
		return fmt.Errorf("no tunnels to establish")
	}

	var wg sync.WaitGroup
	errorChan := make(chan error, len(tunnels))

	for _, t := range tunnels {
		wg.Add(1)
		go func(tun TunnelConfig) {
			defer wg.Done()

			err := ssh.CreateSSHTunnel(ctx, host, port, user, sshKey, tun.LocalPort, tun.RemoteAddr)
			if err != nil {
				errorChan <- fmt.Errorf("tunnel %s -> %s failed: %v",
					tun.LocalPort, tun.RemoteAddr, err)
			}
		}(t)
	}

	go func() {
		wg.Wait()
		close(errorChan)
	}()

	select {
	case err := <-errorChan:
		if err != nil {
			return err
		}
	case <-time.After(2 * time.Second):
		return nil
	}

	return nil
}
