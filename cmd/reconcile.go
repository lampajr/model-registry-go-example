package cmd

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/opendatahub-io/model-registry/pkg/api"
	"github.com/opendatahub-io/model-registry/pkg/core"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	// reconcileCmd represents the reconcile command
	reconcileCmd = &cobra.Command{
		Use:   "reconcile",
		Short: "Reconcile the inference services.",
		Long: `This command retrieves all inference services and create the corresponding InferenceService CRs.
		
	InferenceService CRs are created starting from go template, filled in by information found in the model registry.`,
		RunE: reconcile,
	}
)

func reconcile(cmd *cobra.Command, args []string) error {
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

	return doReconcile(service)
}

func doReconcile(service api.ModelRegistryApi) error {
	return fmt.Errorf("method not yet implemented")
}

func init() {
	rootCmd.AddCommand(reconcileCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
}
