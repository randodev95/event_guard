package impact

import (
	"github.com/randodev95/event_guard/internal/storage"
	"github.com/randodev95/event_guard/pkg/ast"
	"github.com/randodev95/event_guard/pkg/normalization"
	"github.com/randodev95/event_guard/pkg/validator"
)

// Breach represents a validation failure during a parity check.
type Breach struct {
	EventName string
	Errors    []string
}

// CheckParity validates previously captured event snapshots against a new version of the tracking plan.
// This identifies breaking changes that would affect data integrity if the new plan is deployed.
func CheckParity(db *storage.DB, prevSHA string, plan *ast.TrackingPlan) ([]Breach, error) {
	snapshots, err := db.GetSnapshots(prevSHA)
	if err != nil {
		return nil, err
	}

	var breaches []Breach

	for _, snapshot := range snapshots {
		schema, err := plan.ResolveEventSchema(snapshot.EventName)
		if err != nil {
			breaches = append(breaches, Breach{
				EventName: snapshot.EventName,
				Errors:    []string{err.Error()},
			})
			continue
		}

		for _, payload := range snapshot.Payloads {
			mapper := normalization.NewDefaultMapper()
			norm, err := mapper.Map([]byte(payload))
			if err != nil {
				return nil, err
			}

			result, err := validator.Validate(norm, schema)
			if err != nil {
				return nil, err
			}

			if !result.Valid {
				breaches = append(breaches, Breach{
					EventName: snapshot.EventName,
					Errors:    result.Errors,
				})
				break
			}
		}
	}

	return breaches, nil
}
