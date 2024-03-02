/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package wait

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
)

// loopConditionUntilContext executes the provided condition at intervals defined by
// the provided timer until the provided context is cancelled, the condition returns
// true, or the condition returns an error. If sliding is true, the period is computed
// after condition runs. If it is false then period includes the runtime for condition.
<<<<<<< HEAD
<<<<<<< HEAD
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
// If immediate is false the first delay happens before any call to condition, if
// immediate is true the condition will be invoked before waiting and guarantees that
// the condition is invoked at least once, regardless of whether the context has been
// cancelled. The returned error is the error returned by the last condition or the
// context error if the context was terminated.
<<<<<<< HEAD
=======
// If immediate is false the first delay happens before any call to condition. The
// returned error is the error returned by the last condition or the context error if
// the context was terminated.
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
//
// This is the common loop construct for all polling in the wait package.
func loopConditionUntilContext(ctx context.Context, t Timer, immediate, sliding bool, condition ConditionWithContextFunc) error {
	defer t.Stop()

	var timeCh <-chan time.Time
	doneCh := ctx.Done()

<<<<<<< HEAD
<<<<<<< HEAD
	if !sliding {
		timeCh = t.C()
	}

	// if immediate is true the condition is
	// guaranteed to be executed at least once,
	// if we haven't requested immediate execution, delay once
	if immediate {
		if ok, err := func() (bool, error) {
			defer runtime.HandleCrash()
			return condition(ctx)
		}(); err != nil || ok {
			return err
		}
	}

	if sliding {
		timeCh = t.C()
	}

	for {

		// Wait for either the context to be cancelled or the next invocation be called
=======
	// if we haven't requested immediate execution, delay once
	if !immediate {
		timeCh = t.C()
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
=======
	if !sliding {
		timeCh = t.C()
	}

	// if immediate is true the condition is
	// guaranteed to be executed at least once,
	// if we haven't requested immediate execution, delay once
	if immediate {
		if ok, err := func() (bool, error) {
			defer runtime.HandleCrash()
			return condition(ctx)
		}(); err != nil || ok {
			return err
		}
	}

	if sliding {
		timeCh = t.C()
	}

	for {

		// Wait for either the context to be cancelled or the next invocation be called
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		select {
		case <-doneCh:
			return ctx.Err()
		case <-timeCh:
		}
<<<<<<< HEAD
<<<<<<< HEAD

		// IMPORTANT: Because there is no channel priority selection in golang
		// it is possible for very short timers to "win" the race in the previous select
		// repeatedly even when the context has been canceled.  We therefore must
		// explicitly check for context cancellation on every loop and exit if true to
		// guarantee that we don't invoke condition more than once after context has
		// been cancelled.
=======
	}

	for {
		// checking ctx.Err() is slightly faster than checking a select
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
=======

		// IMPORTANT: Because there is no channel priority selection in golang
		// it is possible for very short timers to "win" the race in the previous select
		// repeatedly even when the context has been canceled.  We therefore must
		// explicitly check for context cancellation on every loop and exit if true to
		// guarantee that we don't invoke condition more than once after context has
		// been cancelled.
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
		if err := ctx.Err(); err != nil {
			return err
		}

		if !sliding {
			t.Next()
		}
		if ok, err := func() (bool, error) {
			defer runtime.HandleCrash()
			return condition(ctx)
		}(); err != nil || ok {
			return err
		}
		if sliding {
			t.Next()
		}
<<<<<<< HEAD
<<<<<<< HEAD
=======

		if timeCh == nil {
			timeCh = t.C()
		}

		// NOTE: b/c there is no priority selection in golang
		// it is possible for this to race, meaning we could
		// trigger t.C and doneCh, and t.C select falls through.
		// In order to mitigate we re-check doneCh at the beginning
		// of every loop to guarantee at-most one extra execution
		// of condition.
		select {
		case <-doneCh:
			return ctx.Err()
		case <-timeCh:
		}
>>>>>>> fe31cef4 (Update vendor github.com/MottainaiCI/lxd-compose@d928eed0eddfde18d58fe3a8ae780328c1b0d55c)
=======
>>>>>>> d5bb6cf2 (Upgrade vendor github.com/MottainaiCI/lxd-compose@v0.33.0)
	}
}
