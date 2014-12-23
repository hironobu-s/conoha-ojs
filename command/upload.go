package command

import (
	"errors"
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	"net/http"
	"os"
)

type Upload struct {
	srcFiles      []string
	destContainer string

	*Command
}

func NewUpload() *Upload {
	cmd := new(Upload)
	return cmd
}

func (cmd *Upload) parseFlags() error {

	if len(os.Args) < 4 {
		msg := fmt.Sprintf(`Usage: %s upload <container> <file or directory>...

Upload files or directories to a container.

<container>          Name of container to upload.
<file or directory>  Name of file or directory to upload.
`, COMMAND_NAME)
		return errors.New(msg)
	}

	// アップロード先コンテナ
	cmd.destContainer = os.Args[2]

	// アップロードするファイル
	for i := 3; i < len(os.Args); i++ {
		filename := os.Args[i]

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

	// PUTリクエストを実行
	req, err := http.NewRequest("PUT", uri.String(), file)
	if err != nil {
		return err
	}
	req.Header.Set("Content-type", "text/plain")
	req.Header.Set("X-Auth-Token", c.Token)

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
	log.Infof("%s uploaded.", filename)
	return nil
}
