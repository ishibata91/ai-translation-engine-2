## MODIFIED Requirements

### Requirement: Wailsバインディングの拡張
Task 管理機能は、controller 層が従来の `GetActiveTasks`, `CancelTask`, `ResumeTask` などの Wails バインディングを提供しつつ、実際のユースケース進行は workflow 契約へ委譲しなければならない。Wails バインディングは task 状態管理の公開入口であるが、slice 実行順序や infrastructure 制御を直接実装してはならない。

#### Scenario: Wails バインディングは workflow を経由して再開する
- **WHEN** UI から `ResumeTask` が呼び出される
- **THEN** controller は workflow の再開契約を呼び出さなければならない
- **AND** controller 自身が queue や persona slice の具体的な制御ロジックを持ってはならない

#### Scenario: 既存の公開 API は維持される
- **WHEN** フロントエンドが既存の `GetActiveTasks`, `CancelTask`, `ResumeTask` を利用する
- **THEN** システムは同等の公開 API を継続提供しなければならない
- **AND** 内部の接続先変更だけを理由に UI 契約を破壊してはならない

### Requirement: Task は進捗値を決定して Progress へ報告しなければならない
task 管理機能は、workflow が段階進捗の `phase/current/total/message` を決定し、`progress` へ報告しなければならない。`total` は全 request 件数を分母として扱わなければならない。controller は進捗値を再計算してはならず、usecase slice は進捗通知の公開契約を直接持ってはならない。

#### Scenario: フェーズ遷移時に workflow が進捗を報告する
- **WHEN** workflow が enqueue/dispatch/save/complete の各段階へ遷移する
- **THEN** workflow は対応する `phase` と進捗値を `progress` へ報告しなければならない

#### Scenario: Progress は workflow が渡した値をそのまま配信する
- **WHEN** workflow が `phase/current/total/message` を `progress` へ報告する
- **THEN** UI へ配信される進捗値は workflow が指定した値と一致しなければならない

### Requirement: Task 状態遷移は task イベントで UI へ通知しなければならない
task の状態遷移（`pending/running/completed/failed/canceled`）は `task` イベント（例: `task:updated`）として UI に通知しなければならない。状態遷移通知は `progress` 経路を介してはならない。controller は当該イベントの公開窓口として振る舞い、workflow が決定した状態をそのまま外部へ通知しなければならない。

#### Scenario: task 状態が UI へ直接通知される
- **WHEN** workflow が task 状態を `running` から `completed` に変化させる
- **THEN** UI は `task` イベントで状態変更を受け取り、ステータス表示を更新しなければならない

#### Scenario: controller は状態遷移を独自解釈しない
- **WHEN** workflow が失敗またはキャンセル状態を通知する
- **THEN** controller は受け取った状態をそのまま UI へ通知しなければならない
- **AND** controller が別の状態名へ変換してはならない
