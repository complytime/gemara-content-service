package basic

import (
	"testing"

	gemara "github.com/gemaraproj/go-gemara"
	"github.com/stretchr/testify/assert"

	"github.com/complytime/gemara-content-service/api"
	"github.com/complytime/gemara-content-service/mapper"
)

func TestNewBasicMapper(t *testing.T) {
	basicMapper := NewBasicMapper()

	assert.NotNil(t, basicMapper)
	assert.Equal(t, ID, basicMapper.PluginName())
	assert.NotNil(t, basicMapper.plans)
	assert.Empty(t, basicMapper.plans)
}

func TestBasicMapper_MapWithPlans(t *testing.T) {
	tests := []struct {
		name           string
		policyRuleId   string
		expectedStatus api.ComplianceEnrichmentStatus
	}{
		{
			name:           "mapped policy rule returns success",
			policyRuleId:   "AC-1",
			expectedStatus: api.Success,
		},
		{
			name:           "unmapped policy rule returns unmapped",
			policyRuleId:   "UNMAPPED",
			expectedStatus: api.Unmapped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			basicMapper := NewBasicMapper()

			// Add a test plan
			plans := []mapper.AssessmentPlan{
				{
					Control: mapper.PlanMapping{EntryId: "AC-1", ReferenceId: "test-catalog"},
					Assessments: []mapper.Assessment{
						{
							Requirement: mapper.PlanMapping{EntryId: "AC-1-REQ", ReferenceId: "test-catalog"},
							Procedures: []mapper.AssessmentProcedure{
								{
									Id:            "AC-1",
									Documentation: "Test procedure",
								},
							},
						},
					},
				},
			}
			basicMapper.AddEvaluationPlan("test-catalog", plans...)

			// Create a test catalog
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

			// Create test policy
			policy := api.Policy{
				PolicyEngineName: "test-policy-engine",
				PolicyRuleId:     tt.policyRuleId,
			}
			scope := mapper.Scope{
				"test-catalog": catalog,
			}

			// Test Map method
			compliance := basicMapper.Map(policy, scope)
			assert.NotNil(t, compliance)
			assert.Equal(t, tt.expectedStatus, compliance.EnrichmentStatus)

			if tt.expectedStatus == api.Success {
				assert.Equal(t, "AC-1-REQ", compliance.Control.Id)
				assert.Equal(t, "Access Control", compliance.Control.Category)
				assert.Equal(t, "test-catalog", compliance.Control.CatalogId)
				assert.NotNil(t, compliance.Control.RemediationDescription)
				assert.Equal(t, "Test procedure", *compliance.Control.RemediationDescription)
				assert.Contains(t, compliance.Frameworks.Frameworks, "NIST-800-53")
			} else {
				assert.Equal(t, tt.policyRuleId, compliance.Control.Id)
				assert.Equal(t, "UNCATEGORIZED", compliance.Control.Category)
				assert.Equal(t, "UNMAPPED", compliance.Control.CatalogId)
			}
		})
	}
}

func TestBasicMapper_MapUnmapped(t *testing.T) {
	basicMapper := NewBasicMapper()
	policy := api.Policy{
		PolicyEngineName: "test-policy-engine",
		PolicyRuleId:     "AC-1",
	}
	scope := make(mapper.Scope)

	// Test Map method for unmapped policy
	compliance := basicMapper.Map(policy, scope)
	assert.NotNil(t, compliance)
	assert.Equal(t, api.Unmapped, compliance.EnrichmentStatus)
	assert.Equal(t, "UNMAPPED", compliance.Control.Id)
	assert.Equal(t, "UNCATEGORIZED", compliance.Control.Category)
	assert.Equal(t, "UNMAPPED", compliance.Control.CatalogId)
	assert.Empty(t, compliance.Frameworks.Requirements)
	assert.Empty(t, compliance.Frameworks.Frameworks)
}

func TestBasicMapper_AddEvaluationPlan(t *testing.T) {
	t.Run("adds evaluation plan", func(t *testing.T) {
		basicMapper := NewBasicMapper()
		plans := []mapper.AssessmentPlan{
			{Control: mapper.PlanMapping{ReferenceId: "AC-1"}},
		}

		basicMapper.AddEvaluationPlan("test-catalog", plans...)

		assert.Len(t, basicMapper.plans, 1)
		assert.Contains(t, basicMapper.plans, "test-catalog")
		assert.Len(t, basicMapper.plans["test-catalog"], 1)
		assert.Equal(t, "AC-1", basicMapper.plans["test-catalog"][0].Control.ReferenceId)
	})

	t.Run("appends to existing evaluation plans", func(t *testing.T) {
		basicMapper := NewBasicMapper()
		initialPlans := []mapper.AssessmentPlan{
			{Control: mapper.PlanMapping{ReferenceId: "AC-1"}},
		}
		additionalPlans := []mapper.AssessmentPlan{
			{Control: mapper.PlanMapping{ReferenceId: "AC-2"}},
		}

		basicMapper.AddEvaluationPlan("test-catalog", initialPlans...)
		basicMapper.AddEvaluationPlan("test-catalog", additionalPlans...)

		assert.Len(t, basicMapper.plans, 1)
		assert.Len(t, basicMapper.plans["test-catalog"], 2)
		assert.Equal(t, "AC-1", basicMapper.plans["test-catalog"][0].Control.ReferenceId)
		assert.Equal(t, "AC-2", basicMapper.plans["test-catalog"][1].Control.ReferenceId)
	})
}
