package command

import (
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	"os"
)

type Deauth struct {
	*Command
}

// コマンドライン引数を処理する
func (cmd *Deauth) parseFlags() (exitCode int, err error) {
	return ExitCodeOK, nil
}

// コマンドを実行して結果を出力する
// 実行ステータスを数値で返す
func (cmd *Deauth) Run() (exitCode int, err error) {
	exitCode, err = cmd.parseFlags()
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	path := cmd.config.ConfigFilePath()

	fi, _ := os.Stat(path)
	if fi != nil {
		err = os.Remove(path)
		if err != nil {
			return ExitCodeError, err
		}
	}

	return exitCode, nil
}

func (cmd *Deauth) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s deauth 

Remove an authentication file (~/.conoha-ojs) from local machine.

`, lib.COMMAND_NAME)
}
