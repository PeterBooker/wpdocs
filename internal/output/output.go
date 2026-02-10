package output

import (
	"github.com/peter/wpdocs/internal/model"
)

// Generator is the interface for documentation output backends.
type Generator interface {
	Generate(reg *model.Registry) error
}
