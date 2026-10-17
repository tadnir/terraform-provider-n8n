// Copyright (c) 2024 Florent (Kodflow). All rights reserved.
// Licensed under the Sustainable Use License 1.0
// See LICENSE in the project root for license information.

package workflow

import (
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// jsonEquivalent reports whether a configured JSON string holds the same value
// as the JSON n8n returned, ignoring key order and formatting.
// n8n returns nodes, connections and settings with its own key order, so
// comparing the strings would report a change after every apply.
//
// Params:
//   - configured: The JSON string from the plan or prior state
//   - fromAPI: The JSON serialized from the API response
//   - adjust: Optional hook that edits the decoded values before comparing
//
// Returns:
//   - bool: True if both decode to the same value
func jsonEquivalent(configured types.String, fromAPI []byte, adjust func(configured, fromAPI any)) bool {
	// Nothing configured to keep.
	if configured.IsNull() || configured.IsUnknown() {
		// Return not equivalent.
		return false
	}

	var configuredValue, apiValue any
	// Check for error.
	if err := json.Unmarshal([]byte(configured.ValueString()), &configuredValue); err != nil {
		// Return not equivalent.
		return false
	}
	// Check for error.
	if err := json.Unmarshal(fromAPI, &apiValue); err != nil {
		// Return not equivalent.
		return false
	}

	// Check for non-nil value.
	if adjust != nil {
		adjust(configuredValue, apiValue)
	}

	// Return comparison result.
	return reflect.DeepEqual(configuredValue, apiValue)
}

// dropGeneratedNodeIDs removes from the API's nodes the ids n8n generated for
// nodes configured without one, so that such nodes still compare equal.
//
// Params:
//   - configured: The decoded configured nodes
//   - fromAPI: The decoded nodes from the API, edited in place
func dropGeneratedNodeIDs(configured, fromAPI any) {
	configuredNodes, okConfigured := configured.([]any)
	apiNodes, okAPI := fromAPI.([]any)
	// Check condition.
	if !okConfigured || !okAPI || len(configuredNodes) != len(apiNodes) {
		// Return without changes.
		return
	}

	// Iterate over items.
	for i, node := range configuredNodes {
		configuredNode, okConfiguredNode := node.(map[string]any)
		apiNode, okAPINode := apiNodes[i].(map[string]any)
		// Check condition.
		if !okConfiguredNode || !okAPINode {
			continue
		}
		// Check condition.
		if _, hasID := configuredNode["id"]; !hasID {
			delete(apiNode, "id")
		}
	}
}
