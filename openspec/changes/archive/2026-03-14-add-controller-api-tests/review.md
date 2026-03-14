# add-controller-api-tests Review Checklist

この change の実装レビュー結果を記録する。
共通観点は `openspec/review_standard.md` を前提とする。

## 1. ユーザが出した完了条件

- 完了条件 1: `pkg/controller` の現行 controller 公開 API が table-driven の API テストで回帰保護されていること
- 完了条件 2: Dictionary Builder / Master Persona の主要導線で入力写像・error 伝播・context 伝播を検証できること
- 完了条件 3: バックエンド品質ゲート（`backend:lint:file -> lint:backend -> go test ./pkg/...`）が通ること

## 2. 品質ゲート確認

### Backend

- [x] 変更中ファイルに対して `npm run backend:lint:file -- <file...>` を逐次実行した
- [x] `backend:lint:file -> 修正 -> 再実行 -> 最後に lint:backend` の順で進めた
- [x] 作業中または完了前に `npm run lint:backend` を実行した
- [x] 必要に応じて `npm run backend:check` または `npm run backend:watch` で品質確認した（本変更では `go test ./pkg/...` を実行）

## 3. 実行メモ

- 実行したコマンド:
  - `npm run backend:lint:file -- pkg/controller/dictionary_controller.go pkg/controller/file_dialog_controller.go pkg/controller/model_catalog_controller.go pkg/controller/persona_controller.go pkg/controller/persona_task_controller.go pkg/controller/task_controller.go pkg/workflow/task/manager.go pkg/workflow/master_persona_service.go pkg/workflow/pipeline/mapper.go`
  - `npm run lint:backend`
  - `go test ./pkg/...`
- 未実行の品質ゲートと理由:
  - フロントエンド品質ゲートは対象外（バックエンド変更のみ）
- レビュー時の補足:
  - `FileDialogController` は seam (`openFileDialog` / `openMultipleFilesDialog`) 経由でテスト可能
  - Dictionary は trace_id の伝播一致を検証
  - PersonaTask は `ResumeTask` / `GetTaskRequestState` / `GetTaskRequests` の workflow error 伝播を追加検証
