# md2http

カレントディレクトリ（または指定ディレクトリ）の Markdown ファイルを HTML に変換して返すだけの簡易 HTTP サーバー。Go 製。

## 使い方

```sh
go run . [-port 8080] [-dir .] [PORT]
```

- `-port` / 位置引数 / 環境変数 `PORT` でポート番号を指定
- `-dir` で配信するディレクトリを指定（デフォルトはカレントディレクトリ）

`.md` / `.markdown` は Markdown として HTML に変換（GFM 対応、テーブル・コードブロック等）。それ以外のファイルは静的配信。ディレクトリは一覧表示。

## License

GPLv3 (or later). See [LICENSE](LICENSE). Third-party notices in [NOTICE](NOTICE).

```sh
go build -o md2http .
./md2http 8080
```
EOF
echo done