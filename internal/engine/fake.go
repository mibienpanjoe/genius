package engine

import "context"

// Fake is a deterministic in-memory Engine for tests. If GenFunc is set it is
// called; otherwise Reply is returned. The last prompts are recorded.
type Fake struct {
	Reply   string
	Err     error
	GenFunc func(sys, user string) (string, error)

	LastSys  string
	LastUser string
	Calls    int

	DescribeReply string
	DescribeErr   error
	LastImage     string
}

func (f *Fake) Name() string { return "fake" }

func (f *Fake) Generate(_ context.Context, sys, user string) (string, error) {
	f.LastSys = sys
	f.LastUser = user
	f.Calls++
	if f.GenFunc != nil {
		return f.GenFunc(sys, user)
	}
	return f.Reply, f.Err
}

// DescribeReply is returned by Describe; DescribeErr (e.g. ErrNoVision) takes
// precedence when set.
func (f *Fake) Describe(_ context.Context, imagePath, _ string) (string, error) {
	f.LastImage = imagePath
	if f.DescribeErr != nil {
		return "", f.DescribeErr
	}
	return f.DescribeReply, nil
}
