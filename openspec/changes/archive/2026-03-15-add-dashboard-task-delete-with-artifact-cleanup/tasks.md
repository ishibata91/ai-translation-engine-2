## 1. Backend task delete 基盤

- [x] 1.1 `pkg/workflow/task` に task 永続レコード削除 API を追加する
- [x] 1.2 `pkg/workflow/task.Manager.DeleteTask` を実装し、実行中 task の削除拒否、manager 管理対象の整理、`frontend_tasks` 削除の順序を保証する
- [x] 1.3 `pkg/controller/task_controller.go` に `DeleteTask` を追加し、Wails binding の公開入口を `TaskController.DeleteTask` に一本化する
- [x] 1.4 `pkg/workflow/task` と `pkg/controller/task_controller.go` の変更ファイルに対して `npm run backend:lint:file -- <file...>` を実行し、指摘を解消して再実行する

## 2. Dashboard 削除 UI

- [x] 2.1 `frontend/src/store/taskStore.ts` に task 削除 action を追加し、成功時に task 一覧 state から削除対象を除去できるようにする
- [x] 2.2 `frontend/src/hooks/features/dashboard/useDashboard.ts` に削除対象選択、modal 開閉、削除確定 action を追加する
- [x] 2.3 `frontend/src/pages/Dashboard.tsx` に削除ボタンとシンプルな確認 modal を追加し、削除可能 status にだけ表示する
- [x] 2.4 フロント変更ファイルに対して `frontend/` 配下で `npm run lint:file -- <file...>` を実行し、指摘を解消して再実行する

## 3. 検証と品質ゲート

- [x] 3.1 backend 側の task 削除と実行中拒否を検証するユニットテストを追加する
- [x] 3.2 frontend 側の Dashboard 削除表示、modal 確認、削除成功時反映を検証するテストを追加または更新する
- [x] 3.3 `npm run lint:backend` を実行して backend 品質ゲートを通す
- [x] 3.4 `frontend/` 配下で `npm run typecheck` と `npm run lint:frontend` を実行する
- [x] 3.5 Playwright E2E で Dashboard から task を削除するシナリオを実行し、削除後に一覧から消えることを確認する
