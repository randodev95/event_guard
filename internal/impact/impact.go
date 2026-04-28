package impact

import (
	"github.com/eventcanvas/eventcanvas/pkg/normalization"
	"github.com/eventcanvas/eventcanvas/internal/storage"
	"github.com/eventcanvas/eventcanvas/pkg/validator"
	"github.com/eventcanvas/eventcanvas/pkg/ast"
)

type Breach struct {
	EventName string
	Errors    []string
}

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
			norm, err := normalization.Normalize([]byte(payload))
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
