package cmd

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/lampajr/model-registry-go-example/internal/utils"
	"github.com/opendatahub-io/model-registry/pkg/api"
	"github.com/opendatahub-io/model-registry/pkg/core"
	"github.com/opendatahub-io/model-registry/pkg/openapi"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// registerCmd represents the register command
	registerCmd = &cobra.Command{
		Use:   "register",
		Short: "Register a new model version",
		Long:  `This command registers a new model version, if the model does not exist yet it creates it.`,
		RunE:  registerModel,
	}
)

func registerModel(cmd *cobra.Command, args []string) error {
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
	log.Printf("connected to mlmd server")
	defer conn.Close()

	service, err := core.NewModelRegistryService(conn)
	if err != nil {
		return fmt.Errorf("error creating core service: %v", err)
	}
	log.Printf("model registry service created")

	return doRegisterModel(service)
}

func doRegisterModel(service api.ModelRegistryApi) error {

	// preconditions
	if modelRegistrationCfg.RegisteredModelName == "" || modelRegistrationCfg.ModelVersionName == "" {
		return fmt.Errorf("missing required parameter: [model, version] are mandatory and cannot be empty")
	}

	// get or create a registered model
	registeredModel, err := service.GetRegisteredModelByParams(utils.Of(modelRegistrationCfg.RegisteredModelName), nil)
	if err != nil {
		log.Printf("unable to find model %s: %v", modelRegistrationCfg.RegisteredModelName, err)
		log.Printf("registering new model %s..", modelRegistrationCfg.RegisteredModelName)

		// register a new model
		registeredModel, err = service.UpsertRegisteredModel(&openapi.RegisteredModel{
			Name:        utils.Of(modelRegistrationCfg.RegisteredModelName),
			ExternalID:  utils.Of(modelRegistrationCfg.RegisteredModelName),
			Description: utils.Of(modelRegistrationCfg.RegisteredModelDescription),
		})
		if err != nil {
			return fmt.Errorf("error registering model: %v", err)
		}
	}

	log.Printf("registered model: %s", utils.Marshal(registeredModel))

	modelVersion, err := service.GetModelVersionByParams(utils.Of(modelRegistrationCfg.ModelVersionName), registeredModel.Id, nil)
	if err != nil {
		log.Printf("unable to find model version %s: %v", modelRegistrationCfg.ModelVersionName, err)
		log.Printf("registering new model version %s..", modelRegistrationCfg.ModelVersionName)

		modelVersion, err = service.UpsertModelVersion(&openapi.ModelVersion{
			Name:        utils.Of(modelRegistrationCfg.ModelVersionName),
			ExternalID:  utils.Of(modelRegistrationCfg.ModelVersionName),
			Description: utils.Of(modelRegistrationCfg.ModelVersionDescription),
			CustomProperties: &map[string]openapi.MetadataValue{
				"score": {
					MetadataDoubleValue: &openapi.MetadataDoubleValue{
						DoubleValue: utils.Of(rand.Float64()),
					},
				},
			},
		}, registeredModel.Id)
		if err != nil {
			return fmt.Errorf("error registering model version: %v", err)
		}
	}

	log.Printf("model version: %s", utils.Marshal(modelVersion))

	modelArtifactName := utils.CreateArtifactName(modelRegistrationCfg.ModelVersionName)
	// TODO: remove once https://github.com/opendatahub-io/model-registry/pull/165 get merged
	// modelArtifact, err := service.GetModelArtifactByParams(utils.Of(modelArtifactName), modelVersion.Id, nil)
	// if err != nil {
	// 	log.Printf("unable to find model artifact %s: %v", modelArtifactName, err)
	// 	log.Printf("creating new model artifact %s..", modelArtifactName)

	modelArtifact, err := service.UpsertModelArtifact(&openapi.ModelArtifact{
		Name:        utils.Of(modelArtifactName),
		Description: utils.Of(fmt.Sprintf("model artifact for model %s", modelRegistrationCfg.ModelVersionName)),
		// serving infos
		State:              utils.Of(openapi.ARTIFACTSTATE_UNKNOWN),
		ModelFormatName:    utils.Of(modelRegistrationCfg.ModelFormatName),
		ModelFormatVersion: utils.Of(modelRegistrationCfg.ModelFormatVersion),
		StorageKey:         utils.Of(modelRegistrationCfg.StorageKey),
		StoragePath:        utils.Of(modelRegistrationCfg.StoragePath),
	}, modelVersion.Id)
	if err != nil {
		return fmt.Errorf("error creating model artifact: %v", err)
	}
	// }

	log.Printf("model artifact: %s", utils.Marshal(modelArtifact))

	return nil
}

func init() {
	rootCmd.AddCommand(registerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	registerCmd.Flags().StringVar(&modelRegistrationCfg.RegisteredModelName, "model", modelRegistrationCfg.RegisteredModelName, "Registered model name")
	registerCmd.Flags().StringVar(&modelRegistrationCfg.ModelVersionName, "version", modelRegistrationCfg.ModelVersionName, "Specific model version name")
	registerCmd.Flags().StringVar(&modelRegistrationCfg.RegisteredModelDescription, "model-description", modelRegistrationCfg.RegisteredModelDescription, "Registered model description")
	registerCmd.Flags().StringVar(&modelRegistrationCfg.ModelVersionDescription, "version-description", modelRegistrationCfg.ModelVersionDescription, "Model version description")
	registerCmd.Flags().StringVar(&modelRegistrationCfg.ModelFormatName, "format-name", modelRegistrationCfg.ModelFormatName, "Model artifact format name, e.g., onnx")
	registerCmd.Flags().StringVar(&modelRegistrationCfg.ModelFormatVersion, "format-version", modelRegistrationCfg.ModelFormatVersion, "Model artifact format version, e.g., 1")
	registerCmd.Flags().StringVar(&modelRegistrationCfg.StorageKey, "key", modelRegistrationCfg.StorageKey, "Model artifact data connection key")
	registerCmd.Flags().StringVar(&modelRegistrationCfg.StoragePath, "path", modelRegistrationCfg.StoragePath, "Model artifact data connection storage path")
}

type ModelRegistrationConfig struct {
	RegisteredModelName        string
	RegisteredModelDescription string
	ModelVersionName           string
	ModelVersionDescription    string
	ModelFormatName            string
	ModelFormatVersion         string
	StorageKey                 string
	StoragePath                string
}

var modelRegistrationCfg = ModelRegistrationConfig{
	RegisteredModelName:        "",
	RegisteredModelDescription: "",
	ModelVersionName:           "",
	ModelVersionDescription:    "",
	ModelFormatName:            "onnx",
	ModelFormatVersion:         "1",
	StorageKey:                 "aws-connection-models",
	StoragePath:                "",
}
