package mapper

// Local evaluation plan types. These were removed from the gemara Go SDK
// in v1.0.0 when the SDK moved to github.com/gemaraproj/go-gemara.
// They are defined here to preserve the existing evaluation plan YAML
// format used by this service.

// EvaluationPlan represents a set of assessment plans loaded from YAML.
type EvaluationPlan struct {
	Metadata EvaluationMetadata `yaml:"metadata" json:"metadata"`
	Plans    []AssessmentPlan   `yaml:"plans" json:"plans"`
}

// EvaluationMetadata holds metadata for an evaluation plan.
type EvaluationMetadata struct {
	Id string `yaml:"id" json:"id"`
}

// AssessmentPlan links a control to its assessments.
type AssessmentPlan struct {
	Control     PlanMapping  `yaml:"control" json:"control"`
	Assessments []Assessment `yaml:"assessments" json:"assessments"`
}

// Assessment links a requirement to assessment procedures.
type Assessment struct {
	Requirement PlanMapping           `yaml:"requirement" json:"requirement"`
	Procedures  []AssessmentProcedure `yaml:"procedures" json:"procedures"`
}

// AssessmentProcedure defines a specific assessment procedure.
type AssessmentProcedure struct {
	Id            string `yaml:"id" json:"id"`
	Name          string `yaml:"name" json:"name"`
	Description   string `yaml:"description" json:"description"`
	Documentation string `yaml:"documentation" json:"documentation"`
}

// PlanMapping represents a reference mapping used in evaluation plans.
type PlanMapping struct {
	ReferenceId string `yaml:"reference-id" json:"reference-id"`
	EntryId     string `yaml:"entry-id" json:"entry-id"`
	Remarks     string `yaml:"remarks,omitempty" json:"remarks,omitempty"`
}
