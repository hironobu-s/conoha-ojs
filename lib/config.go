package lib

import (
	"encoding/json"
	"os"
	"os/user"

	log "github.com/Sirupsen/logrus"
)

var Version string

const (
	CONFIGFILE   = ".conoha-ojs"
	COMMAND_NAME = "conoha-ojs"
)

// コンフィグ
type Config struct {
	// 実行コマンド
	Command int

	// API認証トークンと有効期限
	// TokenExpiresは文字列でなくtime型で保存したいが
	// JSONのEncode/Decodeで正しく動作しない
	Token        string
	TokenExpires string

	// 認証情報
	ApiUsername string
	ApiPassword string
	TenantId    string
	EndPointUrl string
}

func init() {
	// ログの初期化
	// https://github.com/Sirupsen/logrus

	// テキストフォーマットで出力
	log.SetFormatter(&log.TextFormatter{})

	// 標準エラー出力に出す
	log.SetOutput(os.Stderr)

	// ログレベルの設定
	log.SetLevel(log.InfoLevel)
}

func NewConfig() *Config {
	config := new(Config)

	// アカウント情報を読み込む
	path := config.ConfigFilePath()
	err := config.Read(path)
	if err != nil {
		// コンフィグファイルが読めなくてもwriteConfigFile()で上書きされるので無視して良い。
	}

	return config
}

// アカウント情報ファイルのパスを返す
// 基本的には、ホームディレクトリの.conoha-ojsというファイルになる
func (c *Config) ConfigFilePath() string {
	u, err := user.Current()
	if err == nil {
		return u.HomeDir + string(os.PathSeparator) + CONFIGFILE
	} else {
		// ここに来ることはなさそうだが、その場合はカレントディレクトリを決め打ちする
		return ".conoha-ojs"
	}
}

// 設定ファイル(~/.conoha-ojs)を読み込む
func (c *Config) Read(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		// 失敗した場合はあきらめる
		log.Warnln("Cannot read the config file.")
		return err
	}

	return nil
}

// コンフィグをファイルに書き出す
func (c *Config) Save(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// パーミッションを0600に変更する
	os.Chmod(path, 0600)

	encoder := json.NewEncoder(file)
	err = encoder.Encode(c)
	if err != nil {
		return err
	}

	return nil
}
