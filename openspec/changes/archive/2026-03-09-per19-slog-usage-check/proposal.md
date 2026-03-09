## Why

`backend_coding_standards.md` では `slog.*Context`、lower_snake_case のログキー、機械可読な構造化ログが推奨されているが、現行品質ゲートでは逸脱を十分に検出できない。ログ解析品質を安定させるため、最低限の logging 規約違反を静的検査で落とせるようにする必要がある。

## What Changes

- `tools/backendquality` に `slog` 利用規約を検出する静的チェックを追加する。
- `slog.Info/Error/Warn/Debug` の直接利用、`*Context` 未使用、主要ログ key の lower_snake_case 違反を代表検出対象に含める。
- logger wrapper や許容パターンを区別し、誤検知を抑えるルールを定義する。
- `backend:lint` 導線と fixture ベースの回帰確認を整備する。

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `backend-quality-gates`: バックエンド品質ゲートに `slog` 利用規約の静的チェックを追加する。

## Impact

- Affected code:
  - `tools/backendquality/**`
  - `pkg/**` の `slog` 呼び出し箇所
- Affected dependencies:
  - `golang.org/x/tools/go/analysis`
- Affected process:
  - `npm run backend:lint` の失敗条件に logging 規約違反が追加される
