package format

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/zclconf/go-cty/cty"

	"github.com/muratcelep/terraform/not-internal/addrs"
	"github.com/muratcelep/terraform/not-internal/configs/configschema"
	"github.com/muratcelep/terraform/not-internal/plans"
	"github.com/muratcelep/terraform/not-internal/states"
	"github.com/muratcelep/terraform/not-internal/terraform"
	"github.com/mitchellh/colorstring"
)

// StateOpts are the options for formatting a state.
type StateOpts struct {
	// State is the state to format. This is required.
	State *states.State

	// Schemas are used to decode attributes. This is required.
	Schemas *terraform.Schemas

	// Color is the colorizer. This is optional.
	Color *colorstring.Colorize
}

// State takes a state and returns a string
func State(opts *StateOpts) string {
	if opts.Color == nil {
		panic("colorize not given")
	}

	if opts.Schemas == nil {
		panic("schemas not given")
	}

	s := opts.State
	if len(s.Modules) == 0 {
		return "The state file is empty. No resources are represented."
	}

	buf := bytes.NewBufferString("[reset]")
	p := blockBodyDiffPrinter{
		buf:     buf,
		color:   opts.Color,
		action:  plans.NoOp,
		verbose: true,
	}

	// Format all the modules
	for _, m := range s.Modules {
		formatStateModule(p, m, opts.Schemas)
	}

	// Write the outputs for the root module
	m := s.RootModule()

	if m.OutputValues != nil {
		if len(m.OutputValues) > 0 {
			p.buf.WriteString("Outputs:\n\n")
		}

		// Sort the outputs
		ks := make([]string, 0, len(m.OutputValues))
		for k := range m.OutputValues {
			ks = append(ks, k)
		}
		sort.Strings(ks)

		// Output each output k/v pair
		for _, k := range ks {
			v := m.OutputValues[k]
			p.buf.WriteString(fmt.Sprintf("%s = ", k))
			if v.Sensitive {
				p.buf.WriteString("(sensitive value)")
			} else {
				p.writeValue(v.Value, plans.NoOp, 0)
			}
			p.buf.WriteString("\n")
		}
	}

	trimmedOutput := strings.TrimSpace(p.buf.String())
	trimmedOutput += "[reset]"

	return opts.Color.Color(trimmedOutput)

}

func formatStateModule(p blockBodyDiffPrinter, m *states.Module, schemas *terraform.Schemas) {
	// First get the names of all the resources so we can show them
	// in alphabetical order.
	names := make([]string, 0, len(m.Resources))
	for name := range m.Resources {
		names = append(names, name)
	}
	sort.Strings(names)

	// Go through each resource and begin building up the output.
	for _, key := range names {
		for k, v := range m.Resources[key].Instances {
			// keep these in order to keep the current object first, and
			// provide deterministic output for the deposed objects
			type obj struct {
				header   string
				instance *states.ResourceInstanceObjectSrc
			}
			instances := []obj{}

			addr := m.Resources[key].Addr
			resAddr := addr.Resource

			taintStr := ""
			if v.Current != nil && v.Current.Status == 'T' {
				taintStr = " (tainted)"
			}

			instances = append(instances,
				obj{fmt.Sprintf("# %s:%s\n", addr.Instance(k), taintStr), v.Current})

			for dk, v := range v.Deposed {
				instances = append(instances,
					obj{fmt.Sprintf("# %s: (deposed object %s)\n", addr.Instance(k), dk), v})
			}

			// Sort the instances for consistent output.
			// Starting the sort from the second index, so the current instance
			// is always first.
			sort.Slice(instances[1:], func(i, j int) bool {
				return instances[i+1].header < instances[j+1].header
			})

			for _, obj := range instances {
				header := obj.header
				instance := obj.instance
				p.buf.WriteString(header)
				if instance == nil {
					// this shouldn't happen, but there's nothing to do here so
					// don't panic below.
					continue
				}

				var schema *configschema.Block

				provider := m.Resources[key].ProviderConfig.Provider
				if _, exists := schemas.Providers[provider]; !exists {
					// This should never happen in normal use because we should've
					// loaded all of the schemas and checked things prior to this
					// point. We can't return errors here, but since this is UI code
					// we will try to do _something_ reasonable.
					p.buf.WriteString(fmt.Sprintf("# missing schema for provider %q\n\n", provider.String()))
					continue
				}

				switch resAddr.Mode {
				case addrs.ManagedResourceMode:
					schema, _ = schemas.ResourceTypeConfig(
						provider,
						resAddr.Mode,
						resAddr.Type,
					)
					if schema == nil {
						p.buf.WriteString(fmt.Sprintf(
							"# missing schema for provider %q resource type %s\n\n", provider, resAddr.Type))
						continue
					}

					p.buf.WriteString(fmt.Sprintf(
						"resource %q %q {",
						resAddr.Type,
						resAddr.Name,
					))
				case addrs.DataResourceMode:
					schema, _ = schemas.ResourceTypeConfig(
						provider,
						resAddr.Mode,
						resAddr.Type,
					)
					if schema == nil {
						p.buf.WriteString(fmt.Sprintf(
							"# missing schema for provider %q data source %s\n\n", provider, resAddr.Type))
						continue
					}

					p.buf.WriteString(fmt.Sprintf(
						"data %q %q {",
						resAddr.Type,
						resAddr.Name,
					))
				default:
					// should never happen, since the above is exhaustive
					p.buf.WriteString(resAddr.String())
				}

				val, err := instance.Decode(schema.ImpliedType())
				if err != nil {
					fmt.Println(err.Error())
					break
				}

				path := make(cty.Path, 0, 3)
				result := p.writeBlockBodyDiff(schema, val.Value, val.Value, 2, path)
				if result.bodyWritten {
					p.buf.WriteString("\n")
				}

				p.buf.WriteString("}\n\n")
			}
		}
	}
	p.buf.WriteString("\n")
}
