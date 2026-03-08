# violation-targets

## 調査方針

- 対象: `pkg/**`
- 参照規約: `openspec/specs/backend_coding_standards.md`
- 補助確認: `npm run backend:lint` は 2026-03-08 時点で通過
- 位置づけ: lint では検知されない規約違反候補を、修正タスク化しやすい粒度で整理する

## 優先度高

### 1. `context.Context` 未伝播と `context.Background()` への置換

規約根拠:
- 「`context.Context` は公開 Contract メソッドの第一引数で受け取り、内部処理へ必ず伝播すること」
- 「テストおよび本番コードのデバッグは構造化ログ前提で行い、問題解析のために `context.Context` を途切れさせないこと」

対象:
- `pkg/task/bridge.go`
  - `GetAllTasks()` が `context.Background()` を使って `manager.store.GetAllTasks` を呼ぶ
  - `GetTaskRequestState(taskID string)` が `context.Background()` を使って workflow を呼ぶ
  - `GetTaskRequests(taskID string)` が `context.Background()` を使って workflow を呼ぶ
- `pkg/config/config_service.go`
  - `UIStateGetJSON`, `UIStateSetJSON`, `UIStateDelete`, `ConfigGet`, `ConfigSet`, `ConfigSetMany`, `ConfigDelete`, `ConfigGetAll` が全て `context.Background()` 固定
- `pkg/task/manager.go`
  - `ResumeTask`, `AddTaskWithCompletionStatus`, `markTaskRunning`, `throttleDBUpdate`, `handleTaskCompletionWithStatus`, `runCompletionHooks`, `handleTaskFailure`, `handleTaskCancellation` で store / hook 呼び出しに `context.Background()` を多用
- `pkg/workflow/master_persona_service.go`
  - `ResumeMasterPersona(_ context.Context, taskID string)` が受け取った ctx を破棄
  - `CancelMasterPersona(_ context.Context, taskID string)` が受け取った ctx を破棄

修正タスクの意図:
- controller / Wails binding から store・workflow・hook 末端まで同一コンテキストを通す
- `task.Manager` 内の background 実行は、起点 ctx から必要な情報を引き継いだ専用 ctx に寄せる

### 2. error wrap 欠落

規約根拠:
- 「返却する `error` は呼び出し元で原因を追跡できるように `fmt.Errorf("...: %w", err)` などで文脈を付与すること」

対象:
- `pkg/pipeline/store.go`
  - `SaveState`, `GetState`, `ListActiveStates`, `DeleteState`, `runMigrations` が DB エラーをそのまま返す
- `pkg/workflow/master_persona_service.go`
  - `StartMasterPersona` 内で `LoadExtractedJSON`, `PreparePrompts`, `SubmitTaskRequests`, `manager.AddTaskWithCompletionStatus` 由来の error をほぼ素通しで返す
  - `runPersonaExecution` 内で `GetTaskRequestState`, `PrepareTaskResume`, `ProcessProcessIDWithOptions`, `persistPersonaResponses` 由来の error をそのまま返す箇所がある
  - `persistPersonaResponses` が queue / reporter / personaGenerator 由来 error を文脈なしで返す
- `pkg/config/config_service.go`
  - `ConfigGet`, `ConfigSetMany`, `ConfigGetAll` が呼び出し文脈を付けずに error を返す
- `pkg/task/manager.go`
  - `ResumeTask` の `GetAllTasks` 失敗時に error をそのまま返す

修正タスクの意図:
- package 境界をまたぐ返却は原則 wrap し、`task_id`, `namespace`, `process_id` など識別子に応じた文脈を保持する

## 優先度中

### 3. 構造化ログの Context 非利用 / ログメッセージの機械可読性不足

規約根拠:
- 「ログ出力は `slog.*Context` を使用し、構造化フィールドで slice 名・入力件数・識別子などの分析可能な情報を記録すること」
- 「ログメッセージは機械可読な識別子を優先し、日本語説明文は必要最小限に留める」
- 「ログキー名は `slice`, `record_count`, `task_id` のように意味が固定された lower_snake_case を優先する」

対象:
- `pkg/task/manager.go`
  - `runCompletionHooks` の `m.logger.Error(...)` が `ErrorContext` ではない
  - `throttleDBUpdate` の `m.logger.Error(...)` が `ErrorContext` ではない
  - `Initialize` の `Initializing TaskManager`、`ResumeTask` の一部ログなど human-readable 寄りのメッセージが残る
- `pkg/pipeline/handler.go`
  - `Failed to upgrade websocket`, `Failed to execute slice`, `Failed to encode start process response` など lower_snake_case でないメッセージが残る
- `pkg/pipeline/manager.go`
  - `ENTER Recover`, `EXIT Recover`, `Recovering process` などトレースしにくいメッセージ形式が残る

修正タスクの意図:
- `slog.*Context` に統一し、メッセージを `task.resume.started` のような識別子寄りへ寄せる
- 重要ログは `task_id`, `process_id`, `slice`, `namespace` などの固定キーをそろえる

### 4. 失敗の握りつぶし

規約根拠:
- error wrap と構造化ログ前提の解析方針に反する

対象:
- `pkg/workflow/master_persona_service.go`
  - `updatedState, stateErr := s.queue.GetTaskRequestState(...)` の失敗時に `return nil` しており、失敗理由を落としている
  - `s.manager.Store().SaveMetadata(...)` を `_ =` で捨てる箇所がある
- `pkg/pipeline/manager.go`
  - `m.store.DeleteState(...)` を `_ =` で捨てる
  - `state, err := m.store.GetState(...)` 失敗時ログに `err` 自体の属性が残らない

修正タスクの意図:
- 失敗を無視してよい箇所と許されない箇所を切り分け、許容時も warning と理由を残す

## 優先度低

### 5. 公開 API の責務過多候補

規約根拠:
- 「公開メソッドは 1 つの責務に保ち、複雑な分岐や処理列は同一ファイル内のプライベートメソッドへ分割すること」

対象:
- `pkg/task/manager.go`
  - `ResumeTask` が状態解決、runner 解決、DB 更新、goroutine 起動、ログ出力まで担っている
- `pkg/workflow/master_persona_service.go`
  - `StartMasterPersona` が入力検証、task metadata 構築、JSON 読み込み、DTO 変換、queue 保存、進捗通知、ログ出力を一括で担っている
  - `runPersonaExecution` が resume 準備、進捗通知、worker 実行、保存、完了イベント送出まで担っている

修正タスクの意図:
- 主要公開メソッドはフローの入口に絞り、同一ファイル内 private method 抽出で責務を見える化する

## 備考

- 今回の change ではまず上記の高優先度項目を Must 違反中心に是正し、中・低優先度はスコープに応じて段階対応する想定
- `backend:lint` で未検知だったため、proposal / design / tasks では lint 追加で防げるものとコード修正でしか防げないものを分けて整理するとよい

## 2026-03-08 実装確認メモ

- `pkg/task/bridge.go`, `pkg/task/manager.go`, `pkg/workflow/master_persona_service.go` では `context.Background()` 固定呼び出しを除去し、保持済み ctx または引数 ctx を store / queue / hook / runner へ伝播する形へ更新した。
- `pkg/config/config_service.go` は `ConfigController` 方針へ整理し、Wails 起点の ctx を保持して `UIStateStore` / `Config` 呼び出しへ流すようにした。互換維持のため `ConfigService` alias と `NewConfigService` ラッパーは残している。
- `pkg/pipeline/store.go`, `pkg/pipeline/manager.go`, `pkg/pipeline/handler.go` では DB / queue / save / cleanup 失敗を文脈付き error と warning/error log へ整理し、メッセージを識別子ベースへ統一した。
- `tools/backendquality/main.go` は確認したが、今回の違反群に対して低コストで安定運用できる静的検査追加は見送った。`context.Background()` 検出や error wrap 強制は誤検知コストが高く、現状の `golangci-lint + go-cleanarch + lint:file` 運用へ素直に乗らないため、今回はコード是正とテスト追加を優先した。
