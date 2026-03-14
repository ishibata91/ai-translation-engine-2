## Why

現状の `pkg/controller/**` には個別の Go テストがあるものの、アプリケーション外部から見た公開入口を API としてまとめて検証する実行基盤がない。結果として、controller の公開メソッドに対する退行をどの粒度で確認するか、どのセットアップを共通化するか、どのコマンドで品質ゲートへ接続するかが人依存になっている。

`architecture.md` が定める責務境界では、外部入力の入口は `pkg/controller` に集約される。したがって API テスト基盤も controller を対象に定義し、`standard_test_spec.md` に沿う table-driven なテストを追加しやすい形へ揃える必要がある。

## What Changes

- `pkg/controller/**` の公開 API を対象にした Go テスト実行基盤を追加する
- controller API テストで再利用する共通セットアップ、テストデータ配置、補助ヘルパーの責務を定義する
- `go test ./pkg/...` と既存 `backend:check` / `backend:test` / `lint:backend` の流れに controller API テストを自然に組み込めるようにする
- `standard_test_spec.md` に沿って、controller の公開メソッドを table-driven に検証する方針を明文化する
- de facto standard の範囲で必要な Go テスト支援ライブラリを採用し、過剰な独自基盤は作らない

## Capabilities

### New Capabilities
- `api-test`: `pkg/controller/**` の公開 API を対象に、共通セットアップ、テスト補助、実行コマンド、推奨配置を持つ API テスト実行基盤を定義する

### Modified Capabilities
- `backend-quality-gates`: `pkg/controller/**` の API テストが既存のバックエンド品質確認フローの中で実行され、ローカル確認手順と整合するよう更新する

## Impact

- OpenSpec:
  - `openspec/specs/` 配下に controller API テスト実行基盤の spec を追加する
  - `openspec/specs/backend-quality-gates/spec.md` を更新する
- Backend code:
  - `pkg/controller/**` の公開 API テスト追加方針に影響する
  - controller 向けの共有テスト補助コードや `testdata` 配置方針に影響する
- Tooling / process:
  - `go test ./pkg/...`、`npm run backend:test`、`npm run backend:check` の運用に controller API テストが含まれる
  - 変更時の確認手順として `backend:lint:file -> 修正 -> 再実行 -> lint:backend -> backend:test` の流れがより明確になる
- Dependencies:
  - 原則として Go 標準 `testing` を中心にし、必要な場合のみ `stretchr/testify` や `go.uber.org/goleak` など既に一般的なライブラリに限定して採用を検討する
- Non-goals:
  - `pkg/workflow/**`、`pkg/slice/**`、`pkg/runtime/**` の内部契約テスト基盤整備
  - フロントエンド E2E の追加
  - 本番 API サーバーや外部実サービスへの結合テスト基盤の導入
