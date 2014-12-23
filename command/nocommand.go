package command

import (
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	"io"
)

type Nocommand struct {
	*Command
}

func NewNocommand(stdSteram io.Writer, errStream io.Writer) (cmd *Nocommand) {
	cmd = &Nocommand{
		Command: NewCommand(stdSteram, errStream),
	}
	return cmd
}

// コマンドライン引数を処理する
func (cmd *Nocommand) parseFlags() error {
	return nil
}

// コマンドを実行して結果を出力する
// 実行ステータスを数値で返す
func (cmd *Nocommand) Run(c *lib.Config) (exitCode int, err error) {
	cmd.Usage()
	return ExitCodeError, nil
}

func (cmd *Nocommand) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s COMMAND [OPTIONS]

A CLI-tool for ConoHa Object Storage.

Commands: 
  auth      Authenticate a user.
  list      List a container or objects within a container.
  upload    Upload files or directories to a container.
  download  Download objects from a container.

`, COMMAND_NAME)
}
