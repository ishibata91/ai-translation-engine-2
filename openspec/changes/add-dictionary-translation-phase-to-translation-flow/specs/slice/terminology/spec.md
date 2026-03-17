## MODIFIED Requirements

### Requirement: 2フェーズモデル（提案/保存モデル） (Propose/Save Model)
**Reason**: バッチAPI等の長時間待機を伴うLLM通信に対応するため、スライスの責務を「プロンプト生成」と「結果保存」の2段階に分離し、通信制御をインフラ層（JobQueue/Pipeline）へ委譲する。

#### Scenario: 用語翻訳ジョブの提案 (Phase 1: Propose)
- **WHEN** プロセスマネージャーまたは translation flow workflow から、task ID などの artifact 境界情報と terminology 用 request/prompt 設定を受け取った
- **THEN** slice は必要な入力を artifact から読み出して `TermTranslatorInput` を組み立てなければならない
- **AND** 内部の辞書検索を行い、operator 指定 prompt と request 設定を反映した LLM プロンプトを構築しなければならない
- **AND** 構築されたプロンプトの配列 `[]llm.Request` を返す（自身ではLLMクライアントを呼び出さない）
- **AND** `specs/log-guide.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

#### Scenario: 用語翻訳結果の保存 (Phase 2: Save)
- **WHEN** プロセスマネージャーまたは translation flow workflow から、自身の生成したリクエストに対応する `[]llm.Response` が渡された
- **THEN** 各レスポンスから `TL: |にほんご|` フォーマットを抽出する
- **AND** パースに成功したテキストを terminology DB の該当テーブルへ UPSERT する
- **AND** 保存状態を status カラムへ更新しなければならない
- **AND** `specs/log-guide.md` に従い、関数の開始・終了ログを TraceID 付きで出力する

#### Scenario: パース失敗・エラーレスポンスの処理
- **WHEN** レスポンスがエラー（`Success == false`）である、または `TL: |...|` 形式が含まれていない
- **THEN** 該当するレコードの更新を安全にスキップし、エラー詳細を構造化ログとして記録する
- **AND** 保存状態の失敗を status カラムへ反映しなければならない
- **AND** 処理全体を中断せず、他の正常なレスポンスの処理を続行する

### Requirement: terminology は artifact 境界を通じて translation flow から起動されなければならない
terminology slice は、translation flow から task ID や file 識別子などの artifact 境界情報を受け取り、slice 内で必要な入力を組み立てて Pass 1 の単語翻訳を実行できなければならない。terminology slice は translation flow の内部 DTO や UI state に依存してはならない。

#### Scenario: translation flow が terminology を起動する
- **WHEN** translation flow workflow が task ID を terminology slice に渡す
- **THEN** terminology slice は artifact から必要データを読み出して既存の提案/保存モデルで処理できなければならない
- **AND** translation flow 専用の内部型を terminology slice に import してはならない

### Requirement: terminology は operator 指定の request/prompt 設定を prompt 生成へ反映しなければならない
システムは、translation flow UI で確定された terminology 用の request 設定と prompt 設定を terminology slice が受け取り、`PreparePrompts` の生成結果へ反映しなければならない。slice は frontend の config 保存形式に直接依存せず、専用 DTO で受け取らなければならない。

#### Scenario: prompt と request 設定を反映して提案を生成する
- **WHEN** workflow が terminology 用設定 DTO を指定して `PreparePrompts` を呼ぶ
- **THEN** terminology slice は system prompt、user prompt、model request 設定を用いて LLM リクエストを生成しなければならない
- **AND** frontend の保存キー文字列を slice 内で直接解釈してはならない

### Requirement: terminology は単一の terminology.db にファイル名ベースの mod テーブルを作成して保存しなければならない
システムは、terminology の保存先を mod ごとの別 DB ファイルではなく単一の `terminology.db` とし、ファイル名を基準に決定した mod ごとのテーブル名で翻訳結果を管理しなければならない。本文翻訳など後続フェーズは同一 DB 内の対象 mod テーブルを参照できなければならない。

#### Scenario: 単語翻訳結果を保存する
- **WHEN** terminology slice が特定 mod ファイルに対する単語翻訳結果を保存する
- **THEN** システムはファイル名から決定した当該 mod 用テーブルへ結果を保存しなければならない
- **AND** 別 mod の翻訳結果を別 DB ファイルへ分散保存してはならない

#### Scenario: 後続フェーズが用語辞書を参照する
- **WHEN** 本文翻訳などの後続フェーズが用語辞書を参照する
- **THEN** システムは `terminology.db` 内の対象 mod テーブルを参照できなければならない
- **AND** task metadata に DB パスを別途持たなくても参照先を決定できなければならない

### Requirement: terminology は phase 状態を status カラムで保持しなければならない
システムは、translation flow から起動される単語翻訳 phase の状態を task metadata の JSON ではなく terminology 側の status カラムで保持しなければならない。workflow と UI はこの status を正本として参照しなければならない。

#### Scenario: 単語翻訳 phase が進行中である
- **WHEN** terminology slice が提案または保存フェーズを実行中である
- **THEN** システムは terminology 側の status カラムを進行中状態へ更新しなければならない
- **AND** task metadata の JSON 値だけで進行状態を保持してはならない

#### Scenario: 単語翻訳 phase が完了または失敗した
- **WHEN** terminology slice の保存フェーズが完了または失敗した
- **THEN** システムは terminology 側の status カラムを完了または失敗状態へ更新しなければならない
- **AND** workflow はこの status をもとに phase 完了判定を復元できなければならない
