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

	// 実行するコマンド
	var cmd command.Commander

	// 出力先
	stdStream := os.Stdout
	errStream := os.Stderr

	// 設定を読み込む
	config := lib.NewConfig()

	// コマンドを実行
	if len(os.Args) <= 1 {
		// コマンドが未指定の場合は使用法を表示
		cmd = command.NewCommand("nocommand", config, stdStream, errStream)
		return cmd.Run()
	}

	command_name := os.Args[1]

	log.Debugf("Command: \"%v\"", command_name)

	if command_name == "auth" {
		// 認証情報を更新(コマンドライン引数から認証情報を設定する)
		auth := command.NewCommand("auth", config, stdStream, errStream)
		exitCode, err = auth.Run()
		if err != nil {
			return exitCode, err
		}

	} else if command_name == "deauth" {
		// 認証情報を削除
		auth := command.NewCommand("deauth", config, stdStream, errStream)
		exitCode, err = auth.Run()
		if err != nil {
			return exitCode, err
		}

	} else {
		// 認証情報を更新
		auth := command.NewCommand("auth", config, stdStream, errStream).(*command.Auth)
		if err = auth.CheckTokenIsExpired(config); err != nil {
			return 1, err
		}

		// コマンドを実行
		cmd := command.NewCommand(command_name, config, stdStream, errStream)

		code, err := cmd.Run()
		if err != nil {
			return code, err
		}
	}

	return command.ExitCodeOK, nil
}
