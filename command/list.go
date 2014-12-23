package command

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	"net/http"
	"os"
)

type List struct {
	// 表示するコンテナ名
	containerName string

	*Command
}

func NewList() *List {
	cmd := new(List)
	return cmd
}

// コマンドライン引数を処理して、一覧表示するコンテナ名を返す
// 引数が省略された場合はルートを決め打ちする
func (cmd *List) parseFlags() error {

	if len(os.Args) == 2 {
		// コンテナ名が指定されなかったときはルートを決め打ちする
		cmd.containerName = "/"

	} else if len(os.Args) < 2 {
		// 引数が足りない
		msg := fmt.Sprintf(`Usage: %s list <container_or_object>

List container or object.

<container_or_object> Name of container or object.

`, COMMAND_NAME)

		return errors.New(msg)

	} else {
		cmd.containerName = os.Args[2]
	}

	return nil
}

// コマンドを実行する
func (cmd *List) Run(cfg *lib.Config) (err error) {

	err = cmd.parseFlags()
	if err != nil {
		return err
	}

	list, err := cmd.List(cfg, cmd.containerName)
	if err != nil {
		return err
	}

	for _, item := range list {
		fmt.Fprintf(os.Stdout, "%s\n", item)
	}

	return nil
}

//  コンテナやオブジェクトを取得のリストを返す
func (cmd *List) List(cfg *lib.Config, container string) (objects []string, err error) {

	// URLを検証する
	// rawurl := cfg.EndPointUrl + "/" + neturl.QueryEscape(container)
	// url, err := neturl.ParseRequestURI(rawurl)
	url, err := buildStorageUrl(cfg.EndPointUrl, container)
	if err != nil {
		return nil, err
	}

	// リクエストを作成
	req, err := http.NewRequest(
		"GET",
		url.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Auth-Token", cfg.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// HTTPステータスコードがエラーを返した場合
	switch {
	case resp.StatusCode == 404:
		return nil, errors.New("Object or Container was not found.")
	case resp.StatusCode >= 400:
		return nil, errors.New("Return error code from Server.")
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		objects = append(objects, scanner.Text())
	}

	return objects, nil
}
