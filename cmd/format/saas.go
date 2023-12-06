package format

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	uuid "github.com/satori/go.uuid"
	"github.com/xmirrorsecurity/opensca-cli/cmd/config"
	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/logs"
)

// Saas 向saas平台发送检测报告
func Saas(report Report, token, uid string) error {

	if token == "" || uid == "" {
		return nil
	}

	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	w.WriteField("token", token)
	w.WriteField("projectUid", uid)
	w.WriteField("detectOrigin", strconv.Itoa(5))

	// dsdxFile
	dsdxFile, err := w.CreateFormFile("dsdxFile", uuid.NewV4().String()+".dsdx")
	if err != nil {
		return err
	}
	f := common.CreateTemp("dsdx")
	f.Close()
	Dsdx(report, f.Name())
	dsdx, err := os.Open(f.Name())
	if err != nil {
		return err
	}
	io.Copy(dsdxFile, dsdx)
	dsdx.Close()
	os.Remove(f.Name())

	// jsonFile
	jsonFile, err := w.CreateFormFile("jsonFile", uuid.NewV4().String()+".json")
	if err != nil {
		return err
	}
	f = common.CreateTemp("json")
	f.Close()
	Json(report, f.Name())
	json, err := os.Open(f.Name())
	if err != nil {
		return err
	}
	io.Copy(jsonFile, json)
	json.Close()
	os.Remove(f.Name())

	url := config.Conf().Origin.Url + "/oss-saas/api-v1/ide-plugin/sync/result"
	req, err := http.NewRequest("POST", url, body)
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
