package command

// コンテナやオブジェクトのメタデータを操作する
// わりと仕様がめんどくさい。以下を参照
//
// http://docs.openstack.org/api/openstack-object-storage/1.0/content/object-metadata.html
// http://docs.openstack.org/api/openstack-object-storage/1.0/content/Update_Container_Metadata-d1e1900.html
// http://docs.openstack.org/api/openstack-object-storage/1.0/content/delete-container-metadata.html

import (
	"errors"
	"fmt"
	"../lib"
	flag "github.com/ogier/pflag"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// 引数-mで複数のデータを受け取れるようにする。
// http://qiita.com/hironobu_s/items/96e8397ec453dfb976d4
type strmap map[string]string

func (s *strmap) String() string {
	return "" //fmt.Sprintf("%v", s)
}

func (s *strmap) Set(arg string) error {

	keyvalue := strings.Split(arg, ":")
	if len(keyvalue) != 2 {
		return errors.New(fmt.Sprintf("\"%s\" is invalid metadata.", arg))
	}

	smap := *s
	smap[keyvalue[0]] = keyvalue[1]

	return nil
}

type Post struct {
	objectName string
	metadatas  strmap
	readAcl    string
	writeAcl   string

	*Command
}

// 引数-mでメタデータを指定する
// メタデータはコロン区切りの key:value の形式で渡す。
func (cmd *Post) parseFlags() (exitCode int, err error) {

	// 初期化
	cmd.metadatas = strmap{}

	var showUsage bool

	fs := flag.NewFlagSet("conoha-ojs-post", flag.ContinueOnError)

	fs.BoolVarP(&showUsage, "help", "h", false, "Print usage.")
	fs.VarP(&cmd.metadatas, "meta", "m", "Key and value of metadata.")
	fs.StringVarP(&cmd.readAcl, "read-acl", "r", "_no_assign_", "Read ACL for containers.")
	fs.StringVarP(&cmd.writeAcl, "write-acl", "w", "_no_assign_", "Write ACL for containers.")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	if showUsage {
		return ExitCodeUsage, nil
	}

	// メタデータを操作するオブジェクト
	cmd.objectName = fs.Arg(0)
	if cmd.objectName == "" {
		return ExitCodeParseFlagError, errors.New("Not enough arguments.")
	}

	return ExitCodeOK, nil
}

func (cmd *Post) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s post <container or object>

Update meta datas for the container, or object.
If the container is not found, it will be created automatically.

<container or object>  Name of container or object to post to.

  -m, --meta:      Set a meta data item. This option may be repeated.
                   Example: -m Hoge:Fuga -m Foo:Bar
                   If value is blank, a meta data will be deleted.
                   Example: -m Hoge:

  -r, --read-acl:  Set Read ACL for containers. 
                   Example: -r ".r:*,.rlistings"

  -w, --write-acl: Set Write ACL for containers.
                   Example: -w "account1, account2"

`, lib.COMMAND_NAME)
}

func (cmd *Post) Run() (exitCode int, err error) {

	exitCode, err = cmd.parseFlags()
	if err != nil || exitCode == ExitCodeUsage {
		cmd.Usage()
		return exitCode, err
	}

	err = cmd.Post(cmd.objectName)
	if err != nil {
		return ExitCodeError, err
	}

	return ExitCodeOK, err
}

func (cmd *Post) Post(path string) error {

	// stat して対象が存在するか調べる
	s := NewCommand("stat", cmd.config, cmd.stdStream, cmd.errStream).(*Stat)
	item, err := s.Stat(path)
	if err == nil {
		// 対象が存在している
		err = cmd.request("POST", item)

	} else {
		// エラーの場合は存在しないと仮定してコンテナを作成する
		item = &Container{
			Container: path,
		}
		err = cmd.request("PUT", item)
	}

	if err != nil {
		return err
	}

	return nil
}

// PostやPutに使うURLを構築
func (cmd *Post) storageUrl(item Item) (u *url.URL, err error) {

	_, isContainer := item.(*Container)

	// URLを構築
	if isContainer {
		return buildStorageUrl(cmd.config.EndPointUrl, item.(*Container).Container)
	} else {
		return buildStorageUrl(cmd.config.EndPointUrl, item.(*Object).Object)
	}
}

// 引数のhttp.Requestに、メタデータ、ReadACL, WriteACLのヘッダ情報を追加する
func (cmd *Post) addHeaders(req *http.Request, item Item) {

	log := lib.GetLogInstance()
	_, isContainer := item.(*Container)

	// メタデータをセット
	for name, value := range cmd.metadatas {
		var header = "X-"

		// valueが空の場合はメタデータの削除
		if value == "" {
			header += "Remove-"
		}

		// コンテナの場合とオブジェクトの場合でヘッダ名が違う
		if isContainer {
			header += "Container-Meta-"
		} else {
			header += "Object-Meta-"
		}

		header += name
		req.Header.Add(header, value)

		log.Debugf("Set meta data: %s=%s", header, value)
	}

	// Read-ACLとWrite-ACL
	if isContainer && cmd.readAcl != "_no_assign_" {
		req.Header.Add("X-Container-Read", cmd.readAcl)
		log.Debugf("Set Read ACL: %s", cmd.readAcl)
	}

	if isContainer && cmd.writeAcl != "_no_assign_" {
		req.Header.Add("X-Container-Write", cmd.writeAcl)
		log.Debugf("Set Write ACL: %s", cmd.readAcl)
	}
}

// コンテナやオブジェクトにPOSTリクエストを送信する
func (cmd *Post) request(method string, item Item) (err error) {

	if method != "POST" && method != "PUT" {
		return errors.New(`Argument method should be "POST" or "PUT".`)
	}

	// リクエストを準備
	u, err := cmd.storageUrl(item)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		method,
		u.String(),
		nil,
	)

	req.Header.Set("X-Auth-Token", cmd.config.Token)

	// HTTPヘッダをセットする
	cmd.addHeaders(req, item)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 404:
		return errors.New("Object was not found.")
	case resp.StatusCode >= 400:
		msg := fmt.Sprintf("Return %d status code from the server with message. [%s].",
			resp.StatusCode,
			extractErrorMessage(resp.Body),
		)
		return errors.New(msg)
	}

	return nil
}
