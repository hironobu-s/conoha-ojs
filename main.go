package main

import (
	"fmt"
	"github.com/hironobu-s/conoha-ojs/command"
	"github.com/hironobu-s/conoha-ojs/lib"
	"os"
)

type commandNotFoundError struct {
	cmd string
}

func (e commandNotFoundError) Error() string {
	return fmt.Sprintf("%#v", e.cmd)
}

func main() {
	exitCode, err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
	}
	os.Exit(exitCode)
}

func run() (exitCode int, err error) {
	// log
	log := lib.GetLogInstance()

	// 出力先
	stdStream := os.Stdout
	errStream := os.Stderr

	// 設定を読み込む
	c := lib.NewConfig()

	// コマンドを実行
	if len(os.Args) <= 1 {
		// コマンドが未指定の場合は使用法を表示
		cmd := command.NewNocommand(stdStream, errStream)

		return cmd.Run(c)
	}

	command_name := os.Args[1]

	log.Infof("Command: \"%v\"", command_name)

	if command_name == "auth" {
		// 認証情報を更新(コマンドライン引数から認証情報を設定する)
		auth := command.NewAuth(stdStream, errStream)
		exitCode, err = auth.Run(c)
		if err != nil {
			return exitCode, err
		}

	} else {
		// 認証情報を更新
		auth := command.NewAuth(stdStream, errStream)
		if err = auth.CheckTokenIsExpired(c); err != nil {
			return 1, err
		}

		var cmd command.Commander
		switch command_name {

		case "list":
			cmd = command.NewList(stdStream, errStream)

		case "download":
			cmd = command.NewDownload(stdStream, errStream)

		case "upload":
			cmd = command.NewUpload(stdStream, errStream)

		default:
			// 定義されてないコマンド
			cmd = command.NewNocommand(stdStream, errStream)
		}

		code, err := cmd.Run(c)
		if err != nil {
			return code, err
		}
	}

	// アカウント情報を書き出す
	err = c.Save()
	if err != nil {
		return command.ExitCodeError, err
	}

	return command.ExitCodeOK, nil
}
