# Proposal: term-translator-slice

## Motivation
`specs/refactoring_strategy.md` の Phase 2 (Term Translation) および `specs/requirements.md` に基づき、`specs/TermTranslatorSlice/` に定義されている仕様の具象実装を開始する。
本変更は、抽出された名詞等（NPC名、アイテム名など）の翻訳とキャッシュへの保存を担う `TermTranslatorSlice` を `pkg/term_translator` に実装することを目的とする。

## Capabilities

### New Capabilities
- `term-translator-slice`: `TermTranslatorSlice` の具象実装パッケージ (`pkg/term_translator`)。

### Modified Capabilities
- なし

## Impact
- `pkg/term_translator` ディレクトリの新規追加および実装クラス群の作成。
- 依存性の注入(DI)設定 (`wire.go` など) への新規プロバイダの追加。
- 構造化ログ (`slog` + OpenTelemetry) および網羅的パラメタライズドテストの導入。
