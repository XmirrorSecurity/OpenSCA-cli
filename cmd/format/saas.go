package format

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	// dsdx
	dsdxWriter, err := w.CreateFormFile("dsdxFile", uid.String()+".dsdx")
	if err != nil {
		return err
	}
	f := common.CreateTemp("dsdx")
	f.Close()
	defer os.Remove(f.Name())
	Dsdx(report, f.Name())
	dsdxFile, err := os.Open(f.Name())
	if err != nil {
		return err
	}
	defer dsdxFile.Close()
	io.Copy(dsdxWriter, dsdxFile)

	// json
	jsonWriter, err := w.CreateFormFile("jsonFile", uid.String()+".json")
	if err != nil {
		return err
	}
	f = common.CreateTemp("json")
	f.Close()
	defer os.Remove(f.Name())
	Json(report, f.Name())
	jsonFile, err := os.Open(f.Name())
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	io.Copy(jsonWriter, jsonFile)

	w.Close()

	req, err := http.NewRequest("POST", url+"/oss-saas/api-v1/ide-plugin/sync/result", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := common.HttpDownloadClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	logs.Debugf("saas resp: %s", string(data))
	saasResp := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}{}
	json.Unmarshal(data, &saasResp)
	if saasResp.Code == 0 && saasResp.Message == "success" {
		logs.Infof("saas url: %s/%s", url, saasResp.Data)
		fmt.Printf("saas url: %s/%s\n", url, saasResp.Data)
	}

	return nil
}
