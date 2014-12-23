package command

import (
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	"os"
)

type Nocommand struct {
}

// コマンドライン引数を処理する
func (cmd *Nocommand) parseFlags() error {
	return nil
}

func NewNocommand() *Nocommand {
	cmd := new(Nocommand)
	return cmd
}

// コマンドを実行して結果を出力する
// 実行ステータスを数値で返す
func (cmd *Nocommand) Run(cfg *lib.Config) error {
	cmd.Usage()
	return nil
}

func (cmd *Nocommand) Usage() {
	fmt.Fprintf(os.Stderr, `Usage: %s COMMAND [OPTIONS]

A CLI-tool for ConoHa Object Storage.

Commands: 
  auth      Authenticate a user.
  list      List a container or objects within a container.
  upload    Upload files or directories to a container.
  download  Download objects from a container.

`, COMMAND_NAME)
}
