package tick

import (
	"context"
)

// Processor defines an interface for processing ticks.
type Processor interface {
	// teamID is optional; if non-empty, the nuclear option will be executed.
	ProcessTick(ctx context.Context, teamID string) error
}
