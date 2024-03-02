package internal

import (
	"context"

	"github.com/onsi/ginkgo/v2/types"
)

type SpecContext interface {
	context.Context

	SpecReport() types.SpecReport
	AttachProgressReporter(func() string) func()
}

type specContext struct {
	context.Context
	*ProgressReporterManager

<<<<<<< HEAD
<<<<<<< HEAD
	cancel context.CancelCauseFunc
=======
	cancel context.CancelFunc
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
=======
	cancel context.CancelCauseFunc
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)

	suite *Suite
}

/*
SpecContext includes a reference to `suite` and embeds itself in itself as a "GINKGO_SPEC_CONTEXT" value.  This allows users to create child Contexts without having down-stream consumers (e.g. Gomega) lose access to the SpecContext and its methods.  This allows us to build extensions on top of Ginkgo that simply take an all-encompassing context.

Note that while SpecContext is used to enforce deadlines by Ginkgo it is not configured as a context.WithDeadline.  Instead, Ginkgo owns responsibility for cancelling the context when the deadline elapses.

This is because Ginkgo needs finer control over when the context is canceled.  Specifically, Ginkgo needs to generate a ProgressReport before it cancels the context to ensure progress is captured where the spec is currently running.  The only way to avoid a race here is to manually control the cancellation.
*/
func NewSpecContext(suite *Suite) *specContext {
	ctx, cancel := context.WithCancelCause(context.Background())
	sc := &specContext{
		cancel:                  cancel,
		suite:                   suite,
		ProgressReporterManager: NewProgressReporterManager(),
	}
	ctx = context.WithValue(ctx, "GINKGO_SPEC_CONTEXT", sc) //yes, yes, the go docs say don't use a string for a key... but we'd rather avoid a circular dependency between Gomega and Ginkgo
	sc.Context = ctx                                        //thank goodness for garbage collectors that can handle circular dependencies

	return sc
}

func (sc *specContext) SpecReport() types.SpecReport {
	return sc.suite.CurrentSpecReport()
}
