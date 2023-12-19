package format

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	"github.com/xmirrorsecurity/opensca-cli/v3/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/v3/opensca/logs"
)

// Saas 向saas平台发送检测报告
func Saas(report Report) error {

	url := config.Conf().Origin.Url
	token := config.Conf().Origin.Token
	proj := config.Conf().Origin.Proj

	if url == "" || token == "" || proj == nil {
		return nil
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	w.WriteField("token", token)
	w.WriteField("projectUid", *proj)
	w.WriteField("detectOrigin", strconv.Itoa(5))

	uid, err := uuid.NewV6()
	if err != nil {
		return err
	}

	// dsdxFile
	dsdxFile, err := w.CreateFormFile("dsdxFile", uid.String()+".dsdx")
	if err != nil {
		return err
	}
	f := common.CreateTemp("dsdx")
	f.Close()
	defer os.Remove(f.Name())
	Dsdx(report, f.Name())
	dsdx, err := os.Open(f.Name())
	if err != nil {
		return err
	}
	defer dsdx.Close()
	io.Copy(dsdxFile, dsdx)

	// jsonFile
	jsonFile, err := w.CreateFormFile("jsonFile", uid.String()+".json")
	if err != nil {
		return err
	}
	f = common.CreateTemp("json")
	f.Close()
	defer os.Remove(f.Name())
	Json(report, f.Name())
	json, err := os.Open(f.Name())
	if err != nil {
		return err
	}
	defer json.Close()
	io.Copy(jsonFile, json)

	w.Close()

	req, err := http.NewRequest("POST", url+"/oss-saas/api-v1/ide-plugin/sync/result", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logs.Debugf("saas resp: %s", string(data))

	return nil
}
