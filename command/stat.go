package command

import (
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	flag "github.com/ogier/pflag"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Statで取得したオブジェクトのヘッダ情報を格納する構造体
type Item interface {
	fmt.Stringer
}

type Object struct {
	Object        string
	ContentType   string
	ContentLength uint64
	LastModified  time.Time
	ETag          string

	// メタデータ
	MetaDatas map[string]string

	// すべてのヘッダ情報
	header map[string][]string
}

func (item *Object) String() string {

	padding := 14
	for name, _ := range item.MetaDatas {
		if len(name) > padding {
			padding = len(name)
		}
	}

	format := "%" + strconv.Itoa(padding) + "s: "

	lines := []string{}

	lines = append(lines, fmt.Sprintf(format+"%s", "Object", item.Object))
	lines = append(lines, fmt.Sprintf(format+"%s", "Content Type", item.ContentType))
	lines = append(lines, fmt.Sprintf(format+"%d", "Content Length", item.ContentLength))
	lines = append(lines, fmt.Sprintf(format+"%s", "LastModified", item.LastModified.Format(time.RFC1123)))
	lines = append(lines, fmt.Sprintf(format+"%s", "ETag", item.ETag))

	for name, value := range item.MetaDatas {
		lines = append(lines, fmt.Sprintf(format+"%s", name, value))
	}
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

type Container struct {
	Container string
	Objects   uint64
	Bytes     uint64
	ReadAcl   string
	WriteAcl  string
	SyncTo    string // ConoHa未対応のため未実装
	SyncKey   string // ConoHa未対応のため未実装

	// メタデータ
	MetaDatas map[string]string

	// すべてのヘッダ情報
	header map[string][]string
}

func (item *Container) String() string {

	padding := 10
	for name, _ := range item.MetaDatas {
		if len(name) > padding {
			padding = len(name)
		}
	}

	format := "%" + strconv.Itoa(padding) + "s: "

	lines := []string{}

	lines = append(lines, fmt.Sprintf(format+"%s", "Container", item.Container))
	lines = append(lines, fmt.Sprintf(format+"%d", "Objects", item.Objects))
	lines = append(lines, fmt.Sprintf(format+"%d", "Bytes", item.Bytes))
	lines = append(lines, fmt.Sprintf(format+"%s", "Read ACL", item.ReadAcl))
	lines = append(lines, fmt.Sprintf(format+"%s", "Write ACL", item.WriteAcl))

	for name, value := range item.MetaDatas {
		lines = append(lines, fmt.Sprintf(format+"%s", name, value))
	}
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

// ------------------------------------------------------------------------------

type Stat struct {
	objectName string

	*Command
}

func (cmd *Stat) parseFlags() (exitCode int, err error) {
	var showUsage bool

	fs := flag.NewFlagSet("conoha-ojs-post", flag.ContinueOnError)

	fs.BoolVarP(&showUsage, "help", "h", false, "Print usage.")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	if showUsage {
		return ExitCodeUsage, nil
	}

	cmd.objectName = fs.Arg(0)
	if cmd.objectName == "" {
		// オブジェクトが指定されなかった場合、一番上位のコンテナが指定されるのでそのまま進めて良い
	}
	return ExitCodeOK, nil
}

func (cmd *Stat) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s <container or object>

Show informations for container or object.

<container or object>  Name of container or object to post to.

`, lib.COMMAND_NAME)
}

func (cmd *Stat) Run() (exitCode int, err error) {
	exitCode, err = cmd.parseFlags()
	if err != nil || exitCode == ExitCodeUsage {
		cmd.Usage()
		return exitCode, err
	}

	item, err := cmd.Stat(cmd.objectName)
	if err != nil {
		return ExitCodeError, err
	}

	// 詳細を出力
	fmt.Fprintf(cmd.stdStream, item.String())

	return ExitCodeOK, nil
}

func (cmd *Stat) Stat(path string) (item Item, err error) {

	headers, err := cmd.request(path)
	if err != nil {
		return nil, err
	}

	return cmd.populate(path, headers)
}

// ヘッダー情報からItem構造体を構築する
func (cmd *Stat) populate(path string, headers map[string][]string) (item Item, err error) {

	var convert64 = func(strval string) (intval uint64, err error) {
		intval, err = strconv.ParseUint(strval, 10, 64)

		if err != nil {
			msg := fmt.Sprintf("Can't convert to int value. [%s]", strval)
			return 0, errors.New(msg)
		}

		return intval, nil
	}

	// コンテナかオブジェクトかを判断する良い方法が無さげなので、
	// X-Container-Object-Count と X-Container-Bytes-Used があればコンテナと判断する
	_, exists1 := headers["X-Container-Object-Count"]
	_, exists2 := headers["X-Container-Bytes-Used"]
	if exists1 && exists2 {
		item := &Container{
			Container: path,
			MetaDatas: map[string]string{},
		}

		for name, value := range headers {
			switch name {
			case "X-Container-Bytes-Used":
				item.Bytes, err = convert64(value[0])
			case "X-Container-Object-Count":
				item.Objects, err = convert64(value[0])
			case "X-Container-Read":
				item.ReadAcl = value[0]
			case "X-Container-Write":
				item.WriteAcl = value[0]
			default:
				item.MetaDatas[name] = value[0]
			}

			if err != nil {
				return nil, err
			}
		}
		return item, nil

	} else {
		item := &Object{
			Object:    path,
			MetaDatas: map[string]string{},
		}

		for name, value := range headers {
			switch name {
			case "Content-Type":
				item.ContentType = value[0]
			case "Content-Length":
				item.ContentLength, err = convert64(value[0])
			case "Etag":
				item.ETag = value[0]
			case "Last-Modified":
				d, err := time.Parse(time.RFC1123, value[0])
				if err != nil {
					return nil, err
				}

				item.LastModified = d
			default:
				item.MetaDatas[name] = value[0]
			}
		}
		return item, nil
	}
}

// オブジェクトのヘッダー情報を取得する
func (cmd *Stat) request(path string) (headers map[string][]string, err error) {
	u, err := buildStorageUrl(cmd.config.EndPointUrl, path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"HEAD",
		u.String(),
		nil,
	)

	req.Header.Set("X-Auth-Token", cmd.config.Token)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == 404:
		return nil, errors.New("Object was not found.")
	case resp.StatusCode >= 400:
		msg := fmt.Sprintf("Return %d status code from the server with message. [%s].",
			resp.StatusCode,
			extractErrorMessage(resp.Body),
		)
		return nil, errors.New(msg)
	}

	headers = map[string][]string{}
	for name, value := range resp.Header {
		headers[name] = value
	}

	return headers, nil
}
