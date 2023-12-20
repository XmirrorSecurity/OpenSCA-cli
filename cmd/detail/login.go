package detail

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
	"golang.org/x/term"
)

func Login() error {
	fmt.Println("Log in with your username to access cloud-based software supply-chain risk data from OpenSCA SaaS.")
	fmt.Println("If you don't have an account, please register at https://opensca.xmirror.cn/")

	fmt.Print("Enter username or email: ")
	username, err := bufio.NewReader(os.Stdin).ReadString('\n')
	username = strings.TrimRight(username, "\r\n")
	if err != nil {
		return err
	}

	fmt.Print("Enter password: ")
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	password = bytes.TrimRight(password, "\r\n")
	if err != nil {
		return err
	}

	m := md5.New()
	m.Write(password)
	pswdmd5 := hex.EncodeToString(m.Sum(nil))

	fmt.Printf("\n%s login ...\n", username)

	url := config.Conf().Origin.Url + "/oss-saas/api-v1/open-sca-client/token"
	url += fmt.Sprintf("?usernameOrEmail=%s&password=%s", username, pswdmd5)

	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logs.Debugf("login response: %s", string(data))

	loginResp := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}{}
	json.Unmarshal(data, &loginResp)
	if loginResp.Code == 0 && loginResp.Message == "success" {
		config.Conf().Origin.Token = loginResp.Data
	} else {
		return errors.New(loginResp.Message)
	}

	return nil
}
