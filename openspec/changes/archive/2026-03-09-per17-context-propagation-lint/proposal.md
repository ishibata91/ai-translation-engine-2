## Why

`backend_coding_standards.md` では `context.Context` の伝播が MUST だが、現行の `backend:lint` は `context.Background()` の乱用や受領 `ctx` の未伝播を機械的に検出できない。品質ゲートに静的検査を追加し、レビュー依存だった違反を早期に落とせる状態へ進める必要がある。

## What Changes

- `tools/backendquality` から実行できる `context` 伝播違反向けの静的検査を追加する。
- 公開入口で受けた `ctx` の未伝播、`context.Background()` / `context.TODO()` の不適切利用、goroutine 起点での `ctx` 脱落を代表ケースとして検出対象に含める。
- 誤検知を抑えるため、初期化処理や純粋関数など検査対象外にする境界を定義する。
- `backend:lint` 導線へ統合し、代表 fixture で検出結果を回帰確認できるようにする。

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `backend-quality-gates`: バックエンド品質ゲートに `context.Context` 伝播違反を検出するカスタム静的検査を追加する。

## Impact

- Affected code:
  - `tools/backendquality/**`
  - `pkg/**` の `context.Context` 利用箇所
- Affected dependencies:
  - `golang.org/x/tools/go/analysis`
  - 必要に応じて `buildssa` / `inspect` など `go/analysis` 標準拡張
- Affected process:
  - `npm run backend:lint` の失敗条件に `context` 伝播違反が追加される
