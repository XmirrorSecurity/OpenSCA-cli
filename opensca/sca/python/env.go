package python

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

// ParseRequirementWithEnv 使用pipenv解析requirement文件
func ParseRequirementWithEnv(ctx context.Context, file *model.File) *model.DepGraph {
	runEnv(ctx, file.Abspath())
	return nil
}

// ParseRequirementInWithEnv 使用pipenv解析requirement.in文件
func ParseRequirementInWithEnv(ctx context.Context, file *model.File) *model.DepGraph {
	runEnv(ctx, file.Abspath())
	return nil
}

// ParseSetupCfgWithEnv 使用pipenv解析setup.cfg文件
func ParseSetupCfgWithEnv(ctx context.Context, file *model.File) *model.DepGraph {
	runEnv(ctx, file.Abspath())
	return nil
}

// ParseSetupPyWithEnv 使用pipenv解析setup.py文件
func ParseSetupPyWithEnv(ctx context.Context, file *model.File) *model.DepGraph {
	runEnv(ctx, file.Abspath())
	return nil
}

func runEnv(ctx context.Context, file string) (data []byte, ok bool) {

	if _, err := exec.LookPath("python"); err != nil {
		return
	}
	if _, err := exec.LookPath("pipenv"); err != nil {
		return
	}

	dir, name := filepath.Split(file)
	if _, ok = runCmd(ctx, dir, "pipenv", "install", "pip-tools"); !ok {
		return
	}
	defer runCmd(ctx, dir, "pipenv", "--rm")

	return runCmd(ctx, dir, "pipenv", "run", "pip-compile", name)
}

func runCmd(ctx context.Context, dir string, args ...string) ([]byte, bool) {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return out, err != nil
}
