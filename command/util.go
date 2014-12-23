package command

import (
	"github.com/hironobu-s/conoha-ojs/lib"
	"net/url"
	"path/filepath"
	"strings"
)

// 引数で渡された文字列を解決して、ローカルの絶対パスを返す
func (cmd *Download) resolveLocalPath(paths ...string) (abspath string, err error) {

	// 引数の文字列を連結してパスを作る
	p := strings.Join(paths, string(filepath.Separator))

	// 正規化する
	p = filepath.Clean(p)

	// 絶対パスを取得
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}

	return abs, nil
}

// 引数で渡された文字列を解決して、オブジェクトストレージのURIを返す
func buildStorageUrl(endpointUrl string, paths ...string) (url *url.URL, err error) {
	log := lib.GetLogInstance()

	// オブジェクトストレージのURIを構築する
	rawurl := endpointUrl

	// EndPointUrl の末尾のスラッシュを削除
	if strings.HasSuffix(rawurl, "/") {
		rawurl = rawurl[0 : len(rawurl)-1]
	}

	// パスを連結する先頭のスラッシュを補完
	for i := 0; i < len(paths); i++ {
		paths[i] = strings.Trim(paths[i], "/")
	}

	rawurl += "/" + strings.Join(paths, "/")

	log.Infof("%v", rawurl)

	return url.Parse(rawurl)
}
