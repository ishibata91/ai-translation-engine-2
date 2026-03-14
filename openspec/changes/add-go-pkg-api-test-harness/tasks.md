## 1. controller API テスト基盤の配置確定

- [ ] 1.1 `pkg/controller/**` の既存テストを棚卸しし、共通化対象のセットアップを洗い出す
- [ ] 1.2 controller 専用 harness の配置先を確定し、責務を DB 初期化・logger・`context.Context`・依存準備に限定する
- [ ] 1.3 `standard_test_spec.md` に整合する controller API テストの標準パターンを決める

## 2. 共通 harness の実装

- [ ] 2.1 controller API テストで再利用する harness または helper を追加する
- [ ] 2.2 インメモリ SQLite、config store、logger、trace 付き `context.Context` を共通生成できるようにする
- [ ] 2.3 controller 固有ロジックを harness に入れず、seed と assert は各テスト側へ残す構成にする

## 3. 既存 controller テストの API テスト化

- [ ] 3.1 `pkg/controller/**` の既存テストを harness 利用へ寄せる
- [ ] 3.2 正常系・異常系を table-driven で表現できる箇所から整理する
- [ ] 3.3 `context.Context` 伝播と error wrap の確認が継続して検証できるようにする

## 4. 品質ゲート確認

- [ ] 4.1 変更中の Go ファイルに対して `npm run backend:lint:file -- <file...>` を逐次実行して違反を解消する
- [ ] 4.2 `npm run lint:backend` を実行し、backend lint 全体が通ることを確認する
- [ ] 4.3 `go test ./pkg/...` と `npm run backend:test` または `npm run backend:check` を実行し、controller API テストが既存導線で通ることを確認する
