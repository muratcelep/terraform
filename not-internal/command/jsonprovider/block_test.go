package jsonprovider

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/zclconf/go-cty/cty"

	"github.com/muratcelep/terraform/not-internal/configs/configschema"
)

func TestMarshalBlock(t *testing.T) {
	tests := []struct {
		Input *configschema.Block
		Want  *block
	}{
		{
			nil,
			&block{},
		},
		{
			Input: &configschema.Block{
				Attributes: map[string]*configschema.Attribute{
					"id":  {Type: cty.String, Optional: true, Computed: true},
					"ami": {Type: cty.String, Optional: true},
				},
				BlockTypes: map[string]*configschema.NestedBlock{
					"network_interface": {
						Nesting: configschema.NestingList,
						Block: configschema.Block{
							Attributes: map[string]*configschema.Attribute{
								"device_index": {Type: cty.String, Optional: true},
								"description":  {Type: cty.String, Optional: true},
							},
						},
					},
				},
			},
			Want: &block{
				Attributes: map[string]*attribute{
					"ami": {AttributeType: json.RawMessage(`"string"`), Optional: true, DescriptionKind: "plain"},
					"id":  {AttributeType: json.RawMessage(`"string"`), Optional: true, Computed: true, DescriptionKind: "plain"},
				},
				BlockTypes: map[string]*blockType{
					"network_interface": {
						NestingMode: "list",
						Block: &block{
							Attributes: map[string]*attribute{
								"description":  {AttributeType: json.RawMessage(`"string"`), Optional: true, DescriptionKind: "plain"},
								"device_index": {AttributeType: json.RawMessage(`"string"`), Optional: true, DescriptionKind: "plain"},
							},
							DescriptionKind: "plain",
						},
					},
				},
				DescriptionKind: "plain",
			},
		},
	}

	for _, test := range tests {
		got := marshalBlock(test.Input)
		if !cmp.Equal(got, test.Want) {
			t.Fatalf("wrong result:\n %v\n", cmp.Diff(got, test.Want))
		}
	}
}
