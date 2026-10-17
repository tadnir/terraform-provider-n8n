package workflow

import "time"

// Values repeated across this package's data source, plural data source and
// resource schemas. They are the wire-level attribute names and the strings
// Terraform shows to users, so the three files must never drift apart.
const (
	attrName   = "name"
	attrActive = "active"
)

const (
	descID     = "Workflow identifier"
	descName   = "Workflow name"
	descActive = "Whether the workflow is active"
)

const (
	// deleteConflictRetries is how many times a delete answered with 409 Conflict
	// is retried while n8n 2.x finishes unpublishing the workflow.
	deleteConflictRetries int = 30
	// deleteConflictDelayDefault is the wait between those retries.
	deleteConflictDelayDefault time.Duration = time.Second
)
