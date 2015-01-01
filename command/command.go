package command

import (
	"github.com/hironobu-s/conoha-ojs/lib"
	"io"
)

const (
	ExitCodeOK = iota
	ExitCodeError
	ExitCodeParseFlagError // 引数解析に失敗
	ExitCodeUsage          // Usageを表示
)

type Commander interface {

	// コマンドライン引数を処理する
	parseFlags() (exitCode int, err error)

	// ヘルプを表示する
	Usage()

	// コマンドを実行して結果を出力する
	// 実行ステータスを数値で返す
	Run() (exitCode int, err error)
}

type Command struct {
	// 設定
	config *lib.Config

	// 出力先
	stdStream io.Writer
	errStream io.Writer
}

// コマンドを作成して返す
func NewCommand(action string, config *lib.Config, stdStream io.Writer, errStream io.Writer) (cmd Commander) {

	command := &Command{
		config:    config,
		stdStream: stdStream,
		errStream: errStream,
	}

	switch action {
	case "list":
		cmd = &List{Command: command}
	case "auth":
		cmd = &Auth{Command: command}
	case "download":
		cmd = &Download{Command: command}
	case "upload":
		cmd = &Upload{Command: command}
	case "stat":
		cmd = &Stat{Command: command}
	case "delete":
		cmd = &Delete{Command: command}
	default:
		cmd = &Nocommand{Command: command}
	}

	return cmd
}
