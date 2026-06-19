package builder

import "context"

// Fake is a Builder for tests. It records the spec and params it was built with
// and counts Build calls.
type Fake struct {
	Spec   Spec
	Params []byte
	Err    error
	Built  int
}

// NewFake returns a builder constructor (matching the signature of New) that
// produces Fakes appended to *sink, so tests can inspect what would be built.
func NewFake(sink *[]*Fake) func(kind string, params []byte, spec Spec) (Builder, error) {
	return func(_ string, params []byte, spec Spec) (Builder, error) {
		f := &Fake{Spec: spec, Params: params}
		*sink = append(*sink, f)
		return f, nil
	}
}

// Build implements Builder.
func (f *Fake) Build(_ context.Context) error {
	f.Built++
	return f.Err
}
