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

func (cmd *Upload) parseFlags() (exitCode int, err error) {

	var showUsage bool

	fs := flag.NewFlagSet("conoha-ojs-upload", flag.ContinueOnError)

	// コマンドライン引数の定義を追加
	fs.BoolVarP(&showUsage, "help", "h", false, "Print usage.")
	fs.StringVarP(&cmd.contentType, "content-type", "c", "", "Set Content-type")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	if showUsage {
		return ExitCodeUsage, nil
	}

	if fs.NArg() < 2 {
		return ExitCodeParseFlagError, errors.New("Not enough arguments.")
	}

	// アップロード先コンテナ
	cmd.destContainer = fs.Arg(0)

	// アップロードするファイル
	for i := 1; i < fs.NArg(); i++ {
		filename := fs.Arg(i)

		_, err := os.Stat(filename)
		if err != nil {
			msg := fmt.Sprintf("File \"%s\" not found.", filename)
			return ExitCodeError, errors.New(msg)
		}

		cmd.srcFiles = append(cmd.srcFiles, filename)
	}

	return ExitCodeOK, nil
}

func (cmd *Upload) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s upload <container> <file or directory>...

Upload files or directories to a container.

<container>          Name of container to upload.
<file or directory>  Name of file or directory to upload.

  -c, --content-type: Set Content-type. If not set, Content-type will be "application/octet-strem".

`, lib.COMMAND_NAME)
}

func (cmd *Upload) Run() (exitCode int, err error) {
	exitCode, err = cmd.parseFlags()
	if err != nil {
		return exitCode, err
	}

	if exitCode == ExitCodeUsage {
		cmd.Usage()
		return exitCode, nil
	}

	for _, filename := range cmd.srcFiles {
		err = cmd.uploadObject(filename)
		if err != nil {
			return ExitCodeError, err
		}
	}

	return ExitCodeOK, nil
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

func (cmd *Upload) uploadObject(filename string) (err error) {

	// アップロードするファイルへのReaderを作成
	file, err := os.OpenFile(filename, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}

	// アップロード先のURIを準備
	uri, err := buildStorageUrl(cmd.config.EndPointUrl, cmd.destContainer, filename)
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
	req.Header.Set("X-Auth-Token", cmd.config.Token)

	// リクエストを実行
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	switch {
	case resp.StatusCode == 404:
		return errors.New("Container was not found.")

	case resp.StatusCode >= 400:
		msg := fmt.Sprintf("Return %d status code from the server with message. [%s].",
			resp.StatusCode,
			extractErrorMessage(resp.Body),
		)
		return errors.New(msg)
	}

	log := lib.GetLogInstance()
	log.Infof("%s (content-type: %s) was uploaded.", filename, contentType)

	return nil
}
