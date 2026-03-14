## Why

現在の `parser` / `export` は外部フォーマット境界としての責務が明文化されておらず、slice・workflow・gateway のどこに属するのかを都度判断する必要がある。`PER-32` では Skyrim 入力と xTranslator 出力を `format` 配下へ再配置することで、外部仕様への適応責務を明確にし、今後の format 追加や import 更新の影響範囲を限定する。

## What Changes

- `pkg/format` を外部フォーマット境界として導入し、Skyrim 向け parser と xTranslator 向け exporter の責務配置を明確化する。
- 既存 parser 実装を Skyrim 固有の入力アダプタとして `pkg/format/parser/skyrim` 相当へ再編し、呼び出し側の import / DI 構成を更新する。
- 既存 exporter 実装を xTranslator 固有の出力アダプタとして `pkg/format/exporter/xtranslator` 相当へ再編し、呼び出し側の import / DI 構成を更新する。
- OpenSpec 上でも parser / export の capability を `format` 境界前提で再整理し、責務説明と依存方向を更新する。
- 既存の XML 出力仕様および抽出 JSON の解釈仕様は維持し、配置変更による回帰が起きないことを確認する。

## Capabilities

### New Capabilities
- `format`: 外部フォーマット境界として、Skyrim parser と xTranslator exporter の配置原則、依存方向、workflow との接続点を定義する。

### Modified Capabilities
- `export`: xTranslator XML 出力 capability の責務所属を `format` 境界へ更新し、xTranslator 固有 exporter として扱う要件へ変更する。
- `slice/parser`: 抽出 JSON の読み込み capability の責務所属を `format` 境界へ更新し、Skyrim 固有 parser として扱う要件へ変更する。

## Impact

- 影響コード: `pkg/parser` 系、`pkg/export` 系、関連する Wire / DI 初期化、workflow / pipeline からの import。
- 影響ドキュメント: `openspec/specs/architecture.md` を参照しつつ、`openspec/specs/export/spec.md`、`openspec/specs/slice/parser/spec.md`、新規 `openspec/specs/format/spec.md` を対象に差分整理する。
- 外部仕様影響: Skyrim 抽出 JSON と xTranslator XML の入出力仕様自体は維持するため、利用者向け API / ファイル形式の破壊的変更は想定しない。
- 依存ライブラリ: 新規ライブラリ追加は想定せず、既存の Go 標準ライブラリと現行 DI 構成を継続利用する。
