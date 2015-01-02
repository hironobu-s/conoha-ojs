package lib

import (
	"github.com/Sirupsen/logrus"
	"os"
)

func init() {
}

var instance *logrus.Logger

func GetLogInstance() *logrus.Logger {
	if instance == nil {

		// テキストフォーマットで出力
		logrus.SetFormatter(&logrus.TextFormatter{})
		//logrus.SetFormatter(&logrus.JSONFormatter{})

		// 標準エラー出力に出す
		logrus.SetOutput(os.Stderr)

		// ログレベルの設定
		instance = logrus.New()
		//instance.Level = logrus.DebugLevel
		instance.Level = logrus.InfoLevel
		//instance.SetOutput(os.Stderr)
	}
	return instance
}
