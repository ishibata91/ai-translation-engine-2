# 共有ハンドオフ

slice 間で受け渡す共有データ、中間成果物、resume 用状態の保存・検索境界を定義し、workflow 主導の handoff を成立させる。

## Requirements

### Requirement: slice間共有データはartifactへ配置しなければならない
システムは、複数 slice から参照される共有データ、中間成果物、resume 用状態を `artifact` に配置しなければならない。ある slice の内部保存物を、後続 slice が直接参照してはならない。

#### Scenario: 後続sliceが前段sliceの成果物を利用する
- **WHEN** ある slice が後続 slice へ渡す共有データを保存する必要がある
- **THEN** 共有データは `artifact` の保存・検索契約へ格納されなければならない
- **AND** 後続 slice は前段 slice の内部 DB や内部 DTO を直接参照してはならない

### Requirement: slice間受け渡しはworkflowがartifact境界で束ねなければならない
システムは、slice 間の受け渡しを `workflow` が `artifact` 識別子、検索条件、batch / page / cursor を用いて束ねなければならない。

#### Scenario: slice間連携を実装する
- **WHEN** 開発者が parser の出力を persona や translator へ受け渡す処理を実装する
- **THEN** `workflow` が artifact 上の識別子または検索条件を束ねて後続 slice を呼び出さなければならない
- **AND** artifact は保存・検索以外の業務判断を持ってはならない

### Requirement: 翻訳フローのパース済みデータは task_id 基点の構造化 artifact テーブルへ保存されなければならない
システムは、翻訳フローのデータロードで生成されたパース済みデータを、`task_id` を親キーとする構造化テーブル群へ保存しなければならない。システムは `artifact_records` に parser のレスポンス全体を保存してはならない。

#### Scenario: task 単位でパース済みデータを保存する
- **WHEN** workflow が翻訳フローで複数ファイルのパース結果を受け取る
- **THEN** システムは `task_id` 配下にファイル親テーブルと section 別テーブルへデータを保存しなければならない
- **AND** システムは翻訳フロー専用 task を新設せず、既存の翻訳プロジェクト task の `task_id` をそのまま使わなければならない

#### Scenario: parser の入れ子構造を保持して保存する
- **WHEN** システムが `DialogueGroup` と `Quest` を保存する
- **THEN** システムは `DialogueGroup -> DialogueResponse` と `Quest -> QuestStage / QuestObjective` の親子関係を別テーブルで保持しなければならない
- **AND** システムは入れ子構造を失う形で単一 JSON blob に変換して保存してはならない

### Requirement: 翻訳フローの artifact ファイル親テーブルは preview 合計件数を保存時に確定しなければならない
システムは、翻訳フローのファイル親テーブルに preview 用合計件数 `preview_row_count` を保持し、保存時にその値を確定しなければならない。システムは section 別の冗長な件数列を必須としない。

#### Scenario: ファイル保存時に preview 行数を確定する
- **WHEN** システムが 1 ファイル分の parser 出力を artifact へ保存する
- **THEN** システムは preview 対象となる全 section の行数を集計して `preview_row_count` に保存しなければならない
- **AND** システムは初期表示のために section 別件数を必須で保持してはならない

### Requirement: 翻訳フローと後続フェーズは task_id を用いて構造化 artifact テーブルからパース済みデータを取得しなければならない
システムは、翻訳フローのロード済みデータ表示および後続フェーズへの受け渡しにおいて、`task_id` を用いて構造化 artifact テーブル群を検索しなければならない。システムは parser の内部 DB や内部 DTO を UI または後続 slice へ直接公開してはならない。

#### Scenario: データロード画面を再表示する
- **WHEN** ユーザーが同じ翻訳タスクを再度開き、既存のパース済みデータを表示する
- **THEN** システムは `task_id` を用いて artifact テーブル群からファイル一覧と preview 行を復元しなければならない
- **AND** システムは parser の内部保存先を直接参照して復元してはならない

#### Scenario: 後続フェーズへロード結果を受け渡す
- **WHEN** workflow がデータロード完了後に用語フェーズ以降へ進行する
- **THEN** システムは `task_id` と file / section 識別子を用いて後続フェーズへパース済みデータを参照させなければならない
- **AND** システムは parser の内部 DTO を後続フェーズへ直接渡してはならない
