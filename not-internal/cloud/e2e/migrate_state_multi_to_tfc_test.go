package main

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	expect "github.com/Netflix/go-expect"
	tfe "github.com/hashicorp/go-tfe"
	"github.com/muratcelep/terraform/not-internal/e2e"
	tfversion "github.com/muratcelep/terraform/version"
)

func Test_migrate_multi_to_tfc_cloud_name_strategy(t *testing.T) {
	skipIfMissingEnvVar(t)
	skipWithoutRemoteTerraformVersion(t)

	ctx := context.Background()

	cases := map[string]struct {
		operations  []operationSets
		validations func(t *testing.T, orgName string)
	}{
		"migrating multiple workspaces to cloud using name strategy; current workspace is 'default'": {
			operations: []operationSets{
				{
					prep: func(t *testing.T, orgName, dir string) {
						tfBlock := terraformConfigLocalBackend()
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Successfully configured the backend "local"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
						{
							command:           []string{"workspace", "new", "prod"},
							expectedCmdOutput: `Created and switched to workspace "prod"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
						{
							command:           []string{"workspace", "select", "default"},
							expectedCmdOutput: `Switched to workspace "default".`,
						},
					},
				},
				{
					prep: func(t *testing.T, orgName, dir string) {
						wsName := "new-workspace"
						tfBlock := terraformConfigCloudBackendName(orgName, wsName)
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Do you want to copy only your current workspace?`,
							userInput:         []string{"yes"},
							postInputOutput:   []string{`Terraform Cloud has been successfully initialized!`},
						},
						{
							command:           []string{"workspace", "show"},
							expectedCmdOutput: `new-workspace`, // this comes from the `prep` function
						},
						{
							command:           []string{"output"},
							expectedCmdOutput: `val = "default"`, // this was the output of the current workspace selected before migration
						},
					},
				},
			},
			validations: func(t *testing.T, orgName string) {
				wsList, err := tfeClient.Workspaces.List(ctx, orgName, tfe.WorkspaceListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				if len(wsList.Items) != 1 {
					t.Fatalf("Expected the number of workspaces to be 1, but got %d", len(wsList.Items))
				}
				ws := wsList.Items[0]
				// this workspace name is what exists in the cloud backend configuration block
				if ws.Name != "new-workspace" {
					t.Fatalf("Expected workspace to be `new-workspace`, but is %s", ws.Name)
				}
			},
		},
		"migrating multiple workspaces to cloud using name strategy; current workspace is 'prod'": {
			operations: []operationSets{
				{
					prep: func(t *testing.T, orgName, dir string) {
						tfBlock := terraformConfigLocalBackend()
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Successfully configured the backend "local"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
						{
							command:           []string{"workspace", "new", "prod"},
							expectedCmdOutput: `Created and switched to workspace "prod"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
					},
				},
				{
					prep: func(t *testing.T, orgName, dir string) {
						wsName := "new-workspace"
						tfBlock := terraformConfigCloudBackendName(orgName, wsName)
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Do you want to copy only your current workspace?`,
							userInput:         []string{"yes"},
							postInputOutput:   []string{`Terraform Cloud has been successfully initialized!`},
						},
						{
							command:           []string{"workspace", "list"},
							expectedCmdOutput: `new-workspace`, // this comes from the `prep` function
						},
						{
							command:           []string{"output"},
							expectedCmdOutput: `val = "prod"`,
						},
					},
				},
			},
			validations: func(t *testing.T, orgName string) {
				wsList, err := tfeClient.Workspaces.List(ctx, orgName, tfe.WorkspaceListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				ws := wsList.Items[0]
				// this workspace name is what exists in the cloud backend configuration block
				if ws.Name != "new-workspace" {
					t.Fatalf("Expected workspace to be `new-workspace`, but is %s", ws.Name)
				}
			},
		},
		"migrating multiple workspaces to cloud using name strategy; 'default' workspace is empty": {
			operations: []operationSets{
				{
					prep: func(t *testing.T, orgName, dir string) {
						tfBlock := terraformConfigLocalBackend()
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Successfully configured the backend "local"!`,
						},
						{
							command:           []string{"workspace", "new", "workspace1"},
							expectedCmdOutput: `Created and switched to workspace "workspace1"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
						{
							command:           []string{"workspace", "new", "workspace2"},
							expectedCmdOutput: `Created and switched to workspace "workspace2"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
					},
				},
				{
					prep: func(t *testing.T, orgName, dir string) {
						wsName := "new-workspace"
						tfBlock := terraformConfigCloudBackendName(orgName, wsName)
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Do you want to copy only your current workspace?`,
							userInput:         []string{"yes"},
							postInputOutput:   []string{`Terraform Cloud has been successfully initialized!`},
						},
						{
							command:     []string{"workspace", "select", "default"},
							expectError: true,
						},
						{
							command:           []string{"output"},
							expectedCmdOutput: `val = "workspace2"`, // this was the output of the current workspace selected before migration
						},
					},
				},
			},
			validations: func(t *testing.T, orgName string) {
				wsList, err := tfeClient.Workspaces.List(ctx, orgName, tfe.WorkspaceListOptions{})
				if err != nil {
					t.Fatal(err)
				}
				if len(wsList.Items) != 1 {
					t.Fatalf("Expected the number of workspaces to be 1, but got %d", len(wsList.Items))
				}
				ws := wsList.Items[0]
				// this workspace name is what exists in the cloud backend configuration block
				if ws.Name != "new-workspace" {
					t.Fatalf("Expected workspace to be `new-workspace`, but is %s", ws.Name)
				}
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// t.Parallel()
			organization, cleanup := createOrganization(t)
			defer cleanup()
			exp, err := expect.NewConsole(defaultOpts()...)
			if err != nil {
				t.Fatal(err)
			}
			defer exp.Close()

			tmpDir, err := ioutil.TempDir("", "terraform-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			tf := e2e.NewBinary(terraformBin, tmpDir)
			defer tf.Close()
			tf.AddEnv(cliConfigFileEnv)

			for _, op := range tc.operations {
				op.prep(t, organization.Name, tf.WorkDir())
				for _, tfCmd := range op.commands {
					cmd := tf.Cmd(tfCmd.command...)
					cmd.Stdin = exp.Tty()
					cmd.Stdout = exp.Tty()
					cmd.Stderr = exp.Tty()

					err = cmd.Start()
					if err != nil {
						t.Fatal(err)
					}

					if tfCmd.expectedCmdOutput != "" {
						got, err := exp.ExpectString(tfCmd.expectedCmdOutput)
						if err != nil {
							t.Fatalf("error while waiting for output\nwant: %s\nerror: %s\noutput\n%s", tfCmd.expectedCmdOutput, err, got)
						}
					}

					lenInput := len(tfCmd.userInput)
					lenInputOutput := len(tfCmd.postInputOutput)
					if lenInput > 0 {
						for i := 0; i < lenInput; i++ {
							input := tfCmd.userInput[i]
							exp.SendLine(input)
							// use the index to find the corresponding
							// output that matches the input.
							if lenInputOutput-1 >= i {
								output := tfCmd.postInputOutput[i]
								_, err := exp.ExpectString(output)
								if err != nil {
									t.Fatal(err)
								}
							}
						}
					}

					err = cmd.Wait()
					if err != nil && !tfCmd.expectError {
						t.Fatal(err)
					}
				}
			}

			if tc.validations != nil {
				tc.validations(t, organization.Name)
			}
		})
	}
}

func Test_migrate_multi_to_tfc_cloud_tags_strategy(t *testing.T) {
	skipIfMissingEnvVar(t)
	skipWithoutRemoteTerraformVersion(t)

	ctx := context.Background()

	cases := map[string]struct {
		operations  []operationSets
		validations func(t *testing.T, orgName string)
	}{
		"migrating multiple workspaces to cloud using tags strategy; pattern is using prefix `app-*`": {
			operations: []operationSets{
				{
					prep: func(t *testing.T, orgName, dir string) {
						tfBlock := terraformConfigLocalBackend()
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Successfully configured the backend "local"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
						{
							command:           []string{"workspace", "new", "prod"},
							expectedCmdOutput: `Created and switched to workspace "prod"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
						{
							command:           []string{"workspace", "select", "default"},
							expectedCmdOutput: `Switched to workspace "default".`,
						},
						{
							command:           []string{"output"},
							expectedCmdOutput: `val = "default"`,
						},
						{
							command:           []string{"workspace", "select", "prod"},
							expectedCmdOutput: `Switched to workspace "prod".`,
						},
						{
							command:           []string{"output"},
							expectedCmdOutput: `val = "prod"`,
						},
					},
				},
				{
					prep: func(t *testing.T, orgName, dir string) {
						tag := "app"
						tfBlock := terraformConfigCloudBackendTags(orgName, tag)
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Terraform Cloud requires all workspaces to be given an explicit name.`,
							userInput:         []string{"dev", "1", "app-*"},
							postInputOutput: []string{
								`Would you like to rename your workspaces?`,
								"How would you like to rename your workspaces?",
								"Terraform Cloud has been successfully initialized!"},
						},
						{
							command:           []string{"workspace", "select", "app-dev"},
							expectedCmdOutput: `Switched to workspace "app-dev".`,
						},
						{
							command:           []string{"output"},
							expectedCmdOutput: `val = "default"`,
						},
						{
							command:           []string{"workspace", "select", "app-prod"},
							expectedCmdOutput: `Switched to workspace "app-prod".`,
						},
						{
							command:           []string{"output"},
							expectedCmdOutput: `val = "prod"`,
						},
					},
				},
			},
			validations: func(t *testing.T, orgName string) {
				wsList, err := tfeClient.Workspaces.List(ctx, orgName, tfe.WorkspaceListOptions{
					Tags: tfe.String("app"),
				})
				if err != nil {
					t.Fatal(err)
				}
				if len(wsList.Items) != 2 {
					t.Fatalf("Expected the number of workspaecs to be 2, but got %d", len(wsList.Items))
				}
				expectedWorkspaceNames := []string{"app-prod", "app-dev"}
				for _, ws := range wsList.Items {
					hasName := false
					for _, expectedNames := range expectedWorkspaceNames {
						if expectedNames == ws.Name {
							hasName = true
						}
					}
					if !hasName {
						t.Fatalf("Worksapce %s is not in the expected list of workspaces", ws.Name)
					}
				}
			},
		},
		"migrating multiple workspaces to cloud using tags strategy; existing workspaces": {
			operations: []operationSets{
				{
					prep: func(t *testing.T, orgName, dir string) {
						tfBlock := terraformConfigLocalBackend()
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Successfully configured the backend "local"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
						{
							command:           []string{"workspace", "new", "identity"},
							expectedCmdOutput: `Created and switched to workspace "identity"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
						{
							command:           []string{"workspace", "new", "billing"},
							expectedCmdOutput: `Created and switched to workspace "billing"!`,
						},
						{
							command:         []string{"apply", "-auto-approve"},
							postInputOutput: []string{`Apply complete!`},
						},
						{
							command:           []string{"workspace", "select", "default"},
							expectedCmdOutput: `Switched to workspace "default".`,
						},
					},
				},
				{
					prep: func(t *testing.T, orgName, dir string) {
						tag := "app"
						_ = createWorkspace(t, orgName, tfe.WorkspaceCreateOptions{
							Name:             tfe.String("identity"),
							TerraformVersion: tfe.String(tfversion.String()),
						})
						_ = createWorkspace(t, orgName, tfe.WorkspaceCreateOptions{
							Name:             tfe.String("billing"),
							TerraformVersion: tfe.String(tfversion.String()),
						})
						tfBlock := terraformConfigCloudBackendTags(orgName, tag)
						writeMainTF(t, tfBlock, dir)
					},
					commands: []tfCommand{
						{
							command:           []string{"init"},
							expectedCmdOutput: `Terraform Cloud requires all workspaces to be given an explicit name.`,
							userInput:         []string{"dev", "1", "app-*"},
							postInputOutput: []string{
								`Would you like to rename your workspaces?`,
								"How would you like to rename your workspaces?",
								"Terraform Cloud has been successfully initialized!"},
						},
						{
							command:           []string{"workspace", "select", "app-billing"},
							expectedCmdOutput: `Switched to workspace "app-billing".`,
						},
						{
							command:           []string{"workspace", "select", "app-identity"},
							expectedCmdOutput: `Switched to workspace "app-identity".`,
						},
						{
							command:           []string{"workspace", "select", "app-dev"},
							expectedCmdOutput: `Switched to workspace "app-dev".`,
						},
					},
				},
			},
			validations: func(t *testing.T, orgName string) {
				wsList, err := tfeClient.Workspaces.List(ctx, orgName, tfe.WorkspaceListOptions{
					Tags: tfe.String("app"),
				})
				if err != nil {
					t.Fatal(err)
				}
				if len(wsList.Items) != 3 {
					t.Fatalf("Expected the number of workspaecs to be 3, but got %d", len(wsList.Items))
				}
				expectedWorkspaceNames := []string{"app-billing", "app-dev", "app-identity"}
				for _, ws := range wsList.Items {
					hasName := false
					for _, expectedNames := range expectedWorkspaceNames {
						if expectedNames == ws.Name {
							hasName = true
						}
					}
					if !hasName {
						t.Fatalf("Worksapce %s is not in the expected list of workspaces", ws.Name)
					}
				}
			},
		},
	}

	for name, tc := range cases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			// t.Parallel()
			organization, cleanup := createOrganization(t)
			defer cleanup()
			exp, err := expect.NewConsole(defaultOpts()...)
			if err != nil {
				t.Fatal(err)
			}
			defer exp.Close()

			tmpDir, err := ioutil.TempDir("", "terraform-test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tmpDir)

			tf := e2e.NewBinary(terraformBin, tmpDir)
			defer tf.Close()
			tf.AddEnv(cliConfigFileEnv)

			for _, op := range tc.operations {
				op.prep(t, organization.Name, tf.WorkDir())
				for _, tfCmd := range op.commands {
					cmd := tf.Cmd(tfCmd.command...)
					cmd.Stdin = exp.Tty()
					cmd.Stdout = exp.Tty()
					cmd.Stderr = exp.Tty()

					err = cmd.Start()
					if err != nil {
						t.Fatal(err)
					}

					if tfCmd.expectedCmdOutput != "" {
						got, err := exp.ExpectString(tfCmd.expectedCmdOutput)
						if err != nil {
							t.Fatalf("error while waiting for output\nwant: %s\nerror: %s\noutput\n%s", tfCmd.expectedCmdOutput, err, got)
						}
					}

					lenInput := len(tfCmd.userInput)
					lenInputOutput := len(tfCmd.postInputOutput)
					if lenInput > 0 {
						for i := 0; i < lenInput; i++ {
							input := tfCmd.userInput[i]
							exp.SendLine(input)
							// use the index to find the corresponding
							// output that matches the input.
							if lenInputOutput-1 >= i {
								output := tfCmd.postInputOutput[i]
								if output == "" {
									continue
								}
								_, err := exp.ExpectString(output)
								if err != nil {
									t.Fatal(err)
								}
							}
						}
					}

					err = cmd.Wait()
					if err != nil {
						t.Fatal(err)
					}
				}
			}

			if tc.validations != nil {
				tc.validations(t, organization.Name)
			}
		})
	}
}
