## Why

現在のダッシュボードではタスクを停止・再開できても、不要になった task 自体を削除できない。そのため失敗済み・中止済み・再開不要な task が一覧に残り続け、ユーザーが task 管理対象を整理できない。

## What Changes

- ダッシュボードの task 行に削除アクションを追加し、再開不要な task を UI から削除できるようにする。
- task 削除は `frontend_tasks` の永続レコードと task manager が保持する管理対象だけを削除し、artifact 正本は削除対象から除外する。
- task 削除 API を `TaskController.DeleteTask` に一本化し、実行中 task の扱いと削除後の一覧反映を明確化する。
- 削除確認は既存デザインに馴染むシンプルな modal で行う。

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `frontend/dashboard`: ダッシュボードが task ごとの削除操作、modal 確認導線、削除後の一覧反映を提供するよう要件を変更する
- `runtime/task`: Task 管理が `TaskController.DeleteTask` を入口として task 永続レコード削除と task manager 管理対象の整理を提供するよう要件を変更する

## Impact

- `frontend/src/pages/Dashboard.tsx` と `frontend/src/hooks/features/dashboard/useDashboard.ts` に削除 UI と削除アクションが追加される。
- `frontend/src/store/taskStore.ts` と Wails bindings に task 削除 API が追加される。
- `pkg/controller/task_controller.go` と `pkg/workflow/task/*` に task 削除 orchestration が追加される。
- artifact 正本はこの change の削除対象に含めないため、artifact 系 spec と保存契約は変更しない。
