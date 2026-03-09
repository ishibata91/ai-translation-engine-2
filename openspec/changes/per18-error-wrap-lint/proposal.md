## Why

`backend_coding_standards.md` では package 境界や外部 I/O 境界の error wrap が MUST だが、現行品質ゲートでは `return err` や `%w` 欠落を十分に検出できない。レビューでしか見つからない文脈不足エラーを減らすため、品質ゲートに専用検査を追加する必要がある。

## What Changes

- `tools/backendquality` に error wrap 不足を検出するカスタム lint を追加する。
- package 境界での `return err`、`fmt.Errorf` における `%w` 欠落、cleanup 以外の失敗握りつぶしを代表違反として検出対象に含める。
- 例外として許容する cleanup や sentinel error 変換などの境界を定義し、誤検知を抑える。
- `backend:lint` 導線と fixture ベースの回帰確認を整備する。

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `backend-quality-gates`: バックエンド品質ゲートに error wrap 不足を検出するカスタム lint を追加する。

## Impact

- Affected code:
  - `tools/backendquality/**`
  - `pkg/**` の error return / `fmt.Errorf` 利用箇所
- Affected dependencies:
  - `golang.org/x/tools/go/analysis`
- Affected process:
  - `npm run backend:lint` の失敗条件に error wrap 不足が追加される
