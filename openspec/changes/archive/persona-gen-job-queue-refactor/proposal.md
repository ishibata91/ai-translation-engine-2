# Proposal: Persona Gen Job Queue Refactor

## Problem & Motivation
`PersonaGenSlice` は、LLM（Gemini等）を用いてキャラクター（NPC）のプロファイル（ペルソナ）を生成し、`Mod用語DB` へ保存するスライスです。
現在の設計では、`SummaryGenerator` や `TermTranslator` と同様に、自らの内部で複数のNPCレコードに対するLLM処理ループと非同期待機ロジックを抱え込んでいます。
全スライスを統一的かつ安定した「ProcessManager + JobQueue」インフラへ乗せるため、このスライスの設計も破棄し、非結合なコールバック形式（Phase 1 / Phase 2）へ再設計する必要があります。

## Proposal
既存の `PersonaGenSlice` の設計と実装を破棄（大幅リファクタリング）し、新しい「Slice-Owned Callback パターン」に適合させます。
具体的には「Phase 1: 抽出データからプロンプト（Request）を生成する関数」と「Phase 2: LLM結果（Response）を受け取ってSQLiteに保存するコールバック関数」にドメインロジックを分離します。

## Capabilities

### Modified Capabilities
- `PersonaGen`: 単一実行フローを破棄し、準備(Phase1)と保存(Phase2)から成る2フェーズモデルへ再設計。

## Impact
- `pkg/persona_gen` 内部の実装の抜本的変更。
- 外部APIが `GeneratePersonas` 等から、`PreparePrompts` と `SaveResults` へ変更される。
