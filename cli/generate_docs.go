package cli

import (
	"log"

	"github.com/bridgekitio/frodo/generate"
	"github.com/bridgekitio/frodo/parser"
	"github.com/spf13/cobra"
)

// GenerateDocsRequest contains all of the CLI options used in the "frodo docs" command.
type GenerateDocsRequest struct {
	templateOption
	// InputFileName is the service definition to parse/process (the "--service" option)
	InputFileName string
}

// GenerateDocs handles the registration and execution of the 'frodo docs' CLI subcommand.
type GenerateDocs struct{}

// Command creates the Cobra struct describing this CLI command and its options.
func (c GenerateDocs) Command() *cobra.Command {
	request := &GenerateDocsRequest{}
	cmd := &cobra.Command{
		Use:   "docs [flags] FILENAME",
		Short: "Generates the OpenAPI documentation for your service that can be distributed to users.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			request.InputFileName = args[0]
			crapPants(c.Exec(request))
		},
	}
	cmd.Flags().StringVar(&request.Template, "template", "", "Path to a custom OpenAPI/Swagger/docs template file used to generate this artifact.")
	cmd.Flags().BoolVar(&request.Force, "force", false, "Ignore file modification timestamps and generate the artifact no matter what.")
	return cmd
}

// Exec takes all of the parsed CLI flags and generates the service's documentation artifact(s).
func (c GenerateDocs) Exec(request *GenerateDocsRequest) error {
	artifact := request.ToFileTemplate("openapi.yml")

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
