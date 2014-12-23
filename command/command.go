package command

import (
	"github.com/hironobu-s/conoha-ojs/lib"
)

const (
	COMMAND_NAME = "conoha-ojs"
)

type Commander interface {
	// コマンドライン引数を処理してmapで返す
	parseFlags() error

	// コマンドを実行して結果を出力する
	// 実行ステータスを数値で返す
	Run(cfg *lib.Config) error
}

type Command struct {
	config *lib.Config
}
