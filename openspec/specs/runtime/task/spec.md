# Frontend Job Manager 仕様

## Purpose
TBD: フロントエンドからの長時間のタスク（辞書構築、翻訳プロジェクトなど）の管理、状態の永続化、再開、フェーズ完了通知などの処理に関する仕様を定義する。

## Requirements

### Requirement: タスクの永続化と再開ロジック
FrontendTaskManager は、フロントエンド向けタスク（辞書構築、ペルソナ抽出等）の状態を SQLite に永続化しなければならない。
タスクモデルには、再開に必要な十分なコンテキスト（`target_file`, `options` などのJSONメタデータ）と `Phase` (現在のステップ) を含めること。

#### Scenario: 予期せぬ終了からの復帰
- **WHEN** 翻訳タスク実行中にアプリが強制終了し、再起動された
- **THEN** アプリ初期化時に TaskManager が DB の `Running` または `Paused` 状態のタスクを検出し、
- **AND** フロントエンドでの Resume 呼び出し時に保存されたメタデータとフェーズを元に処理を再開する。

### Requirement: フェーズ毎の完了通知とデータ反映（Phase Completion）
タスクが複数フェーズに分割されている場合（例：「ペルソナ抽出処理」から「要約処理」への移行など）、1つのフェーズが完了した時点で、そのフェーズで生成された一括データをメインDBへ保存し、フロントエンドへ個別のイベント `task:phase_completed` をEmitしなければならない。
※N+1問題を防ぐため、1NPCごと・1単語ごとの単体データEmitは行わず、ループ処理中は進捗(Progress)のみを更新し、フェーズの区切りでデータを反映させること。MasterPersona では当該イベントを受信した UI が保存済み全ペルソナ一覧を再取得できるよう、再取得に十分な task 情報を通知しなければならない。

#### Scenario: ペルソナ抽出フェーズの完了とフロントエンド反映
- **WHEN** ユーザーが「翻訳プロジェクト開始」タスクを実行中で、「ペルソナ抽出フェーズ」のループ処理がすべて終わった
- **THEN** TaskManager は生成された全NPCデータをDBにコミットし、`task:phase_completed`（種別: 'persona_extracted'）イベントをEmitする。
- **AND** フロントエンドの状態ストアはそれを受信して、ペルソナ一覧表示の再取得や画面の更新を一括で実施する。

#### Scenario: 再表示時も完了済みデータを再取得できる
- **WHEN** ユーザーが phase 完了後に別画面へ移動してから MasterPersona へ戻る
- **THEN** フロントエンドは完了済みフェーズの永続データを再取得し、保存済み全ペルソナ一覧を再表示しなければならない

### Requirement: ダッシュボードからのフェーズ復帰と画面遷移
タスクの現在の `Phase` が更新された場合、フロントエンドのダッシュボードはそのフェーズ情報を認識し、ユーザーがクリックした際に適切な画面（と必要なパラメータ）へルーティングできる仕組みを提供しなければならない。

#### Scenario: 複数ステップの翻訳プロジェクトの途中遷移
- **WHEN** ダッシュボードのタスク一覧で、ステータスが「ペルソナ抽出待機中(Phase: PersonaExtraction)」となっているタスクをクリックした
- **THEN** アプリケーションは翻訳プロジェクト詳細画面へ遷移し、自動的に「ペルソナパネル」のタブを開いて状態を復元する。

### Requirement: Wailsバインディングの拡張
Task 管理機能は、controller 層が従来の `GetActiveTasks`, `CancelTask`, `ResumeTask` などの Wails バインディングを提供しつつ、MasterPersona の実際のユースケース進行は workflow 契約へ委譲しなければならない。

#### Scenario: Wails バインディングは workflow を経由して再開する
- **WHEN** UI から `ResumeTask` が呼び出される
- **THEN** controller は workflow の再開契約を呼び出さなければならない

#### Scenario: 既存の公開 API は維持される
- **WHEN** フロントエンドが既存の `GetActiveTasks`, `CancelTask`, `ResumeTask` を利用する
- **THEN** システムは同等の公開 API を継続提供しなければならない

### Requirement: Task は進捗値を決定して Progress へ報告しなければならない
`task` / workflow は段階進捗の `phase/current/total/message` を決定し、`progress` へ報告しなければならない。`total` は全 request 件数を分母として扱わなければならない。usecase slice は UI 配信用の進捗値を直接決定せず、進行事実や中間結果を `task` / workflow へ返さなければならない。

#### Scenario: フェーズ遷移時に task が進捗を報告する
- **WHEN** Task が enqueue/dispatch/save/complete の各段階へ遷移する
- **THEN** `task` は対応する `phase` と進捗値を `progress` へ報告しなければならない

#### Scenario: Progress は task が渡した値をそのまま配信する
- **WHEN** `task` が `phase/current/total/message` を `progress` へ報告する
- **THEN** UI へ配信される進捗値は task が指定した値と一致しなければならない

#### Scenario: usecase slice は task 向けに進行事実を返す
- **WHEN** usecase slice が途中経過を task / workflow へ伝える必要がある
- **THEN** usecase slice は進行事実または中間結果を返さなければならない
- **AND** `progress` へ直接報告してはならない

### Requirement: Task 状態遷移は task イベントで UI へ通知しなければならない
`task` の状態遷移（`pending/running/completed/failed/canceled`）は `task` イベント（例: `task:updated`）として UI に通知しなければならない。状態遷移通知は `progress` 経路を介してはならない。

#### Scenario: task 状態が UI へ直接通知される
- **WHEN** task 状態が `running` から `completed` に変化する
- **THEN** UI は `task` イベントで状態変更を受け取り、ステータス表示を更新しなければならない

### Requirement: MasterPersona タスクは中断点を保持して再開可能でなければならない
MasterPersona タスクは workflow 主導で、`task_id`, `phase`, `resume_cursor`, `overwrite_existing` を保持し、再起動後およびキャンセル後の再開要求を受理しなければならない。再開時のモデル・パラメータは常に `config` から再読込しなければならない。

#### Scenario: 再起動後に再開できる
- **WHEN** MasterPersona 実行中にアプリが終了し、再起動後に `ResumeTask` が呼ばれる
- **THEN** workflow は保存済み再開コンテキストを読み出し、未完了フェーズから処理を再開しなければならない

#### Scenario: キャンセル後に再開できる
- **WHEN** MasterPersona タスクが cancel 済みで、ユーザーが再開を要求する
- **THEN** workflow は完了済み request を再実行せず、未完了 request のみを再開しなければならない

#### Scenario: 再開時に上書き方針が保持される
- **WHEN** `overwrite_existing=false` で開始した MasterPersona タスクを再開する
- **THEN** workflow は開始時 metadata の `overwrite_existing` を再利用し、再開時に `true` として扱ってはならない

#### Scenario: REQUEST_GENERATED 後の自動再開要求を受理できる
- **WHEN** UI が `REQUEST_GENERATED` 直後に同一 task ID で `ResumeTask` を呼び出す
- **THEN** workflow はその再開要求を受理し、未完了 request の dispatch/save フェーズへ遷移しなければならない

### Requirement: MasterPersona 画面は再表示時に保存済み全ペルソナを再取得できなければならない
Task は、MasterPersona のフェーズ完了および状態更新が UI の一覧再取得トリガーとして利用できるよう、画面再表示後も同一 task に依存せず保存済みデータを再取得できる通知契約を提供しなければならない。UI はローカル state の残存有無に関わらず、永続化済みペルソナ一覧を再ロードできなければならない。

#### Scenario: 画面遷移後に一覧を再取得できる
- **WHEN** ユーザーが MasterPersona 画面を離れてから再度開く
- **THEN** システムは保存済み全ペルソナ一覧を再取得できなければならない
- **AND** 直前表示時のローカル state が失われていても一覧が空のままになってはならない

#### Scenario: 完了通知後に一覧再取得が行える
- **WHEN** MasterPersona タスクが `Completed` または関連フェーズ完了イベントを UI へ通知する
- **THEN** UI はその通知を契機に保存済み全ペルソナ一覧を再取得できなければならない

### Requirement: Task 管理は task 削除時に task manager 管理対象だけを整理しなければならない
Task 管理は、task 削除要求を受けたとき、frontend task の永続レコード削除と task manager が保持する管理対象の整理を同一の削除導線で実行しなければならない。task 削除は artifact 正本や共有成果物を削除してはならない。

#### Scenario: 停止済み task を削除する
- **WHEN** ユーザーが `pending`、`paused`、`request_generated`、`failed`、`cancelled` のいずれかの status を持つ task の削除を要求する
- **THEN** システムは `frontend_tasks` の当該 task レコードを削除しなければならない
- **AND** task manager が保持する当該 task の管理対象を整理しなければならない
- **AND** artifact 正本を削除してはならない

#### Scenario: 実行中 task は削除できない
- **WHEN** ユーザーが `running` status の task の削除を要求する
- **THEN** システムは task 削除を拒否しなければならない
- **AND** ユーザーには停止後に削除できることを通知しなければならない
