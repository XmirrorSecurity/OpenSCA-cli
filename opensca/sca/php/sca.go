package php

import (
	"context"
	"encoding/json"
	"io"
	"path"
	"strings"

	"github.com/xmirrorsecurity/opensca-cli/opensca/common"
	"github.com/xmirrorsecurity/opensca-cli/opensca/model"
	"github.com/xmirrorsecurity/opensca-cli/opensca/sca/filter"
)

type Sca struct{}

func (sca Sca) Language() model.Language {
	return model.Lan_Php
}

func (sca Sca) Filter(relpath string) bool {
	return filter.PhpComposer(relpath) || filter.PhpComposerLock(relpath)
}

func (sca Sca) Sca(ctx context.Context, parent *model.File, files []*model.File, call model.ResCallback) {

	jsonMap := map[string]*ComposerJson{}
	lockMap := map[string]*ComposerLock{}

	path2dir := func(relpath string) string { return path.Dir(strings.ReplaceAll(relpath, `\`, `/`)) }

	for _, f := range files {
		if filter.PhpComposer(f.Relpath()) {
			f.OpenReader(func(reader io.Reader) {
				var js ComposerJson
				json.NewDecoder(reader).Decode(&js)
				js.File = f
				jsonMap[path2dir(f.Relpath())] = &js
			})
		} else if filter.PhpComposerLock(f.Relpath()) {
			f.OpenReader(func(reader io.Reader) {
				var lock ComposerLock
				json.NewDecoder(reader).Decode(&lock)
				lockMap[path2dir(f.Relpath())] = &lock
			})
		}
	}

loop:
	for dir, json := range jsonMap {

		// 通过lock文件补全
		if lock, ok := lockMap[dir]; ok {
			call(json.File, ParseComposerJsonWithLock(json, lock))
			continue
		}

		// vendor中的composer.json没有对应lock则不做处理
		for _, dir := range strings.Split(dir, "/") {
			if strings.EqualFold(dir, "vendor") {
				continue loop
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		// 从数据源下载
		call(json.File, ParseComposerJsonWithOrigin(json))
	}
}

var defaultComposerRepo = []common.RepoConfig{
	{Url: "http://repo.packagist.org/p2"},
}

func RegisterComposerRepo(repos ...common.RepoConfig) {
	newRepo := common.TrimRepo(repos...)
	if len(newRepo) > 0 {
		defaultComposerRepo = newRepo
	}
}
