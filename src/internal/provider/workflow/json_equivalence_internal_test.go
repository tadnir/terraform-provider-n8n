package workflow

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/kodflow/terraform-provider-n8n/sdk/n8nsdk"
	"github.com/kodflow/terraform-provider-n8n/src/internal/provider/workflow/models"
	"github.com/stretchr/testify/assert"
)

func Test_jsonEquivalent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configured types.String
		fromAPI    string
		adjust     func(configured, fromAPI any)
		want       bool
	}{
		{name: "null configured", configured: types.StringNull(), fromAPI: `{}`, want: false},
		{name: "unknown configured", configured: types.StringUnknown(), fromAPI: `{}`, want: false},
		{name: "invalid configured", configured: types.StringValue(`{`), fromAPI: `{}`, want: false},
		{name: "invalid from API", configured: types.StringValue(`{}`), fromAPI: `{`, want: false},
		{name: "same value in another order", configured: types.StringValue(`{"b":1,"a":[1,2]}`), fromAPI: `{"a":[1,2],"b":1.0}`, want: true},
		{name: "different value", configured: types.StringValue(`{"a":1}`), fromAPI: `{"a":2}`, want: false},
		{
			name:       "equal after adjust",
			configured: types.StringValue(`[{"name":"A"}]`),
			fromAPI:    `[{"id":"x","name":"A"}]`,
			adjust:     dropGeneratedNodeIDs,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, jsonEquivalent(tt.configured, []byte(tt.fromAPI), tt.adjust))
		})
	}
}

func Test_dropGeneratedNodeIDs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		configured any
		fromAPI    any
		want       any
	}{
		{
			name:       "configured not a list",
			configured: map[string]any{},
			fromAPI:    []any{map[string]any{"id": "x"}},
			want:       []any{map[string]any{"id": "x"}},
		},
		{
			name:       "API value not a list",
			configured: []any{map[string]any{}},
			fromAPI:    map[string]any{"id": "x"},
			want:       map[string]any{"id": "x"},
		},
		{
			name:       "different lengths",
			configured: []any{},
			fromAPI:    []any{map[string]any{"id": "x"}},
			want:       []any{map[string]any{"id": "x"}},
		},
		{
			name:       "nodes that are not objects",
			configured: []any{"a", map[string]any{}},
			fromAPI:    []any{map[string]any{"id": "x"}, "b"},
			want:       []any{map[string]any{"id": "x"}, "b"},
		},
		{
			name:       "generated id dropped, configured id kept",
			configured: []any{map[string]any{"name": "A"}, map[string]any{"id": "y", "name": "B"}},
			fromAPI:    []any{map[string]any{"id": "x", "name": "A"}, map[string]any{"id": "y", "name": "B"}},
			want:       []any{map[string]any{"name": "A"}, map[string]any{"id": "y", "name": "B"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			dropGeneratedNodeIDs(tt.configured, tt.fromAPI)
			assert.Equal(t, tt.want, tt.fromAPI)
		})
	}
}

func Test_serializeWorkflowJSON_KeepsEquivalentConfig(t *testing.T) {
	t.Parallel()

	nodeID := "generated-id"
	nodeName := "Start"
	nodeType := "n8n-nodes-base.manualTrigger"
	timezone := "Asia/Jerusalem"
	workflow := &n8nsdk.Workflow{
		Nodes:       []n8nsdk.Node{{Id: &nodeID, Name: &nodeName, Type: &nodeType}},
		Connections: map[string]any{"Start": map[string]any{"main": []any{}}},
		Settings:    n8nsdk.WorkflowSettings{Timezone: &timezone},
	}
	nodes := `[{"type":"n8n-nodes-base.manualTrigger","name":"Start"}]`
	connections := `{"Start":{"main":[]}}`
	settings := `{"timezone":"Asia/Jerusalem"}`
	plan := &models.Resource{
		NodesJSON:       types.StringValue(nodes),
		ConnectionsJSON: types.StringValue(connections),
		SettingsJSON:    types.StringValue(settings),
	}

	serializeWorkflowJSON(workflow, plan)

	assert.Equal(t, nodes, plan.NodesJSON.ValueString())
	assert.Equal(t, connections, plan.ConnectionsJSON.ValueString())
	assert.Equal(t, settings, plan.SettingsJSON.ValueString())
}

func Test_serializeWorkflowJSON_ReplacesDifferentConfig(t *testing.T) {
	t.Parallel()

	nodeName := "Changed"
	workflow := &n8nsdk.Workflow{
		Nodes:       []n8nsdk.Node{{Name: &nodeName}},
		Connections: map[string]any{},
		Settings:    n8nsdk.WorkflowSettings{},
	}
	plan := &models.Resource{
		NodesJSON:       types.StringValue(`[{"name":"Start"}]`),
		ConnectionsJSON: types.StringValue(`{"Start":{}}`),
		SettingsJSON:    types.StringValue(`{"timezone":"UTC"}`),
	}

	serializeWorkflowJSON(workflow, plan)

	assert.JSONEq(t, `[{"name":"Changed"}]`, plan.NodesJSON.ValueString())
	assert.JSONEq(t, `{}`, plan.ConnectionsJSON.ValueString())
	assert.JSONEq(t, `{}`, plan.SettingsJSON.ValueString())
}
