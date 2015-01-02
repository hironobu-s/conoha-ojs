package lib

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

// テスト用のJSONデータ
// 全部ダミーです
var sampleConfigJson = `{
  "Command": 0,
  "Token": "owe9f9jh93fe98u23jfi09824f9283f0",
  "TokenExpires": "Sat, 01 Jan 2015 08:18:34 UTC",
  "ApiUsername": "1111111",
  "ApiPassword": "testpassword",
  "TenantId": "oijewofijwoefinow0923f09jw30f9j0",
  "TenantId": "470710ce0ae24060886720fe4e7cf210",
  "TenantName": "1111111",
  "EndPointUrl": "https://objectstore-r1nd1001.cnode.jp/v1/oijewofijwoefinow0923f09jw30f9j0"
}`

func TestNewConfig(t *testing.T) {
	c := NewConfig()
	if c == nil {
		t.Errorf("config should not be nil")
	}
}

func TestConfigFilePath(t *testing.T) {
	c := NewConfig()

	path := c.ConfigFilePath()

	dir, file := filepath.Split(path)

	if fi, err := os.Stat(dir); fi == nil || err != nil {
		t.Errorf("wrong directory")
	}

	if file != CONFIGFILE {
		t.Errorf("file should be ConfigFile")
	}
}

func TestRead(t *testing.T) {
	c := NewConfig()

	dir := os.TempDir()
	file := dir + ".conoha-ojs"

	if err := ioutil.WriteFile(file, []byte(sampleConfigJson), 0775); err != nil {
		t.Error(err)
	}

	err := c.Read(file)
	if err != nil {
		t.Error(err)
	}
}

func TestWrite(t *testing.T) {
	c := NewConfig()

	dir := os.TempDir()
	file := dir + ".conoha-ojs"

	err := c.Save(file)
	if err != nil {
		t.Error(err)
	}
}
