# qiitaexporter

Qiitaから記事をエクスポートするツールです。

## インストール

```sh
$ go get -u github.com/tenntenn/qiitaexporter
```

## アクセストークンの取得

[こちら](https://qiita.com/settings/applications)から取得できます。

## 使い方

以下のように、環境変数`QIITA`でアクセストークンを指定します。

```sh
$ QIITA=xxxx qiitaexporter
```

出力ファイルを変えたい場合は`-postdir`を指定します。
画像ファイルは`posts`以下に出力されますが、変えたい場合は`-imgdir`を指定します。
画像のリンクにプリフィックスをつけたい場合は、`-imgprefix`で指定できます。

デフォルトではHugo向けに出力しますが、`-template`オプションでテンプレートを変更できます。
`template.go`を参考にすると良いでしょう。
