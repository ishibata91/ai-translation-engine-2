## Why

現状の OpenSpec と Master Persona 実行フローは LM Studio を前提にした記述が多く、UI からクラウド LLM を選べても end-to-end ではローカル LLM 前提の制約が残っている。`pkg/gateway/llm` と `pkg/runtime/queue` にはクラウド同期実行や batch 実行の抽象が既に入り始めているため、Gemini / xAI のクラウド LLM とその batch API を正式な要件として整理し、UI・workflow・config・gateway を同じ契約に揃える必要がある。

## What Changes

- LLM capability を LM Studio 中心の仕様から、ローカル同期実行とクラウド同期 / batch 実行を扱える共通契約へ更新する
- Master Persona を起点に、`sync` / `batch` の実行モード選択、投入後の状態追跡、結果回収、resume / cancel を workflow 要件として定義する
- `ModelSettings` と `MasterPersona` 系 UI で、provider ごとの設定と batch モード選択を扱えるように要件を整理する
- `modelcatalog` を provider / mode 前提のモデル一覧取得仕様へ揃え、UI が provider 固有実装へ直接依存しないことを明確化する
- `config` に provider 別設定、API キー、batch 実行戦略の保存責務を追加し、既存 namespace 運用との整合を定義する

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `llm`: LM Studio 専用前提をやめ、Gemini / xAI を含むクラウド LLM と provider-native batch API を扱える契約へ更新する
- `config`: provider 別 endpoint / model / api_key / batch strategy の保存と復元要件を更新する
- `workflow`: provider と実行モードに応じて sync / batch を切り替え、batch job の追跡と再開を扱えるようにする
- `slice/modelcatalog`: provider 切替時のモデル一覧取得を、クラウド LLM を含む正式要件へ更新する
- `slice/master-persona-ui`: `frontend/src/components/ModelSettings.tsx` と `frontend/src/pages/MasterPersona.tsx` を中心に、クラウド LLM と batch モードを選択できる UI 要件を追加する
- `workflow/master-persona-execution-flow`: provider 固定の開始条件を見直し、Gemini / xAI の batch 実行を含む resume flow へ更新する

## Impact

- 影響コード: `pkg/gateway/llm`, `pkg/runtime/queue`, `pkg/workflow/master_persona_service.go`, `pkg/controller/model_catalog_controller.go`
- 影響 UI: `frontend/src/components/ModelSettings.tsx`, `frontend/src/pages/MasterPersona.tsx`, `frontend/src/hooks/features/modelSettings/useModelSettings.ts`, `frontend/src/hooks/features/masterPersona/useMasterPersona.tsx`
- 影響 API / 契約: Wails controller 経由のモデル一覧取得、Master Persona 開始 / 再開、Config namespace の provider 別解決
- 依存ライブラリ方針: 既存の Wails / React / Go 標準 HTTP 実装を前提にし、新規ライブラリ追加は行わず provider API 差分は既存 `pkg/gateway/llm` に集約する
- DB / ERD 影響: SQLite のテーブル追加は想定せず、既存 `config` / `secrets` / task metadata のキー追加と運用整理で対応する想定
