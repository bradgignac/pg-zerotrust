package cmd

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/spf13/cobra"
)

var root = &cobra.Command{
	Use:  "pg-zerotrust",
	RunE: run,
}

func init() {
	root.Version = "0.1.0"

	root.Flags().StringP("port", "p", "", "port to bind to")
	root.Flags().StringP("upstream", "u", "", "upstream address to bind to")

	root.MarkFlagRequired("port")
	root.MarkFlagRequired("upstream")
}

func run(cmd *cobra.Command, args []string) error {
	port, err := cmd.Flags().GetString("port")
	if err != nil {
		return err
	}

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

func handle(conn net.Conn) {
	defer conn.Close()

	log.Printf("Accepted connection")
}

func Execute() {
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
