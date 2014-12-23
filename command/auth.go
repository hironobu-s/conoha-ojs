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
	"io"
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

func NewAuth(stdSteram io.Writer, errStream io.Writer) (cmd *Auth) {
	cmd = &Auth{
		Command: NewCommand(stdSteram, errStream),
	}
	return cmd
}

// コマンドライン引数を処理して返す
func (cmd *Auth) parseFlags() error {

	// コマンドライン引数の定義を追加
	flag.StringVarP(&cmd.username, "api-username", "u", "", "API Username")
	flag.StringVarP(&cmd.password, "api-password", "p", "", "API Password")

	// // コマンドライン引数をパース
	// //fs.Parse(os.Args[2:])
	os.Args = os.Args[1:]
	flag.Parse()

	// ユーザ名、パスワードを未指定の場合はUsageを表示して終了
	if cmd.username == "" || cmd.password == "" {
		return errors.New("Not enough arguments.")
	}

	// ユーザ名とテナント名は同じ
	cmd.tenantname = cmd.username

	return nil
}

func (cmd *Auth) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s auth [OPTIONS]

Authenticate to ConoHa ObjectStorage.

  -u, --api-username: API Username
  -p: --api-password: API Password
`, COMMAND_NAME)
}

// コマンドとして実行する
func (cmd *Auth) Run(c *lib.Config) (exitCode int, err error) {
	// コマンドライン引数を処理
	err = cmd.parseFlags()
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	// *lib.Configに割り当て
	c.ApiUsername = cmd.username
	c.ApiPassword = cmd.password
	c.TenantName = cmd.tenantname

	err = cmd.doAuth(c, c.ApiUsername, c.ApiPassword, c.TenantName)
	if err == nil {
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
		log.Info("Using the cached token.")
		return nil
	}

	return cmd.doAuth(c, c.ApiUsername, c.ApiPassword, c.TenantName)
}

// 認証を実行して、結果をConfigに書き込む
func (cmd *Auth) doAuth(c *lib.Config, username string, password string, tenantname string) error {

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
