## Why

`openspec/specs/governance/requirements/spec.md` が定義する 2-Pass System では、Pass 1 の単語翻訳フェーズで翻訳対象の固有名詞や用語を事前に翻訳し、本文翻訳より先に用語辞書を構築することが求められている。現在の translation flow にはこのフェーズが明示されておらず、翻訳対象単語を先行翻訳してから本文翻訳へ渡す導線が不足している。

加えて、Pass 1 で使う LLM リクエスト設定と prompt は運用中に調整したくなるが、translation flow 側にはその編集 UI がない。既存の `ModelSettings` と `PromptSettingCard` を単語翻訳フェーズでも再利用できるようにし、operator が事前翻訳の品質を画面上で調整できるようにする必要がある。

## What Changes

- translation flow に `単語翻訳` フェーズを追加し、本文翻訳の前に翻訳対象単語を terminology slice で LLM 翻訳できるようにする。
- 単語翻訳フェーズは、slice が artifact から対象単語レコードを抽出し、terminology の 2 フェーズモデルに従って `PreparePrompts` と `SaveResults` を実行する。
- 単語翻訳フェーズの出力は `terminology.db` に保存し、mod ごとのテーブルで管理する。
- workflow が保持する実行結果サマリは `対象件数 / 保存件数 / 失敗件数` に限定する。
- translation flow UI に単語翻訳フェーズの表示を追加し、対象件数、進行状態、結果件数に加えて、LLM リクエスト設定と prompt を編集できるようにする。
- `ModelSettings` と `PromptSettingCard` は master persona 専用の見せ方に閉じず、単語翻訳フェーズでも再利用できる共有 UI として扱う。

## Capabilities

### New Capabilities
- `translation-flow-terminology-phase`: translation flow における単語翻訳 phase の順序、terminology 起動、後続 handoff を定義する。
- `translation-flow-terminology-phase-ui`: translation flow 画面で単語翻訳 phase の状態表示、進行表示、結果確認、LLM リクエスト設定、prompt 編集を定義する。

### Modified Capabilities
- `terminology`: translation flow からの起動条件、artifact 読み出し境界、`terminology.db` での保存方式、operator 指定の request/prompt 設定反映を追加する。
- `master-persona-ui`: `ModelSettings` と `PromptSettingCard` を他 feature でも使える共有 UI として扱う条件を追加する。

## Impact

- `pkg/workflow`: translation flow の phase 進行と terminology 起動 orchestration を追加する。
- `pkg/slice/terminology`: artifact からの対象抽出、LLM 単語翻訳、`terminology.db` 保存、request/prompt 設定反映を担う。
- `pkg/artifact/translationinput` または同等の入力 artifact: terminology slice の読み出し元として利用する。
- `frontend/src/components/translation-flow`: 単語翻訳フェーズの UI、進行表示、LLM 設定、prompt 編集 UI を追加する。
- `frontend/src/components/ModelSettings.tsx`: terminology phase でも使える LLM リクエスト設定 UI へ拡張する。
- `frontend/src/components/masterPersona/PromptSettingCard.tsx`: terminology phase でも使える prompt 編集カードとして再利用する。
- 既存依存の再利用を前提とし、新規ライブラリ追加は想定しない。
