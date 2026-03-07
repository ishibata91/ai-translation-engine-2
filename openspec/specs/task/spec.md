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
※N+1問題を防ぐため、1NPCごと・1単語ごとの単体データEmitは行わず、ループ処理中は進捗(Progress)のみを更新し、フェーズの区切りでデータを反映させること。

#### Scenario: ペルソナ抽出フェーズの完了とフロントエンド反映
- **WHEN** ユーザーが「翻訳プロジェクト開始」タスクを実行中で、「ペルソナ抽出フェーズ」のループ処理がすべて終わった
- **THEN** TaskManager は生成された全NPCデータをDBにコミットし、`task:phase_completed`（種別: 'persona_extracted'）イベントをEmitする。
- **AND** フロントエンドの状態ストアはそれを受信して、ペルソナ一覧表示の再取得や画面の更新を一括で実施する。

### Requirement: ダッシュボードからのフェーズ復帰と画面遷移
タスクの現在の `Phase` が更新された場合、フロントエンドのダッシュボードはそのフェーズ情報を認識し、ユーザーがクリックした際に適切な画面（と必要なパラメータ）へルーティングできる仕組みを提供しなければならない。

#### Scenario: 複数ステップの翻訳プロジェクトの途中遷移
- **WHEN** ダッシュボードのタスク一覧で、ステータスが「ペルソナ抽出待機中(Phase: PersonaExtraction)」となっているタスクをクリックした
- **THEN** アプリケーションは翻訳プロジェクト詳細画面へ遷移し、自動的に「ペルソナパネル」のタブを開いて状態を復元する。

### Requirement: Wailsバインディングの拡張
TaskManager は従来の `GetActiveTasks`, `CancelTask` に加え、中断されたタスクを再開するための `ResumeTask` 等のバインディングを提供しなければならない。

### Requirement: Task は進捗値を決定して Progress へ報告しなければならない
`task` は段階進捗の `phase/current/total/message` を決定し、`progress` へ報告しなければならない。`total` は全 request 件数を分母として扱わなければならない。

#### Scenario: フェーズ遷移時に task が進捗を報告する
- **WHEN** Task が enqueue/dispatch/save/complete の各段階へ遷移する
- **THEN** `task` は対応する `phase` と進捗値を `progress` へ報告しなければならない

#### Scenario: Progress は task が渡した値をそのまま配信する
- **WHEN** `task` が `phase/current/total/message` を `progress` へ報告する
- **THEN** UI へ配信される進捗値は task が指定した値と一致しなければならない

### Requirement: Task 状態遷移は task イベントで UI へ通知しなければならない
`task` の状態遷移（`pending/running/completed/failed/canceled`）は `task` イベント（例: `task:updated`）として UI に通知しなければならない。状態遷移通知は `progress` 経路を介してはならない。

#### Scenario: task 状態が UI へ直接通知される
- **WHEN** task 状態が `running` から `completed` に変化する
- **THEN** UI は `task` イベントで状態変更を受け取り、ステータス表示を更新しなければならない

### Requirement: MasterPersona タスクは中断点を保持して再開可能でなければならない
`task` は MasterPersona のキュー投入後に再開コンテキスト（task_id, phase, resume_cursor）を保存し、再起動後およびキャンセル後の再開要求を受理しなければならない。再開時のモデル・パラメータは常に `config` から再読込しなければならない。

#### Scenario: 再起動後に再開できる
- **WHEN** MasterPersona 実行中にアプリが終了し、再起動後に ResumeTask が呼ばれる
- **THEN** `task` は保存済み再開コンテキストを読み出し、未完了フェーズから処理を再開しなければならない

#### Scenario: キャンセル後に再開できる
- **WHEN** MasterPersona タスクが cancel 済みで、ユーザーが再開を要求する
- **THEN** `task` は完了済み request を再実行せず、未完了 request のみを再開しなければならない
