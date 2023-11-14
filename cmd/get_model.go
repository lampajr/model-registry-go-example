package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lampajr/model-registry-go-example/internal/utils"
	"github.com/opendatahub-io/model-registry/pkg/api"
	"github.com/opendatahub-io/model-registry/pkg/core"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// getCmd represents the get command
	getCmd = &cobra.Command{
		Use:   "get",
		Short: "Retrieve all information for a specific model",
		Long:  `This command gets all information for a specific model.`,
		RunE:  getModel,
	}
)

func getModel(cmd *cobra.Command, args []string) error {
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

	return doGetModel(service)
}

func doGetModel(service api.ModelRegistryApi) error {

	// preconditions
	if getModelCfg.RegisteredModelName == "" {
		return fmt.Errorf("missing required `model` parameter")
	}

	// get or create a registered model
	registeredModel, err := service.GetRegisteredModelByParams(utils.Of(getModelCfg.RegisteredModelName), nil)
	if err != nil {
		log.Printf("unable to find model %s: %v", getModelCfg.RegisteredModelName, err)
		return err
	}

	log.Printf("registered model: %s", utils.Marshal(registeredModel))

	allVersions, err := service.GetModelVersions(api.ListOptions{}, registeredModel.Id)
	if err != nil {
		return fmt.Errorf("error retrieving model versions for model %s: %v", *registeredModel.Id, err)
	} else {
		for _, v := range allVersions.Items {
			log.Printf("model version: %s", utils.Marshal(v))

			allArtifacts, err := service.GetModelArtifacts(api.ListOptions{}, v.Id)
			if err != nil {
				return fmt.Errorf("error retrieving model artifacts for version %s: %v", *v.Id, err)
			} else {
				for _, a := range allArtifacts.Items {
					log.Printf("model artifact: %s", utils.Marshal(a))
				}
			}
		}

	}

	return nil
}

func init() {
	rootCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	getCmd.Flags().StringVar(&getModelCfg.RegisteredModelName, "model", getModelCfg.RegisteredModelName, "Registered model name")
}

type GetModelConfig struct {
	RegisteredModelName string
}

var getModelCfg = GetModelConfig{
	RegisteredModelName: "",
}
