package command

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	flag "github.com/ogier/pflag"
	"net/http"
	"os"
)

type List struct {
	// List表示するコンテナ名
	containerName string

	*Command
}

// コマンドライン引数を処理して、一覧表示するコンテナ名を返す
// 引数が省略された場合はルートを決め打ちする
func (cmd *List) parseFlags() (exitCode int, err error) {

	var showUsage bool
	fs := flag.NewFlagSet("conoha-ojs-list", flag.ContinueOnError)
	fs.BoolVarP(&showUsage, "help", "h", false, "Print usage.")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	if showUsage {
		return ExitCodeUsage, nil
	}

	if len(os.Args) == 2 {
		// コンテナ名が指定されなかったときはルートを決め打ちする
		cmd.containerName = "/"

	} else if len(os.Args) < 2 {
		return ExitCodeParseFlagError, errors.New("Not enough arguments.")

	} else {
		cmd.containerName = os.Args[2]
	}

	return ExitCodeOK, nil
}

func (cmd *List) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s list <container_or_object>

List container or object.

<container_or_object> Name of container or object.

`, lib.COMMAND_NAME)
}

// コマンドを実行する
func (cmd *List) Run() (exitCode int, err error) {

	exitCode, err = cmd.parseFlags()
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	if exitCode == ExitCodeUsage {
		cmd.Usage()
		return exitCode, nil
	}

	list, err := cmd.List(cmd.containerName)
	if err != nil {
		return ExitCodeError, err
	}

	for _, item := range list {
		fmt.Fprintf(cmd.stdStream, "%s\n", item)
	}

	return ExitCodeOK, nil
}

//  コンテナやオブジェクトを取得のリストを返す
func (cmd *List) List(container string) (objects []string, err error) {

	// URLを検証する
	// rawurl := c.EndPointUrl + "/" + neturl.QueryEscape(container)
	// url, err := neturl.ParseRequestURI(rawurl)
	url, err := buildStorageUrl(cmd.config.EndPointUrl, container)
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
	req.Header.Set("X-Auth-Token", cmd.config.Token)

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
