package mapper

import (
	gemara "github.com/gemaraproj/go-gemara"

	"github.com/complytime/gemara-content-service/api"
)

// Mapper defines a set of methods a plugin must implement for
// mapping api.Policy into a `gemara` AssessmentPlan.
type Mapper interface {
	PluginName() ID
	Map(policy api.Policy, scope Scope) api.Compliance
	AddEvaluationPlan(catalogId string, plans ...AssessmentPlan)
}

// ID represents the identity for a transformer.
type ID string

// NewID returns a new ID for a given id string.
func NewID(id string) ID {
	return ID(id)
}

// Set defines Transformers by ID
type Set map[ID]Mapper

// Scope defined in scope Layer2 Catalogs by the
// catalog ID
type Scope map[string]gemara.ControlCatalog
