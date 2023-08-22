package walk

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"
)

// isHttp 是否为http/https协议
func isHttp(url string) bool {
	return strings.HasPrefix(url, "http://") ||
		strings.HasPrefix(url, "https://")
}

// isFtp 是否为ftp协议
func isFtp(url string) bool {
	return strings.HasPrefix(url, "ftp://")
}

// isFile 是否为file协议
func isFile(url string) bool {
	return strings.HasPrefix(url, "file://")
}

// Download 下载文件并保存到目标位置
func Download(origin, output string) error {
	if isHttp(origin) {
		return downloadFromHttp(origin, output)
	} else if isFtp(origin) {
		return downloadFromFtp(origin, output)
	} else if isFile(origin) {
		return downloadFromFile(origin, output)
	} else {
		return copyfile(origin, output)
	}
}

var downloadHttpClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        50,
		MaxConnsPerHost:     50,
		MaxIdleConnsPerHost: 50,
		IdleConnTimeout:     30 * time.Second,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	},
	Timeout: 10 * time.Second,
}

// downloadFromHttp 下载url并保存到目标文件 支持分片下载
func downloadFromHttp(url, output string) error {

	// 获取head
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}
	resp, err := downloadHttpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode > 299 {
		return fmt.Errorf("response code:%d", resp.StatusCode)
	}

	// 创建目标文件
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	// 检测是否支持Accept-Ranges
	if resp.Header.Get("Accept-Ranges") != "bytes" {
		// 不支持分片则尝试直接下载
		r, err := http.Get(url)
		if err != nil {
			return err
		} else {
			defer r.Body.Close()
			io.Copy(f, r.Body)
			return nil
		}
	}

	// 文件总大小
	size, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	offset := 0
	// 分片大小10M
	buffer := 10 * 1024 * 1024

	for offset < size {
		r, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}
		next := offset + buffer
		if next >= size {
			next = size - 1
		}
		r.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", offset, next))
		resp, err := downloadHttpClient.Do(r)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		io.Copy(f, resp.Body)
		offset = next + 1
	}
	return nil
}

// downloadFromFtp 下载url并保存到目标文件
func downloadFromFtp(url, output string) error {
	// 解析参数
	var host, path, username, password string
	host = strings.TrimPrefix(url, "ftp://")
	i := strings.Index(host, "/")
	host, path = host[:i], host[i+1:]
	i = strings.Index(host, "@")
	if i != -1 {
		up := strings.Split(host[:i], ":")
		if len(up) == 2 {
			username, password = up[0], up[1]
		}
		host = host[i+1:]
	}
	// 连接ftp
	c, err := ftp.Dial(host, ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}
	defer func() {
		if err := c.Quit(); err != nil {
			fmt.Println(err)
		}
	}()
	// 登录
	if username != "" && password != "" {
		err = c.Login(username, password)
		if err != nil {
			return err
		}
	}
	// 获取数据
	r, err := c.Retr(path)
	if err != nil {
		return err
	}
	defer r.Close()
	// 创建目标文件
	f, err := os.Create(output)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

// downloadFromFile 下载url并保存到目标文件
func downloadFromFile(url, output string) error {
	input := strings.TrimPrefix(url, "file:///")
	return copyfile(input, output)
}

// copyfile 复制文件
func copyfile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return err
	}
	defer input.Close()
	output, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(output, input)
	return err
}
