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
	if err != nil {
		return exitCode, err
	}

	if exitCode == ExitCodeUsage {
		cmd.Usage()
		return exitCode, nil
	}

	err = cmd.DownloadObjects(cmd.objectName)
	if err == nil {
		return ExitCodeOK, nil
	} else {
		return ExitCodeError, err
	}
}

func (cmd *Download) DownloadObjects(path string) error {
	log := lib.GetLogInstance()

	// pathの末尾にワイルドカードがある場合はそれを処理
	if strings.HasSuffix(path, "*") {
		container := path[0 : len(path)-1]

		// オブジェクトの一覧を取得
		l := &List{Command: cmd.Command}
		list, err := l.List(container)
		if err != nil {
			return err
		}

		for i := 0; i < len(list); i++ {
			u, err := buildStorageUrl(cmd.config.EndPointUrl, container, list[i])
			if err != nil {
				return err
			}

			err = cmd.downloadObject(u)
			if err != nil {
				return err
			}

			log.Infof("%s download complete.", list[i])
		}

	} else {
		u, err := buildStorageUrl(cmd.config.EndPointUrl, path)
		if err != nil {
			return err
		}

		err = cmd.downloadObject(u)
		if err != nil {
			return err
		}
		log.Infof("%s download complete.", path)
	}

	return nil
}

func (cmd *Download) downloadObject(u *url.URL) error {

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
		return errors.New("Return error code from Server.")
	}

	// オブジェクト名と同じファイルをローカルに作成してBodyを書き込む
	reader := bufio.NewReader(resp.Body)
	basename := filepath.Base(u.Path)

	var filename string
	if cmd.destPath != "" {
		filename, err = cmd.resolveLocalPath(cmd.destPath, basename)
	} else {
		filename, err = cmd.resolveLocalPath(basename)
	}
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = io.Copy(writer, reader)
	if err != nil {
		return err
	}
	writer.Flush()

	return nil

}
