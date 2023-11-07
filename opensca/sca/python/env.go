package python

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

// ParsePythonWithEnv 借助pipenv解析python依赖
func ParsePythonWithEnv(ctx context.Context, file *model.File) *model.DepGraph {

	if _, err := exec.LookPath("python"); err != nil {
		return nil
	}
	if _, err := exec.LookPath("pipenv"); err != nil {
		return nil
	}

	// 复制到临时目录
	tempdir := common.MkdirTemp("pipenv")
	tempfile := filepath.Join(tempdir, filepath.Base(file.Abspath()))
	src, _ := os.Open(file.Abspath())
	dst, _ := os.Create(tempfile)
	io.Copy(dst, src)
	src.Close()
	dst.Close()
	defer os.RemoveAll(tempdir)

	dir, name := filepath.Split(tempfile)

	if filter.PythonRequirementsTxt(name) {
		runCmd(ctx, dir, "pipenv", "install", "-r", name, "-i", "https://pypi.tuna.tsinghua.edu.cn/simple")
	} else if filter.PythonPipfile(name) {
		runCmd(ctx, dir, "pipenv", "install", "-i", "https://pypi.tuna.tsinghua.edu.cn/simple")
	} else {
		return nil
	}

	defer runCmd(ctx, dir, "pipenv", "--rm")
	root := pipenvGraph(ctx, dir)
	root.Path = file.Relpath()
	return root
}

func pipenvGraph(ctx context.Context, dir string) *model.DepGraph {

	data, ok := runCmd(ctx, dir, "pipenv", "graph")
	if !ok || len(data) == 0 {
		return nil
	}

	root := &model.DepGraph{}

	s := bufio.NewScanner(bytes.NewReader(data))
	tops := []*model.DepGraph{}

	for s.Scan() {

		line := strings.TrimRight(s.Text(), "\r\n")

		// 当前层级
		level := 0
		// 空格个数
		space := 0
		for i := range line {
			if ('a' <= line[i] && line[i] <= 'z') || ('A' <= line[i] && line[i] <= 'Z') {
				level = (space + 3) / 4
				line = line[i:]
				break
			}
			if line[i] == ' ' {
				space++
			}
		}

		dep := &model.DepGraph{}

		if level == 0 {
			tags := strings.Split(line, "==")
			if len(tags) == 2 {
				dep.Name = tags[0]
				dep.Version = tags[1]
			} else {
				logs.Warnf("parse pipenv graph err in line: %s", line)
				dep.Name = line
			}
			root.AppendChild(dep)
		} else {
			i := strings.Index(line, " ")
			if i == -1 {
				logs.Warnf("parse pipenv graph err in line: %s", line)
				dep.Name = line
			} else {
				dep.Name = line[:i]
			}
			line = strings.Trim(line[i+1:], "[]")
			i = strings.LastIndex(line, "installed: ")
			if i != -1 {
				dep.Version = line[i+len("installed: "):]
			}
		}

		tops = append(tops[:level], dep)
		if level > 0 {
			tops[level-1].AppendChild(dep)
		}

	}

	return root
}

func runCmd(ctx context.Context, dir string, cmd string, args ...string) ([]byte, bool) {
	c := exec.CommandContext(ctx, cmd, args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	return out, err == nil
}
