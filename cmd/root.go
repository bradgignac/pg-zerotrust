package cmd

import (
	"bradgignac/pg-zerotrust/internal/message"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const BufferSize int = 1
const StartSSLHandshake = "S"
const DoNotStartSSLHandshake = "N"

var port string
var upstreamAddr string

var root = &cobra.Command{
	Use:  "pg-zerotrust",
	RunE: run,
}

var logger *slog.Logger

func init() {
	root.Version = "0.1.0"

	root.Flags().StringVarP(&port, "port", "p", "", "port to bind to")
	root.Flags().StringVarP(&upstreamAddr, "upstream", "u", "", "upstream address to bind to")

	root.MarkFlagRequired("port")
	root.MarkFlagRequired("upstream")

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	logger = slog.New(handler)
}

func run(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	logger.Info("Listening for new connections", "port", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(err.Error())
		}

		go handle(conn)
	}
}

func handle(client net.Conn) {
	defer client.Close()

	logger.Info("Initiating new upstream connect for client", "upstream", upstreamAddr)

	upstream, err := net.Dial("tcp", upstreamAddr)
	if err != nil {
		logger.Error("Failed to create upstream connection", "upstream", upstreamAddr)
		return
	}
	defer upstream.Close()

	// TODO: proxy + negotiate can be wrapped into a single function since
	// client-to-proxy and proxy-to-upstream SSL are totally independent
	err = proxySSLRequest(client, upstream)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	err = negotiateSSL(client, upstream)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	err = proxyStartupMessage(client, upstream)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	group := errgroup.Group{}
	group.Go(func() error {
		for {
			if err := proxyClientMessageToUpstream(client, upstream); err != nil {
				return err
			}
		}
	})
	group.Go(func() error {
		for {
			if err := proxyUpstreamMessageToClient(client, upstream); err != nil {
				return err
			}
		}
	})

	err = group.Wait()
	if err != nil {
		logger.Error(err.Error())
	}
}

func proxySSLRequest(client io.Reader, upstream io.WriteCloser) error {
	msg, err := message.ParseSSLRequest(client)
	if err != nil {
		return err
	}

	n, err := upstream.Write(msg.Bytes())
	if err != nil {
		return err
	}

	logger.Info("Successfully proxied SSLRequest to upstream", "msg", msg, "bytes", n)

	return nil
}

func negotiateSSL(client io.Writer, upstream io.Reader) error {
	buffer := make([]byte, 1)
	_, err := io.ReadFull(upstream, buffer)

	if err != nil {
		return err
	} else if string(buffer) == StartSSLHandshake {
		return fmt.Errorf("SSL is not supported")
	}

	n, err := client.Write(buffer)
	if err != nil {
		return err
	}

	logger.Info("Successfully negotiated SSLRequest", "msg", buffer, "bytes", n)

	return nil
}

func proxyStartupMessage(client io.Reader, upstream io.Writer) error {
	msg, err := message.ParseStartupMessage(client)
	if err != nil {
		return err
	}

	n, err := upstream.Write(msg.Bytes())
	if err != nil {
		return err
	}

	logger.Info("Successfully proxied StartupMessage to upstream", "msg", msg, "bytes", n)

	return nil
}

func proxyClientMessageToUpstream(client io.Reader, upstream io.Writer) error {
	msg, err := message.ParseMessage(client)
	if err != nil {
		return err
	}

	n, err := upstream.Write(msg.Bytes())
	if err != nil {
		return err
	}

	logger.Info("Successfully proxied message to upstream", "bytes", n, "msg", msg)

	return nil
}

func proxyUpstreamMessageToClient(client io.Writer, upstream io.Reader) error {
	msg, err := message.ParseMessage(upstream)
	if err != nil {
		return err
	}

	n, err := client.Write(msg.Bytes())
	if err != nil {
		return err
	}

	logger.Info("Successfully proxied message to client", "bytes", n, "msg", msg)

	return nil
}

func Execute() {
	if err := root.Execute(); err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}
