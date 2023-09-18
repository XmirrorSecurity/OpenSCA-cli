package python

import (
	"context"
	"os/exec"
	"path/filepath"

	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
)

/*
  TODO: 调用pipenv解析依赖
*/

// ParseRequirementWithEnv 使用pipenv解析requirement文件
func ParseRequirementWithEnv(ctx context.Context, file *model.File) *model.DepGraph {
	runEnv(ctx, file.Abspath(), nil)
	return nil
}

// ParseRequirementInWithEnv 使用pipenv解析requirement.in文件
func ParseRequirementInWithEnv(ctx context.Context, file *model.File) *model.DepGraph {
	runEnv(ctx, file.Abspath(), nil)
	return nil
}

// ParseSetupCfgWithEnv 使用pipenv解析setup.cfg文件
func ParseSetupCfgWithEnv(ctx context.Context, file *model.File) *model.DepGraph {
	runEnv(ctx, file.Abspath(), nil)
	return nil
}

// ParseSetupPyWithEnv 使用pipenv解析setup.py文件
func ParseSetupPyWithEnv(ctx context.Context, file *model.File) *model.DepGraph {
	runEnv(ctx, file.Abspath(), nil)
	return nil
}

func runEnv(ctx context.Context, file string, stdout func([]byte)) {

	if _, err := exec.LookPath("python"); err != nil {
		return
	}
	if _, err := exec.LookPath("pipenv"); err != nil {
		return
	}

	dir, name := filepath.Split(file)
	if !runCmd(ctx, dir, nil, "pipenv", "install", "pip-tools", "--skip-lock") {
		return
	}
	defer runCmd(ctx, dir, nil, "pipenv", "--rm")

	runCmd(ctx, dir, stdout, "pipenv", "run", "pip-compile", name)
}

func runCmd(ctx context.Context, dir string, stdout func(data []byte), args ...string) bool {
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return false
	}
	if stdout != nil {
		stdout(out)
	}
	return true
}
