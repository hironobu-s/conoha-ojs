package command

import (
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	flag "github.com/ogier/pflag"
)

type Upload struct {
	srcFiles      []string
	destContainer string

	contentType        string
	defaultContentType string

	*Command
}

func NewUpload() *Upload {
	cmd := new(Upload)

	// 規定のContent-typeを設定
	cmd.defaultContentType = "application/octet-stream"

	return cmd
}

func (cmd *Upload) parseFlags() error {

	// コマンドライン引数の定義を追加
	flag.StringVarP(&cmd.contentType, "content-type", "c", "", "Set Content-type")
	os.Args = os.Args[1:]
	flag.Parse()

	if flag.NArg() < 2 {
		msg := fmt.Sprintf(`Usage: %s upload <container> <file or directory>...

Upload files or directories to a container.

<container>          Name of container to upload.
<file or directory>  Name of file or directory to upload.

  -c, --content-type: Set Content-type. If not set, Content-type will be "application/octet-strem".

`, COMMAND_NAME)
		return errors.New(msg)
	}

	// アップロード先コンテナ
	cmd.destContainer = flag.Arg(0)

	// アップロードするファイル
	for i := 1; i < flag.NArg(); i++ {
		filename := flag.Arg(i)

		_, err := os.Stat(filename)
		if err != nil {
			msg := fmt.Sprintf("File \"%s\" not found.", filename)
			return errors.New(msg)
		}

		cmd.srcFiles = append(cmd.srcFiles, filename)
	}

	return nil
}

func (cmd *Upload) Run(c *lib.Config) error {
	// コマンドライン引数を処理
	err := cmd.parseFlags()
	if err != nil {
		return err
	}

	for _, filename := range cmd.srcFiles {
		cmd.uploadObject(c, filename)
	}

	return nil
}

// Content-typeを決定する
func (cmd *Upload) detectContentType(filename string) (contentType string) {

	// 指定がある場合はそれを使う
	if cmd.contentType != "" {
		return cmd.contentType
	}

	ext := filepath.Ext(filename)

	contentType = mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = cmd.defaultContentType
	}
	return contentType
}

func (cmd *Upload) uploadObject(c *lib.Config, filename string) (err error) {

	// アップロードするファイルへのReaderを作成
	file, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}

	// アップロード先のURIを準備
	uri, err := buildStorageUrl(c.EndPointUrl, cmd.destContainer, filename)
	if err != nil {
		return err
	}

	// PUTリクエストを作成
	req, err := http.NewRequest("PUT", uri.String(), file)
	if err != nil {
		return err
	}

	contentType := cmd.detectContentType(filename)
	req.Header.Set("Content-type", contentType)
	req.Header.Set("X-Auth-Token", c.Token)

	// リクエストを実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	switch {
	case resp.StatusCode == 404:
		return errors.New("Container was not found.")
	}

	log := lib.GetLogInstance()
	log.Infof("%s (%s) was uploaded.", filename, contentType)

	return nil
}
