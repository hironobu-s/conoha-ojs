package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	flag "github.com/ogier/pflag"
)

const (
	AUTH_URL = "https://ident-r1nd1001.cnode.jp/v2.0"
)

type Auth struct {
	*Command

	username   string
	password   string
	tenantname string
}

// コマンドライン引数を処理して返す
func (cmd *Auth) parseFlags() (exitCode int, err error) {

	var showUsage bool
	fs := flag.NewFlagSet("conoha-ojs-auth", flag.ContinueOnError)

	// コマンドライン引数の定義を追加
	fs.BoolVarP(&showUsage, "help", "h", false, "Print usage.")
	fs.StringVarP(&cmd.username, "api-username", "u", "", "API Username")
	fs.StringVarP(&cmd.password, "api-password", "p", "", "API Password")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	if showUsage {
		return ExitCodeUsage, nil
	}

	// ユーザ名、パスワードを未指定の場合はUsageを表示して終了
	if cmd.username == "" || cmd.password == "" {
		return ExitCodeParseFlagError, errors.New("Not enough arguments.")
	}

	// ユーザ名とテナント名は同じ
	cmd.tenantname = cmd.username

	return ExitCodeOK, nil
}

func (cmd *Auth) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s auth [OPTIONS]

Authenticate to ConoHa ObjectStorage.

  -u, --api-username: API Username
  -p: --api-password: API Password

`, lib.COMMAND_NAME)
}

func (cmd *Auth) Run() (exitCode int, err error) {
	exitCode, err = cmd.parseFlags()
	if err != nil || exitCode == ExitCodeUsage {
		cmd.Usage()
		return exitCode, err
	}

	// *lib.Configに割り当て
	var c = cmd.config
	c.ApiUsername = cmd.username
	c.ApiPassword = cmd.password
	c.TenantName = cmd.tenantname

	err = cmd.request(c, c.ApiUsername, c.ApiPassword, c.TenantName)
	if err == nil {
		// アカウント情報を書き出す
		err = c.Save(c.ConfigFilePath())
		if err != nil {
			return ExitCodeError, err
		}

		return ExitCodeOK, nil
	} else {
		return ExitCodeError, err
	}
}

// トークンの有効期限のチェックを行う
// 有効期限内の場合は何もしない
// 有効期限切れの場合は再取得を行う
func (cmd *Auth) CheckTokenIsExpired(c *lib.Config) error {
	log := lib.GetLogInstance()

	// configでユーザ名などが空の場合は先に認証(authコマンド)を実行してくださいと返す
	if c.ApiUsername == "" || c.ApiPassword == "" || c.TenantName == "" {
		err := errors.New("ApiUsername, Apipassword and Tenantname was not found in a config file. You should execute an auth command. (See \"conoha-ojs auth\").")
		return err
	}

	// 以下をすべて満たす場合はキャッシュ済みのトークンを使うため、処理をスキップする
	// * トークンが取得済みである(空文字でない)
	// * エンドポイントURLが取得できている(空文字でない)
	// * トークンの有効期限内である
	doUpdate := false

	if c.Token == "" || c.EndPointUrl == "" {
		doUpdate = true
	}

	now := time.Now().UTC()
	te, err := time.Parse(time.RFC1123, c.TokenExpires)

	if err != nil || now.After(te) {
		doUpdate = true
	}

	if !doUpdate {
		log.Debug("Using the cached token.")
		return nil
	}

	return cmd.request(c, c.ApiUsername, c.ApiPassword, c.TenantName)
}

// 認証を実行して、結果をConfigに書き込む
func (cmd *Auth) request(c *lib.Config, username string, password string, tenantname string) error {

	// アカウント情報
	auth := map[string]interface{}{
		"auth": map[string]interface{}{
			"tenantName": tenantname,
			"passwordCredentials": map[string]interface{}{
				"username": username,
				"password": password,
			},
		},
	}

	b, err := json.Marshal(auth)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		AUTH_URL+"/tokens",
		strings.NewReader(string(b)),
	)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return err
	}

	client := &http.Client{}

	// httpリクエスト実行
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 400:
		msg := fmt.Sprintf("Return %d status code from the server with message. [%s].",
			resp.StatusCode,
			extractErrorMessage(resp.Body),
		)
		return errors.New(msg)
	}

	strjson, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// jsonパース
	err = cmd.parseResponse(strjson, c)
	if err != nil {
		return err
	}

	return nil
}

// レスポンスのJSONをパースする
func (cmd *Auth) parseResponse(strjson []byte, config *lib.Config) error {
	// jsonパース
	var auth map[string]interface{}
	var ok bool
	var err error

	err = json.Unmarshal(strjson, &auth)
	if err != nil {
		return err
	}

	// 認証失敗など
	if _, ok = auth["error"]; ok {
		obj := auth["error"].(map[string]interface{})
		msg := fmt.Sprintf("%s(%0.0f): %s",
			obj["title"].(string),
			obj["code"].(float64),
			obj["message"].(string),
		)

		err = errors.New(msg)
		return err
	}

	// アクセストークンを取得
	if _, ok = auth["access"]; !ok {
		err = errors.New("Undefined index: access")
		return err
	}
	access := auth["access"].(map[string]interface{})

	if _, ok = access["token"]; !ok {
		err = errors.New("Undefined index: token")
		return err
	}
	t := access["token"].(map[string]interface{})
	token := t["id"].(string)

	// トークンの有効期限を取得
	tokenExpires, err := time.Parse(time.RFC3339, t["expires"].(string))
	if err != nil {
		return err
	}

	// テナントIDを取得
	tenant := t["tenant"].(map[string]interface{})
	tenantId := tenant["id"].(string)

	// エンドポイントURLを取得
	var endpointUrl string

	if _, ok = access["serviceCatalog"]; !ok {
		err = errors.New("Undefined index: serviceCatalog")
		return err
	}
	catalogs := access["serviceCatalog"].([]interface{})

	for _, item := range catalogs {
		item2 := item.(map[string]interface{})

		if item2["type"] == "object-store" {
			endpoints := item2["endpoints"].([]interface{})
			endpoint := endpoints[0].(map[string]interface{})

			if _, ok := endpoint["publicURL"]; !ok {
				err = errors.New("Undefined index: publicURL")
				return err
			}

			endpointUrl = endpoint["publicURL"].(string)
		}
	}

	// *lib.Configに割り当て
	config.Token = token
	config.TokenExpires = tokenExpires.Format(time.RFC1123)
	config.EndPointUrl = endpointUrl
	config.TenantId = tenantId

	return nil
}
