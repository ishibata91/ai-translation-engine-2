## 1. controller API テスト基盤の配置確定

- [x] 1.1 `pkg/controller/**` の既存テストを棚卸しし、共通化対象のセットアップを洗い出す
- [x] 1.2 `testenv` と helper の配置を `pkg/tests/api_tests` に確定し、責務を共通基盤と controller 別 builder に分離する
- [x] 1.3 `standard_test_spec.md` に整合する controller API テストの標準パターンを決める

## 2. 共通 testenv の実装

- [x] 2.1 `pkg/tests/api_tests` 配下に controller API テストで再利用する `testenv` と helper を追加する
- [x] 2.2 `stretchr/testify` を導入し、API テストの標準アサーション記法を揃える
- [x] 2.3 `tmp/api_test_db/` 配下にテスト用 SQLite DB を生成できるようにし、通常の `db/` と分離する
- [x] 2.4 共通 `testenv` は DB、logger、trace 付き `context.Context`、共通 utility に限定し、controller 固有依存は builder 側へ分離する

## 3. 既存 controller テストの API テスト化

- [x] 3.1 `pkg/controller/**` の既存テストを `testenv` 利用へ寄せる
- [x] 3.2 正常系・異常系を table-driven で表現できる箇所から整理する
- [x] 3.3 `context.Context` 伝播と error wrap の確認が継続して検証できるようにする
- [x] 3.4 controller ごとの差分依存は controller 別 builder または env 作成関数で吸収する

## 4. 品質ゲート確認

- [x] 4.1 `.gitignore` にテスト用 SQLite DB の保存先を追加し、生成物が追跡されないことを確認する
- [x] 4.2 変更中の Go ファイルに対して `npm run backend:lint:file -- <file...>` を逐次実行して違反を解消する
- [x] 4.3 `npm run lint:backend` を実行し、backend lint 全体が通ることを確認する
- [x] 4.4 `go test ./pkg/...` と `npm run backend:test` または `npm run backend:check` を実行し、controller API テストが既存導線で通ることを確認する
