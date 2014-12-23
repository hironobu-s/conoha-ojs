package command

import (
	"github.com/hironobu-s/conoha-ojs/lib"
	"io"
	"os"
)

const (
	ExitCodeOK = iota
	ExitCodeError
	ExitCodeParseFlagError

	COMMAND_NAME = "conoha-ojs"
)

type Commander interface {

	// コマンドライン引数を処理してmapで返す
	parseFlags() error

	// ヘルプを表示する
	Usage()

	// コマンドを実行して結果を出力する
	// 実行ステータスを数値で返す
	Run(c *lib.Config) (exitCode int, err error)
}

type Command struct {
	// 出力先
	stdStream, errStream io.Writer
}

func NewCommand(stdSteram io.Writer, errStream io.Writer) (cmd *Command) {
	cmd = &Command{
		stdStream: os.Stdout,
		errStream: os.Stderr,
	}

	return cmd
}
