package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/adapters"
	"github.com/jdbencardinop/tesserasessions/internal/statusprovider"
	"github.com/spf13/cobra"
)

func (a *appContext) statusCmd() *cobra.Command {
	return newStatusCmd(statusprovider.NewService(adapters.LiveScanners()))
}

func newStatusCmd(service *statusprovider.Service) *cobra.Command {
	var jsonOutput bool
	var timeout time.Duration
	var freshFor time.Duration

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Query live runtime status from a batch JSON request",
		Long: `Read a versioned batch request from stdin and return live runtime status.

The command is side-effect-free: it does not scan or write the inventory,
attach to runtimes, focus panes, or send agent input.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if !jsonOutput {
				return fmt.Errorf("status currently requires --json")
			}
			request, err := decodeStatusRequest(cmd.InOrStdin())
			if err != nil {
				return fmt.Errorf("invalid status request: %w", err)
			}

			service.Timeout = timeout
			service.FreshFor = freshFor
			response, err := service.Query(cmd.Context(), request)
			if err != nil {
				return fmt.Errorf("invalid status request: %w", err)
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetEscapeHTML(false)
			return encoder.Encode(response)
		},
	}
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "read and write the versioned JSON contract")
	cmd.Flags().DurationVar(&timeout, "timeout", 2*time.Second, "maximum duration for each live provider probe")
	cmd.Flags().DurationVar(&freshFor, "fresh-for", 10*time.Second, "freshness window reported for observations")
	return cmd
}

func decodeStatusRequest(reader io.Reader) (statusprovider.Request, error) {
	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()

	var request statusprovider.Request
	if err := decoder.Decode(&request); err != nil {
		return statusprovider.Request{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return statusprovider.Request{}, fmt.Errorf("multiple JSON values are not allowed")
		}
		return statusprovider.Request{}, err
	}
	return request, nil
}
