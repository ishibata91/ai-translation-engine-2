## Why

責務区分と contract が整った後は、最大の実運用経路である MasterPersona を小さく移行して、`controller -> workflow -> persona / runtime` の形が成立することを検証する必要がある。開始、再開、キャンセル、完了 cleanup をまとめて 1 change で移すが、それ以外のユースケースには広げない。

## What Changes

- MasterPersona の開始経路を controller から workflow へ移す
- request enqueue、resume、progress、phase 更新、cleanup を workflow 主導へ移す
- persona との DTO マッピングを workflow に集約する
- 既存 `task` API の互換性を保ちながら内部接続先を workflow へ差し替える
- `pkg/task` に残っている controller 相当責務を整理し、必要なら `pkg/controller` への再配置または lint 区分見直しまで含めて整合させる

## Capabilities

### New Capabilities

### Modified Capabilities
- `task`: 既存 task API の内部実装を workflow 経由へ差し替え、状態通知と進捗報告の責務を整理する
- `persona`: MasterPersona の `PreparePrompts` / `SaveResults` が workflow 経由で呼ばれる前提へ更新する
- `queue`: Completed cleanup と resume の起点を workflow へ寄せる
- `persona-request-preview`: 開始操作が controller -> workflow 経由で処理される前提へ更新する
- `backend-quality-gates`: 実コード配置に合わせて controller / workflow の lint 境界を見直す

## Impact

- 影響コード: `pkg/task`, `pkg/workflow`, `pkg/persona`, `pkg/infrastructure/queue`, Wails binding
- 影響仕様: `openspec/specs/task/spec.md`, `openspec/specs/persona/spec.md`, `openspec/specs/queue/spec.md`, `openspec/specs/persona-request-preview/spec.md`
- API 影響: 公開 Wails API は維持しつつ内部接続先を変更する
