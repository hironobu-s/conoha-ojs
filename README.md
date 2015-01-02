# conoha-ojs - A CLI-tool for ConoHa Object Storage.

conoha-ojsは[ConoHaオブジェクトストレージ](https://www.conoha.jp/)を操作するためのCLIツールです。
オブジェクトストレージのAPIを使用するので、あらかじめコントロールパネルから、APIユーザを作成しておいてください。


## Features

ConoHaオブジェクトストレージはOpenStack Swiftをベースに構築されています。

OpenStack SwiftにはPython実装のCLIツールが存在します(swiftコマンド)。このツールはswiftの劣化クローンではありますが、機能をよく使うものに絞りシンプルに使えることを目的にしています。

このツールの特徴や、swiftコマンドとの違いは以下です。

1. ConoHaオブジェクトストレージで使える機能を一通りカバーしています。
2. Go言語で実装されているため、実行ファイルが一つでインストールが容易です
3. 認証情報をファイルに保持します。コマンド実行の度に認証情報を設定する必要がありません。
4. 他のOpenStack Swiftで構築されたシステムに対しても動作するかもしれません。


## Usage

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
  post      Update meta datas for the container.
  version   Print version.
```


## auth 

オブジェクトストレージの認証を行います。認証に成功した場合、認証情報がファイルに保存されます。

> **NOTE:** 認証は一番最初に行わなくてはなりません。認証を行わないと、他のサブコマンドが使用できません。

> **NOTE:** APIユーザ名とパスワードなどをファイルに保持します。ファイルはホームディレクトリの.conoha-ojsで、パーミッションは0600です。

```bash
$ conoha-ojs auth -u "api-username" -p "******"
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

一つ目の引数にアップロード先のコンテナです。二つ目以降の引数にアップロードするファイルを複数指定できます。

```bash
$ conoha-ojs upload <container> <file or directory>...
```

ワイルドカードも指定できます
```bash
$ conoha-ojs upload <container> *.txt
```

## download

コンテナ/オブジェクトをダウンロードします。

一つ目の引数にダウンロードするオブジェクト。二つ目の引数にダウンロード先のパスを指定します。
第一引数にコンテナをしていすると、コンテナ内のオブジェクトをすべてダウンロードして、コンテナと同じ名前のディレクトリに格納されます。

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

## version

バージョンを表示します。

```bash
$ conoha-ojs version
```


# TODO

* テストをもっと書く
* 英語が間違ってるかも


# License

MTI License
