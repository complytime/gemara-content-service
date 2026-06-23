package service

import (
	"encoding/json"
	"testing"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/complytime/gemara-content-service/api"
	"github.com/complytime/gemara-content-service/mapper"
	"github.com/complytime/gemara-content-service/mapper/plugins/basic"
)

func TestNewService(t *testing.T) {
	mappers := make(mapper.Set)
	scope := make(mapper.Scope)

	service := NewService(mappers, scope)

	assert.NotNil(t, service)
	assert.Equal(t, mappers, service.set)
	assert.Equal(t, scope, service.scope)
}

func TestEnrich(t *testing.T) {
	t.Run("Enrichment with mapping", func(t *testing.T) {
		// Load the OpenAPI spec for validation
		swagger, err := api.GetSpec()
		require.NoError(t, err)

		// Set up a mapper with plans and catalog for successful mapping
		mapperPlugin := basic.NewBasicMapper()
		plans := []mapper.AssessmentPlan{
			{
				Control: mapper.PlanMapping{EntryId: "AC-1", ReferenceId: "test-catalog"},
				Assessments: []mapper.Assessment{
					{
						Requirement: mapper.PlanMapping{EntryId: "AC-1-REQ", ReferenceId: "test-catalog"},
						Procedures: []mapper.AssessmentProcedure{
							{
								Id:            "AC-1",
								Documentation: "Test procedure documentation",
							},
						},
					},
				},
			},
		}
		mapperPlugin.AddEvaluationPlan("test-catalog", plans...)

		catalog := gemara.ControlCatalog{
			Metadata: gemara.Metadata{Id: "test-catalog"},
			Groups: []gemara.Group{
				{
					Id:    "access-control",
					Title: "Access Control",
				},
			},
			Controls: []gemara.Control{
				{
					Id:    "AC-1",
					Group: "access-control",
					Guidelines: []gemara.MultiEntryMapping{
						{
							ReferenceId: "NIST-800-53",
							Entries: []gemara.ArtifactMapping{
								{ReferenceId: "AC-1"},
							},
						},
					},
				},
			},
		}

		evidence := api.Policy{
			PolicyEngineName: "test-policy-engine",
			PolicyRuleId:     "AC-1",
		}
		scope := mapper.Scope{
			"test-catalog": catalog,
		}

		compliance := mapperPlugin.Map(evidence, scope)
		response := api.EnrichmentResponse{
			Compliance: compliance,
		}

		assert.Equal(t, api.Success, response.Compliance.EnrichmentStatus)
		assert.Equal(t, "AC-1-REQ", response.Compliance.Control.Id)
		assert.Equal(t, "test-catalog", response.Compliance.Control.CatalogId)

		err = validateEnrichmentResponse(t, response, swagger)
		assert.NoError(t, err)
	})

	t.Run("Enrichment Unmapped", func(t *testing.T) {
		swagger, err := api.GetSpec()
		require.NoError(t, err)

		// Set up a mapper without plans or with empty scope to trigger unmapped response
		mapperPlugin := basic.NewBasicMapper()
		evidence := api.Policy{
			PolicyEngineName: "test-policy-engine",
			PolicyRuleId:     "AC-1",
		}
		scope := make(mapper.Scope)

		compliance := mapperPlugin.Map(evidence, scope)
		response := api.EnrichmentResponse{
			Compliance: compliance,
		}

		assert.Equal(t, api.Unmapped, response.Compliance.EnrichmentStatus)
		assert.Equal(t, "UNMAPPED", response.Compliance.Control.Id)
		assert.Equal(t, "UNMAPPED", response.Compliance.Control.CatalogId)
		assert.Equal(t, "UNCATEGORIZED", response.Compliance.Control.Category)

		err = validateEnrichmentResponse(t, response, swagger)
		assert.NoError(t, err, "Enrichment response with unmapped status should validate against OpenAPI schema")
	})
}

// validateEnrichmentResponse validates an EnrichmentResponse against the OpenAPI schema
func validateEnrichmentResponse(t *testing.T, response api.EnrichmentResponse, swagger *openapi3.T) error {
	t.Helper()
	pathItem := swagger.Paths.Find("/v1/enrich")
	require.NotNil(t, pathItem)
	operation := pathItem.Post
	require.NotNil(t, operation)

	responsesMap := operation.Responses.Map()
	responseRef, ok := responsesMap["200"]
	require.True(t, ok)
	require.NotNil(t, responseRef)

	responseValue := responseRef.Value
	require.NotNil(t, responseValue)

	content := responseValue.Content
	require.NotNil(t, content)

	mediaType := content.Get("application/json")
	require.NotNil(t, mediaType)

	schemaRef := mediaType.Schema
	require.NotNil(t, schemaRef)
	schema := schemaRef.Value
	require.NotNil(t, schema)

	responseJSON, err := json.Marshal(response)
	require.NoError(t, err)

	var responseBody interface{}
	require.NoError(t, json.Unmarshal(responseJSON, &responseBody))
	assert.NoError(t, schema.VisitJSON(responseBody))

	return nil
}
