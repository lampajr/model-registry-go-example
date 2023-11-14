package cmd

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lampajr/model-registry-go-example/internal/utils"
	"github.com/opendatahub-io/model-registry/pkg/api"
	"github.com/opendatahub-io/model-registry/pkg/core"
	"github.com/opendatahub-io/model-registry/pkg/openapi"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ISParams struct {
	Runtime       string
	ModelArtifact *openapi.ModelArtifact
}

var (
	//go:embed templates/*.yaml.tmpl
	templateFS embed.FS
	// inferenceCmd represents the create-inference command
	inferenceCmd = &cobra.Command{
		Use:   "create-inference",
		Short: "Create an InferenceService custom resource yaml on stdout",
		Long: `This command creates an InferenceService CR yaml on the standard output.
		
	The information required to create that IS is fetched using Model Registry library.`,
		RunE: createInference,
	}
)

func createInference(cmd *cobra.Command, args []string) error {
	// setup grpc connection to ml-metadata
	ctxTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	mlmdAddr := fmt.Sprintf("%s:%d", mlmdHostname, mlmdPort)
	conn, err := grpc.DialContext(
		ctxTimeout,
		mlmdAddr,
		grpc.WithReturnConnectionError(),
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("error dialing connection to mlmd server %s: %v", mlmdAddr, err)
	}
	defer conn.Close()

	service, err := core.NewModelRegistryService(conn)
	if err != nil {
		return fmt.Errorf("error creating core service: %v", err)
	}

	// TODO: move this CR generation in the reconcile.go command
	params, err := getInferenceServiceParams(
		service,
		inferenceServiceCfg.RegisteredModelName,
		inferenceServiceCfg.ModelVersionName,
		inferenceServiceCfg.ModelArtifactName,
		inferenceServiceCfg.ServingEnvironmentName,
	)
	if err != nil {
		return fmt.Errorf("error getting inference service information: %v", err)
	}

	tmpl, err := utils.ParseTemplate(templateFS)
	if err != nil {
		log.Fatal("error parsing templates")
	}

	tmpl.Execute(os.Stdout, params)

	return nil
}

func getInferenceServiceParams(service api.ModelRegistryApi, registeredModelName string, modelVersionName string, modelArtifactName string, servingEnvironmentName string) (ISParams, error) {
	registeredModel, err := service.GetRegisteredModelByParams(&registeredModelName, nil)
	if err != nil {
		return ISParams{}, err
	}

	modelVersion, err := service.GetModelVersionByParams(&modelVersionName, registeredModel.Id, nil)
	if err != nil {
		return ISParams{}, err
	}

	modelArtifact, err := service.GetModelArtifactByParams(&modelArtifactName, modelVersion.Id, nil)
	if err != nil {
		return ISParams{}, err
	}

	servingEnvironment, err := service.GetServingEnvironmentByParams(&servingEnvironmentName, nil)
	if err != nil {
		return ISParams{}, err
	}

	return ISParams{
		Runtime:       *servingEnvironment.Name,
		ModelArtifact: modelArtifact,
	}, nil
}

func init() {
	rootCmd.AddCommand(inferenceCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	inferenceCmd.Flags().StringVar(&inferenceServiceCfg.RegisteredModelName, "model", inferenceServiceCfg.RegisteredModelName, "Registered model name")
	inferenceCmd.Flags().StringVar(&inferenceServiceCfg.ModelVersionName, "version", inferenceServiceCfg.ModelVersionName, "Specific model version name")
	inferenceCmd.Flags().StringVar(&inferenceServiceCfg.ModelArtifactName, "artifact", inferenceServiceCfg.ModelArtifactName, "Model artifact name")
	inferenceCmd.Flags().StringVar(&inferenceServiceCfg.ServingEnvironmentName, "runtime", inferenceServiceCfg.ServingEnvironmentName, "Serving environment runtime name")
}

type InferenceServiceInputConfig struct {
	RegisteredModelName    string
	ModelVersionName       string
	ModelArtifactName      string
	ServingEnvironmentName string
}

var inferenceServiceCfg = InferenceServiceInputConfig{
	RegisteredModelName:    "mnist",
	ModelVersionName:       "v8",
	ModelArtifactName:      "mnist-8",
	ServingEnvironmentName: "model-server",
}
