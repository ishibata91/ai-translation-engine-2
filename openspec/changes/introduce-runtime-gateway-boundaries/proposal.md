## Why

実装 change を小さく進めるには、まず `workflow`、`runtime`、`gateway` の責務境界と contract を明確にし、DI lint で破れない状態を作る必要がある。境界が未確定なまま MasterPersona を移行すると、package 移動と振る舞い変更が同時に起きて検証範囲が広がりすぎる。

## What Changes

- `workflow` を安定層として導入し、controller が依存する契約を定義する
- `queue` `progress` `workflow state` などを `runtime` として再定義する
- `llm` `datastore` `config` などを `gateway` として再定義する
- `runtime -> gateway` の限定依存を明記し、queue worker から LLM gateway を利用できる構造を定義する
- `go-cleanarch` を品質ゲートに追加し、責務区分の依存方向違反を検出する

## Capabilities

### New Capabilities
- `workflow`: controller から usecase slice と runtime を接続する安定層の契約、責務、DI 境界を定義する

### Modified Capabilities
- `queue`: queue を runtime として扱い、workflow 経由利用と `runtime -> gateway` の限定依存を要件へ反映する
- `persona`: persona slice が runtime 制御を持たず、gateway 契約だけに依存する前提を要件へ反映する
- `backend-quality-gates`: `go-cleanarch` を依存方向 lint として品質ゲートへ追加する

## Impact

- 影響コード: `pkg/workflow`, `pkg/infrastructure/queue`, `pkg/infrastructure/llm`, `pkg/infrastructure/datastore`, 品質ゲート導線
- 影響仕様: `openspec/specs/queue/spec.md`, `openspec/specs/persona/spec.md`, `openspec/specs/backend-quality-gates/spec.md`
- 追加依存: `go-cleanarch`
