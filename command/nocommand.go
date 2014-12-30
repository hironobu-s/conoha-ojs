package command

import (
	"fmt"
	"github.com/hironobu-s/conoha-ojs/lib"
	flag "github.com/ogier/pflag"
	"os"
)

type Nocommand struct {
	*Command
}

// コマンドライン引数を処理する
func (cmd *Nocommand) parseFlags() (exitCode int, err error) {
	var showVersion bool
	fs := flag.NewFlagSet("conoha-ojs-nocommand", flag.ContinueOnError)
	fs.BoolVar(&showVersion, "version", false, "Print version.")

	err = fs.Parse(os.Args[1:])
	if err != nil {
		return ExitCodeParseFlagError, err
	}

	if showVersion {
		return ExitCodeUsage, nil
	}
	return ExitCodeOK, nil
}

func (cmd *Nocommand) Version() {
	fmt.Fprintf(cmd.errStream, "Version: %s\n", lib.VERSION)
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
  upload    Upload files or directories to a container.
  download  Download objects from a container.
  delete    Delete a container or objects within a container.
  post      Update meta datas for the container.

`, lib.COMMAND_NAME)
}
