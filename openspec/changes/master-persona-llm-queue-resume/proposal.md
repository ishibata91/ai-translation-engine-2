## Why

`MasterPersona` の LLM リクエストは現在 `pkg/llm` と分離されており、進捗表示・中断復帰・再起動後再開を一貫して扱えない。先に LM Studio のみを対象として、リクエストのキュー永続化と再開経路を統合し、UI/Task/Queue/LLM の契約をそろえる必要がある。

## What Changes

- `pkg/task/master_persona_task.go` を `pkg/llm` 実行経路へ統合し、MasterPersona リクエストを LLM キューに保存して実行する。
- `progress` を用いた段階別進捗通知を追加し、UI でリクエスト作成・送信・保存の進捗を可視化する。
- 再起動後およびキャンセル後の再開を可能にするため、Queue/Task 側で再開可能なリクエスト状態と再実行コンテキストを保持する。
- 対応プロバイダは今回 `lmstudio` のみに限定し、他プロバイダの再開処理は対象外とする。
- `frontend/src/pages/MasterPersona.tsx` の LLM 設定（provider/model/endpoint 等）を永続化・再読込可能にする。
- キューマネージャー経由でのリクエスト再開 API/UI 導線を追加する。
- 実装と受け入れを以下 4 段階に分割し、各段階でユーザー動作テストを実施する。
- 1) LLM キュー保存
- 2) Config 永続化
- 3) LM Studio 経由実行
- 4) ペルソナ保存

## Capabilities

### New Capabilities

- `master-persona-lmstudio-resume-flow`: MasterPersona の段階実行（保存→実行→保存）と再開導線を定義する統合フロー。

### Modified Capabilities

- `task`: MasterPersona タスクでのキュー投入、キャンセル後再開、再起動後再開の契約を追加する。
- `queue`: LLM リクエスト単位の永続状態、再開条件、キューマネージャーからの再開契約を追加する。
- `llm`: `lmstudio` 実行時の再開コンテキスト利用と途中再開時の実行契約を追加する。
- `progress`: MasterPersona の段階別進捗イベント（投入/送信/保存）を追加する。
- `persona`: MasterPersona 経路でのペルソナ保存フェーズと再開時の冪等保存要件を追加する。
- `config`: MasterPersona 画面で使用する LLM 設定の永続化・読込契約を追加する。

## Impact

- Affected backend code:
- `pkg/task/master_persona_task.go`
- `pkg/queue/*`
- `pkg/llm/*`（LM Studio 実行経路）
- `pkg/progress/*`
- `pkg/persona*` または persona 保存に関わる slice
- `pkg/config/*`
- Affected frontend code:
- `frontend/src/pages/MasterPersona.tsx`
- キューマネージャー画面/ストア（再開操作）
- データ影響:
- キューリクエスト状態・再開メタデータの保存先拡張が必要（既存テーブル拡張または追加テーブル）。設計で ERD 影響を確定する。
- 依存ライブラリ（デファクト前提）:
- 既存 `database/sql` + SQLite ドライバ、`context`/`errgroup`、Wails バインディングを利用し、新規に非標準ライブラリは導入しない。
