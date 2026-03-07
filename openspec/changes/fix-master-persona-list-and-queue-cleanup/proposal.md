## Why

`MasterPersona` で、一覧件数の算出、詳細画面の表示内容、画面遷移後の再表示、完了後キューの後始末がそれぞれ実際のデータ状態と一致していない。ペルソナ管理画面の表示契約とタスク完了時のキュー契約をそろえ、ユーザーが全件確認できる状態を安定して維持する必要がある。

## What Changes

- ペルソナ詳細の `RAW response` 表示を、実レスポンスではなく実際に送信した生成リクエストの表示へ置き換え、`生成リクエスト` カラムとして保持・表示する。
- ペルソナ一覧のセリフ数は `npc_personas.dialogue_count` ではなく、関連する `dialogue` レコード数から算出する契約へ変更し、一覧用の誤解を招く集計カラム依存を排除する。
- `MasterPersona` 一覧は画面遷移や再マウント後も消えず、保存済みの全ペルソナを再取得して表示するようにする。
- MasterPersona タスクが `Completed` になった時点で、関連する `llm_queue` job を全件削除し、完了済みタスクにキュー残骸が残らないようにする。
- 必要な DB 変更がある場合は `openspec/specs/database_erd.md` の persona / queue 関連定義を更新する。

## Capabilities

### New Capabilities

- なし

### Modified Capabilities

- `persona`: 一覧件数は関連ダイアログ件数から算出し、詳細画面では生成に使用した実リクエストを確認できる要件へ変更する。
- `queue`: 完了済み MasterPersona タスクに紐づく `llm_queue` job を残さず削除する要件を追加する。
- `task`: 画面遷移やフェーズ完了後の再取得でも MasterPersona 一覧が全件表示されるよう、完了通知と再表示の整合を保証する要件へ変更する。

## Impact

- 影響コード:
  - `frontend/src/pages/MasterPersona.tsx`
  - ペルソナ一覧・詳細表示に関わる frontend store / hooks / Wails binding
  - `pkg/persona/*` の一覧取得・詳細取得・保存 DTO
  - `pkg/task/*` の完了通知と一覧再取得導線
  - `pkg/queue/*` または `llm_queue` 管理ロジック
- 影響データ:
  - persona 詳細で表示する生成リクエスト保存先の追加または既存列の意味変更
  - `llm_queue` 完了後削除に伴う保存ライフサイクルの見直し
  - 必要に応じて `openspec/specs/database_erd.md` の更新
- 依存関係:
  - 新規ライブラリ追加は不要。既存の React、Wails、SQLite 実装で対応する。
