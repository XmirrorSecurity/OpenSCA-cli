/*
 * @Descripation: 用于和服务端通信
 * @Date: 2021-12-11 17:48:40
 */

package client

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	mrand "math/rand"
	"net/http"
	"net/url"
	"os"
	"path"
	"regexp"
	"util/args"
	"util/logs"
	"util/temp"

	"github.com/pkg/errors"
)

// 消息响应格式
type SaasReponse struct {
	// 错误消息
	Message string `json:"message"`
	// 状态码 0表示成功
	Code int `json:"code"`
	// 数据体
	Data interface{} `json:"data"`
}

// 检测结果响应格式
type DetectReponse struct {
	// 加密后的消息
	Message string `json:"aesMessage"`
	Tag     string `json:"aesTag"`
	Nonce   string `json:"aesNonce"`
}

// 检测任务请求格式
type DetectRequst struct {
	// 16位byte base64编码
	Tag string `json:"aesTag"`
	// 在saas注册
	Token string `json:"ossToken"`
	// 16位byte base64编码
	Nonce string `json:"aesNonce"`
	// 要发送的数据 aes加密后base64编码
	Message string `json:"aesMessage"`
	// 16位 大写字母
	ClientId string `json:"clientId"`
}

// GetClientId 获取客户端id
func GetClientId() string {
	// 默认id
	id := "XXXXXXXXXXXXXXXX"
	// 尝试读取.key文件
	idFile := path.Join(temp.GetPwd(), ".key")
	if _, err := os.Stat(idFile); err != nil {
		// 文件不存在则生成随机ID并保存
		if f, err := os.Create(idFile); err != nil {
			logs.Error(err)
		} else {
			defer f.Close()
			const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
			idbyte := []byte(id)
			for i := range idbyte {
				idbyte[i] = chars[mrand.Intn(26)]
			}
			f.Write(idbyte)
			id = string(idbyte)
		}
	} else {
		// 文件存在则读取ID
		idbyte, err := os.ReadFile(idFile)
		if err != nil {
			logs.Error(err)
		}
		if len(idbyte) == 16 {
			if ok, err := regexp.Match(`[A-Z]{16}`, idbyte); ok && err == nil {
				id = string(idbyte)
			}
		}
	}
	return id
}

// Detect 发送任务解析请求
func Detect(dtype string, reqbody []byte) (repbody []byte, err error) {
	repbody = []byte{}
	// 获取aes-key
	key, err := getAesKey()
	if err != nil {
		return repbody, err
	}
	// 随机16位子节
	nonce := make([]byte, 16)
	rand.Read(nonce)
	// aes加密
	ciphertext, tag := encrypt(reqbody, key, nonce)
	// 构建请求
	url := args.Config.Url + "/oss-saas/api-v1/open-sca-client/detect"
	// 添加参数
	param := DetectRequst{}
	param.ClientId = GetClientId()
	param.Token = args.Config.Token
	param.Tag = base64.StdEncoding.EncodeToString(tag)
	param.Nonce = base64.StdEncoding.EncodeToString(nonce)
	// base64编码
	param.Message = base64.StdEncoding.EncodeToString(ciphertext)
	data, err := json.Marshal(param)
	if err != nil {
		return repbody, err
	}
	// 发送数据
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Detect-Type", dtype)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return repbody, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		repbody, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			logs.Error(err)
			return
		} else {
			// 解析响应
			saasrep := SaasReponse{}
			err = json.Unmarshal(repbody, &saasrep)
			if err != nil {
				logs.Error(err)
			}
			if saasrep.Code != 0 {
				// 出现错误
				logs.Warn(fmt.Sprintf("url:%s code:%d message: %s", url, saasrep.Code, saasrep.Message))
				err = errors.New(saasrep.Message)
				return
			} else {
				data, err = json.Marshal(saasrep.Data)
				detect := DetectReponse{}
				err = json.Unmarshal([]byte(data), &detect)
				if err != nil {
					logs.Error(err)
				}
				// base64解码后再aes解密
				var ciphertext []byte
				ciphertext, err = base64.StdEncoding.DecodeString(detect.Message)
				tag, err = base64.StdEncoding.DecodeString(detect.Tag)
				nonce, err = base64.StdEncoding.DecodeString(detect.Nonce)
				repbody = decrypt(ciphertext, key, nonce, tag)
				return
			}
		}
	} else {
		return repbody, fmt.Errorf("%s status code: %d", url, resp.StatusCode)
	}
}

// getAesKey 获取aes-key
func getAesKey() (key []byte, err error) {
	u, err := url.Parse(args.Config.Url + "/oss-saas/api-v1/open-sca-client/aes-key")
	if err != nil {
		return key, err
	}
	// 设置参数
	param := url.Values{}
	param.Set("clientId", GetClientId())
	param.Set("ossToken", args.Config.Token)
	u.RawQuery = param.Encode()
	// 发送请求
	rep, err := http.Get(u.String())
	if err != nil {
		logs.Error(err)
		return
	}
	if rep.StatusCode != 200 {
		err = fmt.Errorf("url: %s,status code:%d", u.String(), rep.StatusCode)
		logs.Error(err)
		return
	} else {
		defer rep.Body.Close()
		data, err := ioutil.ReadAll(rep.Body)
		if err != nil {
			logs.Error(err)
			return key, err
		}
		// 获取响应信息
		saasrep := SaasReponse{}
		json.Unmarshal(data, &saasrep)
		if saasrep.Code != 0 {
			// 出现错误
			logs.Warn(fmt.Sprintf("url:%s code:%d message: %s", u.String(), saasrep.Code, saasrep.Message))
			err = errors.New(saasrep.Message)
			return key, err
		} else {
			key = []byte(saasrep.Data.(string))
			return key, nil
		}
	}
}
