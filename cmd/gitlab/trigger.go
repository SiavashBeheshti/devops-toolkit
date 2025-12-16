package gitlab

import (
	"fmt"

	"github.com/beheshti/devops-toolkit/pkg/output"
	"github.com/spf13/cobra"
)

func newTriggerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trigger",
		Short: "Trigger a new pipeline",
		Long: `Trigger a new GitLab CI/CD pipeline.

Examples:
  devops-toolkit gitlab trigger -p myproject -r main
  devops-toolkit gitlab trigger -p myproject -r main -v KEY=value`,
		RunE: runTrigger,
	}

	cmd.Flags().StringP("ref", "r", "", "Branch or tag to run pipeline on (required)")
	cmd.Flags().StringArrayP("variable", "v", nil, "Pipeline variables (KEY=value)")
	cmd.Flags().Bool("wait", false, "Wait for pipeline to complete")

	cmd.MarkFlagRequired("ref")

	return cmd
}

func runTrigger(cmd *cobra.Command, args []string) error {
	ref, _ := cmd.Flags().GetString("ref")
	variables, _ := cmd.Flags().GetStringArray("variable")
	wait, _ := cmd.Flags().GetBool("wait")

	output.StartSpinner(fmt.Sprintf("Triggering pipeline on %s...", ref))

	client, projectID, err := getClient(cmd)
	if err != nil {
		output.SpinnerError("Failed to connect to GitLab")
		return err
	}

	// Parse variables
	vars := make(map[string]string)
	for _, v := range variables {
		parts := splitVar(v)
		if len(parts) == 2 {
			vars[parts[0]] = parts[1]
		}
	}

	pipeline, err := client.TriggerPipeline(projectID, ref, vars)
	if err != nil {
		output.SpinnerError("Failed to trigger pipeline")
		return fmt.Errorf("failed to trigger pipeline: %w", err)
	}

	output.SpinnerSuccess("Pipeline triggered successfully")
	output.Newline()

	// Show pipeline info
	output.Header("Pipeline Created")
	output.Printf("  %s\n", output.KeyValue("Pipeline ID", fmt.Sprintf("#%d", pipeline.ID)))
	output.Printf("  %s\n", output.KeyValue("Ref", pipeline.Ref))
	output.Printf("  %s\n", output.KeyValue("Status", pipeline.Status))
	output.Printf("  %s\n", output.KeyValue("Web URL", pipeline.WebURL))

	if len(vars) > 0 {
		output.Newline()
		output.Print(output.Section("Variables"))
		for k, v := range vars {
			output.Printf("  %s = %s\n", output.InfoStyle.Render(k), v)
		}
	}

	output.Newline()

	if wait {
		output.StartSpinner("Waiting for pipeline to complete...")

		finalPipeline, err := client.WaitForPipeline(projectID, pipeline.ID)
		if err != nil {
			output.SpinnerError("Error waiting for pipeline")
			return err
		}

		switch finalPipeline.Status {
		case "success", "passed":
			output.SpinnerSuccess(fmt.Sprintf("Pipeline completed successfully in %s", finalPipeline.Duration))
		case "failed":
			output.SpinnerError(fmt.Sprintf("Pipeline failed after %s", finalPipeline.Duration))
		default:
			output.StopSpinner()
			output.Warning(fmt.Sprintf("Pipeline ended with status: %s", finalPipeline.Status))
		}
	}

	return nil
}

func splitVar(v string) []string {
	for i, c := range v {
		if c == '=' {
			return []string{v[:i], v[i+1:]}
		}
	}
	return []string{v}
}

