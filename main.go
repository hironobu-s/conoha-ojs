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
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func run() error {
	// log
	log := lib.GetLogInstance()

	// エラー
	var err error

	// 設定を読み込む
	c := lib.NewConfig()

	// コマンドを実行
	if len(os.Args) <= 1 {
		// コマンドが未指定の場合は使用法を表示
		n := command.NewNocommand()
		return n.Run(c)
	}

	command_name := os.Args[1]

	log.Infof("Command: \"%v\"", command_name)

	if command_name == "auth" {
		// 認証情報を更新(コマンドライン引数から認証情報を設定する)
		auth := command.NewAuth()
		if err = auth.Run(c); err != nil {
			return err
		}

	} else {
		// 認証情報を更新
		auth := command.NewAuth()
		if err = auth.CheckTokenIsExpired(c); err != nil {
			return err
		}

		var cmd command.Commander
		switch command_name {

		case "list":
			cmd = command.NewList()

		case "download":
			cmd = command.NewDownload()

		case "upload":
			cmd = command.NewUpload()

		default:
			// 定義されてないコマンド
			cmd = command.NewNocommand()
		}

		err := cmd.Run(c)
		if err != nil {
			return err
		}
	}

	// アカウント情報を書き出す
	err = c.Save()
	if err != nil {
		return err
	}

	return nil
}
