package command

import (
	"fmt"
	"os"
	"strings"

	"github.com/muratcelep/terraform/not-internal/addrs"
	"github.com/muratcelep/terraform/not-internal/backend"
	"github.com/muratcelep/terraform/not-internal/command/format"
	"github.com/muratcelep/terraform/not-internal/states"
	"github.com/mitchellh/cli"
)

// StateShowCommand is a Command implementation that shows a single resource.
type StateShowCommand struct {
	Meta
	StateMeta
}

func (c *StateShowCommand) Run(args []string) int {
	args = c.Meta.process(args)
	cmdFlags := c.Meta.defaultFlagSet("state show")
	cmdFlags.StringVar(&c.Meta.statePath, "state", "", "path")
	if err := cmdFlags.Parse(args); err != nil {
		c.Ui.Error(fmt.Sprintf("Error parsing command-line flags: %s\n", err.Error()))
		return 1
	}
	args = cmdFlags.Args()
	if len(args) != 1 {
		c.Ui.Error("Exactly one argument expected.\n")
		return cli.RunResultHelp
	}

	// Check for user-supplied plugin path
	var err error
	if c.pluginPath, err = c.loadPluginPath(); err != nil {
		c.Ui.Error(fmt.Sprintf("Error loading plugin path: %s", err))
		return 1
	}

	// Load the backend
	b, backendDiags := c.Backend(nil)
	if backendDiags.HasErrors() {
		c.showDiagnostics(backendDiags)
		return 1
	}

	// We require a local backend
	local, ok := b.(backend.Local)
	if !ok {
		c.Ui.Error(ErrUnsupportedLocalOp)
		return 1
	}

	// This is a read-only command
	c.ignoreRemoteVersionConflict(b)

	// Check if the address can be parsed
	addr, addrDiags := addrs.ParseAbsResourceInstanceStr(args[0])
	if addrDiags.HasErrors() {
		c.Ui.Error(fmt.Sprintf(errParsingAddress, args[0]))
		return 1
	}

	// We expect the config dir to always be the cwd
	cwd, err := os.Getwd()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error getting cwd: %s", err))
		return 1
	}

	// Build the operation (required to get the schemas)
	opReq := c.Operation(b)
	opReq.AllowUnsetVariables = true
	opReq.ConfigDir = cwd

	opReq.ConfigLoader, err = c.initConfigLoader()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error initializing config loader: %s", err))
		return 1
	}

	// Get the context (required to get the schemas)
	lr, _, ctxDiags := local.LocalRun(opReq)
	if ctxDiags.HasErrors() {
		c.showDiagnostics(ctxDiags)
		return 1
	}

	// Get the schemas from the context
	schemas, diags := lr.Core.Schemas(lr.Config, lr.InputState)
	if diags.HasErrors() {
		c.showDiagnostics(diags)
		return 1
	}

	// Get the state
	env, err := c.Workspace()
	if err != nil {
		c.Ui.Error(fmt.Sprintf("Error selecting workspace: %s", err))
		return 1
	}
	stateMgr, err := b.StateMgr(env)
	if err != nil {
		c.Ui.Error(fmt.Sprintf(errStateLoadingState, err))
		return 1
	}
	if err := stateMgr.RefreshState(); err != nil {
		c.Ui.Error(fmt.Sprintf("Failed to refresh state: %s", err))
		return 1
	}

	state := stateMgr.State()
	if state == nil {
		c.Ui.Error(errStateNotFound)
		return 1
	}

	is := state.ResourceInstance(addr)
	if !is.HasCurrent() {
		c.Ui.Error(errNoInstanceFound)
		return 1
	}

	// check if the resource has a configured provider, otherwise this will use the default provider
	rs := state.Resource(addr.ContainingResource())
	absPc := addrs.AbsProviderConfig{
		Provider: rs.ProviderConfig.Provider,
		Alias:    rs.ProviderConfig.Alias,
		Module:   addrs.RootModule,
	}
	singleInstance := states.NewState()
	singleInstance.EnsureModule(addr.Module).SetResourceInstanceCurrent(
		addr.Resource,
		is.Current,
		absPc,
	)

	output := format.State(&format.StateOpts{
		State:   singleInstance,
		Color:   c.Colorize(),
		Schemas: schemas,
	})
	c.Ui.Output(output[strings.Index(output, "#"):])

	return 0
}

func (c *StateShowCommand) Help() string {
	helpText := `
Usage: terraform [global options] state show [options] ADDRESS

  Shows the attributes of a resource in the Terraform state.

  This command shows the attributes of a single resource in the Terraform
  state. The address argument must be used to specify a single resource.
  You can view the list of available resources with "terraform state list".

Options:

  -state=statefile    Path to a Terraform state file to use to look
                      up Terraform-managed resources. By default it will
                      use the state "terraform.tfstate" if it exists.

`
	return strings.TrimSpace(helpText)
}

func (c *StateShowCommand) Synopsis() string {
	return "Show a resource in the state"
}

const errNoInstanceFound = `No instance found for the given address!

This command requires that the address references one specific instance.
To view the available instances, use "terraform state list". Please modify 
the address to reference a specific instance.`

const errParsingAddress = `Error parsing instance address: %s

This command requires that the address references one specific instance.
To view the available instances, use "terraform state list". Please modify 
the address to reference a specific instance.`
