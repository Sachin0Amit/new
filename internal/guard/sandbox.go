package guard

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Sandbox struct {
	Enforcer  *CapabilityEnforcer
	DataRoot  string
}

func NewSandbox(enforcer *CapabilityEnforcer, dataRoot string) *Sandbox {
	absRoot, _ := filepath.Abs(dataRoot)
	return &Sandbox{
		Enforcer: enforcer,
		DataRoot: absRoot,
	}
}

func (s *Sandbox) SandboxedReadFile(ctx context.Context, path string) ([]byte, error) {
	if err := s.Enforcer.Enforce(ctx, CapReadFiles); err != nil {
		return nil, err
	}

	safePath, err := s.validatePath(path)
	if err != nil {
		return nil, err
	}

	return ioutil.ReadFile(safePath)
}

func (s *Sandbox) SandboxedWriteFile(ctx context.Context, path string, data []byte) error {
	if err := s.Enforcer.Enforce(ctx, CapWriteFiles); err != nil {
		return err
	}

	safePath, err := s.validatePath(path)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(safePath, data, 0600)
}

func (s *Sandbox) validatePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", ErrPathTraversal
	}

	// Follow symlinks and re-evaluate
	realPath, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		// If it doesn't exist, we still check the prefix of the parent
		realPath = absPath
	}

	if !strings.HasPrefix(realPath, s.DataRoot) {
		return "", ErrPathTraversal
	}

	return realPath, nil
}
