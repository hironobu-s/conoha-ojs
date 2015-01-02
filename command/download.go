package command

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	flag "github.com/ogier/pflag"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type Download struct {
	objectName string
	destPath   string

	*Command
}

func (cmd *Download) parseFlags() (exitCode int, err error) {

	var showUsage bool

	fs := flag.NewFlagSet("conoha-ojs-download", flag.ContinueOnError)
	fs.BoolVarP(&showUsage, "help", "h", false, "Print usage.")

	err = fs.Parse(os.Args[2:])
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	if showUsage {
		return ExitCodeUsage, nil
	}

	if len(os.Args) < 3 {
		return ExitCodeParseFlagError, errors.New("Not enough arguments.")
	}

	// 取得するオブジェクト名
	cmd.objectName = os.Args[2]

	// 保存先のパス
	if len(os.Args) == 4 {
		cmd.destPath = os.Args[3]

	} else {
		cmd.destPath = "."
	}

	return ExitCodeOK, nil
}

func (cmd *Download) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s download <object_name> <dest_path>

Download objects from a container.

<object_name> Name of object to download.
<dest_path>   (optional) Name of destination path. Default is current directory.

`, lib.COMMAND_NAME)
}

func (cmd *Download) Run() (exitCode int, err error) {

	exitCode, err = cmd.parseFlags()
	if err != nil || exitCode == ExitCodeUsage {
		cmd.Usage()
		return exitCode, err
	}

	err = cmd.DownloadObjects(cmd.objectName, cmd.destPath)
	if err == nil {
		return ExitCodeOK, nil
	} else {
		return ExitCodeError, err
	}
}

func (cmd *Download) DownloadObjects(srcpath string, destpath string) error {
	log := lib.GetLogInstance()

	// 対象の情報を取得
	s := NewCommand("stat", cmd.config, cmd.stdStream, cmd.errStream).(*Stat)
	item, err := s.Stat(srcpath)
	if err != nil {
		return err
	}

	_, isContainer := item.(*Container)

	if isContainer {
		// オブジェクトの一覧を取得
		l := NewCommand("list", cmd.config, cmd.stdStream, cmd.errStream).(*List)
		list, err := l.List(srcpath)
		if err != nil {
			return err
		}

		for i := 0; i < len(list); i++ {
			cmd.DownloadObjects(srcpath+"/"+list[i], destpath)
		}

	} else {

		log.Debugf("Downloading %s => %s", srcpath, destpath)

		u, err := buildStorageUrl(cmd.config.EndPointUrl, srcpath)
		if err != nil {
			return err
		}

		err = cmd.request(u, destpath)
		if err != nil {
			log.Infof("%s download error.", srcpath)
			return err
		}
		log.Infof("%s download complete.", srcpath)
	}

	return nil
}

func (cmd *Download) request(u *url.URL, destpath string) error {

	req, err := http.NewRequest(
		"GET",
		u.String(),
		nil,
	)
	if err != nil {
		return err
	}
	req.Header.Set("X-Auth-Token", cmd.config.Token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// HTTPステータスコードがエラーを返した場合
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

	// オブジェクト名と同じファイルをローカルに作成してBodyを書き込む
	_, err = cmd.store(resp.Body, u, destpath)
	if err != nil {
		return err
	}

	return nil
}

// オブジェクトをファイルに保存する
// 保存したサイズを返す
func (cmd *Download) store(body io.ReadCloser, u *url.URL, destpath string) (written int64, err error) {

	rawurl := u.String()

	// オブジェクトを保存するローカルファイル名を決める
	if !strings.Contains(rawurl, cmd.config.EndPointUrl) {
		return -1, errors.New("Object URL dose not contain the EndPoint URL.")
	}

	// オブジェクトのURLからEndPointUrlの部分を削除して、基準のパスとする
	path := strings.Replace(rawurl, cmd.config.EndPointUrl, "", 1)

	// 保存先が引数で指定されている場合、そのパスを使う
	path = destpath + string(filepath.Separator) + path

	// パスを正規化する
	path = filepath.Clean(path)

	// パスとファイル名に分離
	dir, _ := filepath.Split(path)

	// ディレクトリが存在しない場合は作成する
	_, err = os.Stat(dir)
	if err != nil {
		// 0777 で作成しているがumaskが考慮されるため実際は0755などになる
		err = os.MkdirAll(dir, 0777)
	}

	file, err := os.Create(path)
	if err != nil {
		return -1, err
	}
	defer file.Close()

	// オブジェクトを保存
	reader := bufio.NewReader(body)
	writer := bufio.NewWriter(file)
	written, err = io.Copy(writer, reader)
	if err != nil {
		return -1, err
	}
	writer.Flush()

	return written, nil
}
