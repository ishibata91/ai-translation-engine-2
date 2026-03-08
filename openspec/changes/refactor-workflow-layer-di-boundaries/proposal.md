> この change は実装単位として大きすぎるため、以下の 3 change に分割して進める。
> - `reorganize-architecture-spec-boundaries`
> - `introduce-runtime-gateway-boundaries`
> - `migrate-master-persona-to-workflow`

## Why

現行の `pkg/task` はタスク状態管理、UI バインディング、MasterPersona の実行手順、`queue` 連携、`persona` 保存再開制御を同時に抱えており、安定層とスライス境界が不明瞭になっている。今後 `persona` 以外のユースケーススライスを追加しても接続責務を破綻させないために、`controller -> workflow -> usecase slice` の構造を明示し、`runtime` と `gateway` を論理的に分離したうえで、全接続を DI インターフェース経由へ統一する必要がある。

## What Changes

- `pkg/task` を中心に持っていた UI 入口、状態管理、ユースケース進行の責務を再整理し、安定層として `workflow` を導入する。
- `controller` は `workflow` の契約だけに依存し、`workflow` が `persona`、`parser`、`runtime` などの契約をオーケストレーションする構造へ変更する。
- `persona` はスライス固有の入力・出力契約と保存責務に集中し、`queue` 制御や UI 都合の状態遷移を持たないようにする。
- `queue` や `progress` などの実行制御系は `runtime`、DB や LLM などの外部依頼系は `gateway` として再区分し、それぞれの責務を明確化する。
- 依存注入は原則としてインターフェース経由とし、具象型の直接接続箇所を削減する。
- `openspec/specs/architecture.md` を更新し、責務区分、依存方向、DI 規約を明文化する。
- `go-cleanarch` を導入し、依存方向と DI 境界の lint を品質ゲートへ追加する。

## Capabilities

### New Capabilities
- `workflow`: controller から usecase slice と runtime を接続する安定層の契約、責務、DI 境界を定義する

### Modified Capabilities
- `task`: タスク管理の責務を workflow 管理配下へ再整理し、UI への状態通知と進捗報告の契約境界を更新する
- `persona`: persona slice が workflow から契約経由で呼ばれ、runtime 制御を持たない構造へ要件を更新する
- `queue`: queue をスライス非依存の runtime として扱い、workflow 経由利用の前提を要件へ反映する

## Impact

- 影響コード: `pkg/task`, `pkg/persona`, `pkg/infrastructure/queue`, `pkg/infrastructure/llm`, `pkg/infrastructure/datastore`, `main.go` と関連 DI 配線
- 影響仕様: `openspec/specs/architecture.md`, `openspec/specs/task/spec.md`, `openspec/specs/persona/spec.md`, `openspec/specs/queue/spec.md`
- 影響 API: Wails binding の接続先、および controller が依存するバックエンド契約
- 追加依存: `go-cleanarch` を DI / 依存方向 lint 用候補として導入する
- DB 影響: 本変更の主眼は責務分離であり必須の新規スキーマ変更は想定しないが、既存 queue/task/persona の永続化責務境界に関する記述を見直す
