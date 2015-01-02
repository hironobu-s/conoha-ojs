package command

import (
	"github.com/hironobu-s/conoha-ojs/lib"
	"io"
	"io/ioutil"
	"net/url"
	"strings"
)

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

	log.Debug(rawurl)

	return url.Parse(rawurl)
}

// オブジェクトストレージが返すエラーレスポンスのHTMLデータからメッセージ部分を抜き出す
// メッセージを抜き出せなかった場合は空文字を返す
func extractErrorMessage(Body io.ReadCloser) string {

	b, _ := ioutil.ReadAll(Body)
	rawhtml := string(b)

	pb := strings.Index(rawhtml, "<p>")
	if pb < 0 {
		return rawhtml
	}

	pe := strings.Index(rawhtml, "</p>")
	if pe < 0 {
		return rawhtml
	}

	return rawhtml[pb+3 : pe]
}
