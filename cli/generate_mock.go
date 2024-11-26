package cli

import (
	"log"

	"github.com/bridgekit-io/frodo/generate"
	"github.com/bridgekit-io/frodo/parser"
	"github.com/spf13/cobra"
)

// GenerateMockRequest contains all of the CLI options used in the "frodo mock" command.
type GenerateMockRequest struct {
	templateOption
	// InputFileName is the service definition to parse/process (the "--service" option)
	InputFileName string
}

// GenerateMock handles the registration and execution of the 'frodo mock' CLI subcommand.
type GenerateMock struct{}

// Command creates the Cobra struct describing this CLI command and its options.
func (c GenerateMock) Command() *cobra.Command {
	request := &GenerateMockRequest{}
	cmd := &cobra.Command{
		Use:   "mock [flags] FILENAME",
		Short: "Creates a mock instance of your service for unit testing.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			request.InputFileName = args[0]
			crapPants(c.Exec(request))
		},
	}
	cmd.Flags().StringVar(&request.Template, "template", "", "Path to a custom Go template file used to generate this artifact.")
	cmd.Flags().BoolVar(&request.Force, "force", false, "Ignore file modification timestamps and generate the artifact no matter what.")
	return cmd
}

// Exec takes all of the parsed CLI flags and generates the target mock service artifact.
func (c GenerateMock) Exec(request *GenerateMockRequest) error {
	artifact := request.ToFileTemplate("mock.go")

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
