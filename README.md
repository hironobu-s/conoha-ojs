# CLI-tool for ConoHa Object Storage.

[![Build Status](https://travis-ci.org/hironobu-s/conoha-ojs.svg?branch=travis-test)](https://travis-ci.org/hironobu-s/conoha-ojs)

[ConoHaオブジェクトストレージ](https://www.conoha.jp/)を操作するためのCLIツールです。ConoHaオブジェクトストレージで使える機能を一通りサポートしています。

## API

オブジェクトストレージのAPIを使用するので、あらかじめ[ConoHaのコントロールパネル](https://cp.conoha.jp/)から、APIユーザを作成しておいてください。

## 特徴

1. ConoHaオブジェクトストレージで使える機能を一通りカバーしています
2. Goで実装されているため、実行ファイルは一つでインストールが容易です
3. 2.と同じ理由で、Windows, MacOSX, Linuxなど、だいたいどの環境でも動作すると思います
4. 認証情報をファイルに保持します。コマンド実行の度に認証情報を設定する必要がありません
5. 他のOpenStack Swiftで構築されたシステムに対しても動作するかもしれません

## インストール

### MacOSX

ターミナルなどから以下のコマンドを実行します。

```bash
L=/usr/local/bin/conoha-ojs && curl -sL https://github.com/hironobu-s/conoha-ojs/releases/download/v20150217.1/conoha-ojs-osx.amd64.gz | zcat > $L && chmod +x $L
```

アンインストールする場合は/usr/local/bin/conoha-ojsを削除してください。

### Linux

ターミナルなどから以下のコマンドを実行します。/usr/local/binにインストールされるので、root権限が必要です。他のディレクトリにインストールする場合はL=/usr/local/bin/conoha-ojsの部分を適宜書き換えてください。

```bash
L=/usr/local/bin/conoha-ojs && curl -sL https://github.com/hironobu-s/conoha-ojs/releases/download/v20150217.1/conoha-ojs-linux.amd64.gz | zcat > $L && chmod +x $L
```

アンインストールする場合は/usr/local/bin/conoha-ojsを削除してください。

### Windows

[ZIPファイル](https://github.com/hironobu-s/conoha-ojs/releases/download/v20150217.1/conoha-ojs.amd64.zip)をダウンロードして、適当なフォルダに展開します。

アンインストールする場合はファイルをゴミ箱に入れてください。

## ビルド方法

自分でビルドする場合は以下のようにします。Goの実行環境が用意されていることが前提になります。

```bash
$ git clone https://github.com/hironobu-s/conoha-ojs
$ cd conoha-ojs
$ go build -o conoha-ojs
```

## 使い方

コマンド名(conoha-ojs)に続き、サブコマンドを指定します。

```bash
Usage: conoha-ojs COMMAND [OPTIONS]

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
```

## auth 

オブジェクトストレージの認証を行います。認証に成功した場合、認証情報がファイルに保存されます。

> **NOTE:** 認証は一番最初に行わなくてはなりません。認証を行わないと、他のサブコマンドが使用できません。

> **NOTE:** APIユーザ名とパスワードなどをファイルに保持します。ファイルはホームディレクトリの.conoha-ojsで、パーミッションは0600です。

```bash
$ conoha-ojs auth -u "api-username" -p "******" -t "tenant-id"
```


## list

コンテナ内のオブジェクト一覧を取得します。コンテナを省略した場合、一番上位のコンテナが選択されます。

```bash
$ conoha-ojs list <container-name>
```


## stat

コンテナ/オブジェクトのメタデータや詳細情報を取得します。

```bash
$ conoha-ojs stat <container or object>
```


## upload

ファイルをアップロードします。

一つ目の引数はアップロード先のコンテナです。二つ目以降の引数にアップロードするファイルを複数指定できます。

```bash
$ conoha-ojs upload <container> <file>...
```

ワイルドカードも指定できます
```bash
$ conoha-ojs upload <container> *.txt
```

## download

コンテナ/オブジェクトをダウンロードします。

一つ目の引数にダウンロードするオブジェクト。二つ目の引数にダウンロード先のパスを指定します。
第一引数にコンテナを指定すると、コンテナ内のオブジェクトをすべてダウンロードして、コンテナと同じ名前のディレクトリに格納されます。

```bash
$ conoha-ojs download <object> <dest path>
```

## delete

コンテナ/オブジェクトを削除します。コンテナを指定した場合、コンテナ内のオブジェクトもすべて削除されます。

```bash
$ conoha-ojs delete <container or object> 
```

## post 

コンテナ/オブジェクトにメタデータや、コンテナに対する読み込み権限(Read ACL), 書き込み権限(Write ACL)を設定します。また、空のコンテナを作成するのにも使用します。

メタデータを指定する場合は-mオプションを使用します。メタデータはキーと値を:(コロン)で区切ります。
```bash
$ conoha-ojs post -m foo:bar <container or object> 
```

値が指定されなかった場合は、メタデータを削除します。
```bash
$ conoha-ojs post -m foo: <container or object> 
```

コンテナに対する読み込み権限を設定するには-rオプションを使います。たとえば、コンテナをWeb公開する場合は以下のようになります。
```bash
$ conoha-ojs post -r ".r:*,.rlistings" <container>
```

コンテナに対する書き込み権限を設定するには-wオプションを使います。ただしConoHaオブジェクトストレージはユーザを一つしか作成できません。このオプションは機能しますが、今のところ意味がありません。
```bash
$ conoha-ojs post -w "account1 account2" <container>
```

## deauth 

conoha-ojsが作成した設定ファイルを削除します。設定ファイルにはオブジェクトストレージの認証情報が記録されていますが、必要に応じてこのサブコマンドで削除することができます。再びconoha-ojsを使う場合は、authサブコマンドを使って認証を行ってください。

```bash
$ conoha-ojs deauth
```

## version

バージョンを表示します。

```bash
$ conoha-ojs version
```

# TODO

* ~~バイナリを準備する~~
* 認証情報は環境変数に保存するようにしたい
* ラージオブジェクト対応
* 多数のダウンロード/アップロードは並列処理できる？
* テストが足りない
* 英語が間違ってるかも

# License

MIT License
