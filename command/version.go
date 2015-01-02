package command

import (
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
)

type Version struct {
	*Command
}

// コマンドライン引数を処理する
func (cmd *Version) parseFlags() (exitCode int, err error) {
	return ExitCodeUsage, nil
}

// コマンドを実行して結果を出力する
// 実行ステータスを数値で返す
func (cmd *Version) Run() (exitCode int, err error) {
	exitCode, err = cmd.parseFlags()

	cmd.Usage()
	return exitCode, err
}

func (cmd *Version) Usage() {
	fmt.Fprintf(cmd.errStream, "Version: %s\n", lib.VERSION)
}
