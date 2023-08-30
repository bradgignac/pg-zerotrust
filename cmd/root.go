package cmd

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const BufferSize int = 1

var port string
var upstream string

var root = &cobra.Command{
	Use:  "pg-zerotrust",
	RunE: run,
}

func init() {
	root.Version = "0.1.0"

	root.Flags().StringVarP(&port, "port", "p", "", "port to bind to")
	root.Flags().StringVarP(&upstream, "upstream", "u", "", "upstream address to bind to")

	root.MarkFlagRequired("port")
	root.MarkFlagRequired("upstream")
}

func run(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf(":%s", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	log.Printf("Listening for connections on %s", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
		}

		go handle(conn)
	}
}

func handle(client net.Conn) {
	defer client.Close()

	log.Printf("Received new client connection. Initiating upstream connection to %s", upstream)

	server, err := net.Dial("tcp", upstream)
	if err != nil {
		log.Printf("Failed to create upstream connection to %s", upstream)
		return
	}

	defer server.Close()

	group := errgroup.Group{}
	group.Go(func() error {
		for {
			buffer := make([]byte, BufferSize)
			n, err := io.ReadFull(client, buffer)

			if err == io.EOF {
				return nil
			} else if err != nil && err != io.ErrUnexpectedEOF {
				return err
			}

			log.Printf("Successfully read %d bytes from client: %v", n, buffer)

			n, err = server.Write(buffer)
			if err != nil {
				return err
			}

			log.Printf("Successfully wrote %d bytes to server: %v", n, buffer)
		}
	})
	group.Go(func() error {
		for {
			buffer := make([]byte, BufferSize)
			n, err := io.ReadFull(server, buffer)

			if err == io.EOF {
				return nil
			} else if err != nil && err != io.ErrUnexpectedEOF {
				return err
			}

			log.Printf("Successfully read %d bytes from server: %v", n, buffer)

			n, err = client.Write(buffer)
			if err != nil {
				return err
			}

			log.Printf("Successfully wrote %d bytes to client: %v", n, buffer)
		}
	})

	err = group.Wait()
	if err != nil {
		log.Println(err)
	}
}

func Execute() {
	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
