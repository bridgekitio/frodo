package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bridgekitio/frodo/cli"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "frodo",
		Short: "A code generator for Go-based (micro)services that let you consume them in an RPC or event-based fashion.",
	}
	rootCmd.AddCommand(cli.GenerateServer{}.Command())
	rootCmd.AddCommand(cli.GenerateClient{}.Command())
	rootCmd.AddCommand(cli.GenerateMock{}.Command())
	rootCmd.AddCommand(cli.GenerateDocs{}.Command())
	// rootCmd.AddCommand(cli.CreateService{}.Command())

	log.SetFlags(0)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
