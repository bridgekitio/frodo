package cli

import (
	"log"

	"github.com/bridgekit-io/frodo/generate"
	"github.com/bridgekit-io/frodo/parser"
	"github.com/spf13/cobra"
)

// GenerateServerRequest contains all of the CLI options used in the "frodo client" command.
type GenerateServerRequest struct {
	templateOption
	// InputFileName is the service definition to parse/process (the "--service" option)
	InputFileName string
}

// GenerateServer handles the registration and execution of the 'frodo gateway' CLI subcommand.
type GenerateServer struct{}

// Command creates the Cobra struct describing this CLI command and its options.
func (c GenerateServer) Command() *cobra.Command {
	request := &GenerateServerRequest{}
	cmd := &cobra.Command{
		Use:   "server [flags] FILENAME",
		Short: "Process a Go source file with your service interface to generate gateway listener code for your server.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			request.InputFileName = args[0]
			crapPants(c.Exec(request))
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.Flags().StringVar(&request.Template, "template", "", "Path to a custom Go template file used to generate this artifact.")
	cmd.Flags().BoolVar(&request.Force, "force", false, "Ignore file modification timestamps and generate the artifact no matter what.")
	return cmd
}

// Exec actually executes the parsing/generating logic creating the gateway for the given declaration.
func (c GenerateServer) Exec(request *GenerateServerRequest) error {
	artifact := request.ToFileTemplate("server.go")

	if !request.Force && generate.UpToDate(request.InputFileName, artifact.Name) {
		log.Printf("Skipping '%s'. Artifact is up to date '%s'", request.InputFileName, artifact.Name)
		return nil
	}

	log.Printf("Parsing service definitions: %s", request.InputFileName)
	ctx, err := parser.ParseFile(request.InputFileName)
	if err != nil {
		return err
	}

	log.Printf("Generating artifact '%s'", artifact.Name)
	return generate.File(ctx, artifact)
}
