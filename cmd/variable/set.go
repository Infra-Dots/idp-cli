package variable

import (
	"github.com/spf13/cobra"

	"github.com/infradots/idp-cli/internal/api"
	"github.com/infradots/idp-cli/internal/cli"
	"github.com/infradots/idp-cli/internal/output"
)

func newSetCmd() *cobra.Command {
	var wsName, category, description string
	var sensitive, hcl, env bool

	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Create or update a variable",
		Long: `Create a variable, or update it in place if the key already exists in the
same scope.

Variables are either Terraform variables (the default) or environment variables
(--env), which are exported into the run's environment.`,
		Example: `  idp variable set AWS_REGION us-east-1 --org my-org
  idp variable set AWS_ACCESS_KEY_ID AKIA... --org my-org --env --sensitive
  idp variable set TF_VAR_db_password secret --org my-org --workspace prod --sensitive
  idp variable set instance_tags '{"env":"prod"}' --org my-org --hcl`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			scope, err := cli.NewOrgScope()
			if err != nil {
				return err
			}

			if env && cmd.Flags().Changed("category") {
				return output.NewError("--env and --category are mutually exclusive")
			}
			if env {
				category = api.CategoryEnv
			}
			if category != "" && category != api.CategoryTerraform && category != api.CategoryEnv {
				return output.NewError("invalid --category %q: must be %s or %s",
					category, api.CategoryTerraform, api.CategoryEnv)
			}

			in := api.SetVariableInput{
				Key:         args[0],
				Value:       args[1],
				Sensitive:   sensitive,
				HCL:         hcl,
				Category:    category,
				Description: description,
			}

			// "set" implies upsert: the API has no upsert endpoint and only
			// rejects duplicates at workspace scope, so an existing key would
			// otherwise silently become a second row at org level.
			existing, err := findByKey(scope, wsName, args[0])
			if err != nil {
				return err
			}

			var v *api.Variable
			switch {
			case existing != nil:
				v, err = scope.Client.UpdateVariable(scope.Org, existing.ID, in)
			case wsName != "":
				v, err = scope.Client.CreateWorkspaceVariable(scope.Org, wsName, in)
			default:
				v, err = scope.Client.CreateOrgVariable(scope.Org, in)
			}
			if err != nil {
				return err
			}

			if scope.Printer.Quiet {
				scope.Printer.PrintID(v.ID)
				return nil
			}
			return scope.Printer.Print(v, listHeaders, [][]string{listRow(*v)})
		},
	}
	cmd.Flags().StringVarP(&wsName, "workspace", "w", "", "Workspace name (omit for org-level)")
	cmd.Flags().BoolVar(&sensitive, "sensitive", false, "Mark value as sensitive (write-only)")
	cmd.Flags().BoolVar(&hcl, "hcl", false, "Parse value as HCL")
	cmd.Flags().BoolVar(&env, "env", false, "Set as an environment variable (shorthand for --category env)")
	cmd.Flags().StringVar(&category, "category", "", "Variable category: terraform (default), env")
	cmd.Flags().StringVar(&description, "description", "", "Variable description")
	return cmd
}

// findByKey looks for an existing variable with this key in the given scope,
// returning nil when there is none.
func findByKey(scope *cli.OrgScope, wsName, key string) (*api.Variable, error) {
	var vars []api.Variable
	var err error
	if wsName != "" {
		vars, err = scope.Client.ListWorkspaceVariables(scope.Org, wsName)
	} else {
		vars, err = scope.Client.ListOrgVariables(scope.Org)
	}
	if err != nil {
		return nil, err
	}
	for i, v := range vars {
		if v.Key != key {
			continue
		}
		// The workspace listing includes inherited org-level variables; only an
		// entry actually owned by this scope is the one to update.
		if (wsName == "") == (v.Workspace == "") {
			return &vars[i], nil
		}
	}
	return nil, nil
}
