package ex

import (
	"os/exec"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/util/logs"

	"github.com/axgle/mahonia"
)

const (
	Python            string = "python"
	PipinstallPiptoos string = "pipenv install pip-tools --skip-lock"
	PipCompilein      string = "pipenv run pip-compile requirements.in"
	PipCompileCfg     string = "pipenv run pip-compile setup.cfg -o temp.txt"
	PipcompileSetup   string = "pipenv run pip-compile setup.py"
	RemoveVirtualCmd  string = "pipenv --rm"
)

type CmdOpts struct {
	Name string
	Args []string
	Dir  string
}

func Do(c string, dir string) (out string, err error) {
	cmd := GetCmdOpts(c, dir).BuildCmd()
	out, err = Excute(cmd)
	if err != nil {
		return
	}
	return
}

func CheckPython(py string) (s string, err error) {
	s, err = exec.LookPath(py)
	if err != nil {
		logs.Error(err)
	}
	return
}

func GetCmdOpts(c string, dir string) *CmdOpts {
	list := strings.Fields(string(c))
	if len(list) <= 1 {
		return &CmdOpts{}
	}
	return &CmdOpts{
		Name: list[0],
		Args: list[1:],
		Dir:  dir,
	}
}

func (c *CmdOpts) BuildCmd() (ec *exec.Cmd) {
	ec = exec.Command(c.Name, c.Args...)
	ec.Dir = c.Dir
	return
}

// 执行
func Excute(cmd *exec.Cmd) (s string, err error) {
	stdoutStderr, err := cmd.CombinedOutput()
	if err != nil {
		logs.Error(err)
		return
	}
	s = Convert(string(stdoutStderr), "gbk", "utf-8")
	return
}

// 编码转换
func Convert(s string, source string, target string) string {
	srcCoder := mahonia.NewDecoder(source)
	res := srcCoder.ConvertString(s)
	t := mahonia.NewDecoder(target)
	_, cdata, _ := t.Translate([]byte(res), true)
	result := string(cdata)
	return result
}
