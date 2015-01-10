package command

import (
	"fmt"
	"../lib"
)

type Nocommand struct {
	*Command
}

// コマンドライン引数を処理する
func (cmd *Nocommand) parseFlags() (exitCode int, err error) {
	return ExitCodeOK, nil
}

// コマンドを実行して結果を出力する
// 実行ステータスを数値で返す
func (cmd *Nocommand) Run() (exitCode int, err error) {
	exitCode, err = cmd.parseFlags()

	cmd.Usage()
	return exitCode, err
}

func (cmd *Nocommand) Usage() {
	fmt.Fprintf(cmd.errStream, `Usage: %s COMMAND [OPTIONS]

A CLI-tool for ConoHa Object Storage.

Commands: 
  auth      Authenticate a user.
  list      List a container or objects within a container.
  stat      Show informations for container or object.
  upload    Upload files or directories to a container.
  download  Download objects from a container.
  delete    Delete a container or objects within a container.
  post      Update meta datas for the container or objects;
            create containers if not present.
  deauth    Remove an authentication file (~/.conoha-ojs) from a local machine.
  version   Print version.

`, lib.COMMAND_NAME)
}
