# 辞書DB作成機能 (Dictionary Slice) 仕様書

## 概要
xtranslator形式のXMLファイルから用語と翻訳データを読み込み、SQLiteベースの辞書DBへ登録する機能である。
当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則（データ構造やDB操作の共通化）を捨て、**本Slice自身が「辞書テーブルのスキーマ定義」「DTO」「SQL発行・永続化ロジック」の全ての責務を負う。** 外部機能には一切依存せず、単一の明確なコンテキストとして自己完結する。

## 要件
1. **独立したUI**: ユーザーはWeb UI上から複数のxtranslator XMLファイルを指定し、一括でインポート処理を実行できる。
2. **XML解析**: `SSTXMLRessources > Content > String` 階層から `EDID`, `REC`, `Source`, `Dest` を抽出する。
3. **カプセル化された永続化**: プロセスマネージャーから `*sql.DB` などの**「DBのプーリング・接続管理のためだけのインフラモジュール」**のみをDIで受け取り、本Slice内の `DictionaryStore` が2テーブル構成（`dlc_sources` / `dlc_dictionary_entries`）スキーマを使用し、辞書テーブルに対するすべての操作（ソース管理、エントリに関するCRUD等）を単独で完結させる。
4. **名詞の抽出要件 (フィルタリング)**: 本機能は「用語辞書」であるため、XMLに含まれるすべてのテキストではなく、**対象とする特定のレコード（名詞類）のみ**を抽出して永続化する。対象リストに含まれないRECはすべて無視（パーススキップ）する。
   - **対象とするREC（許可リスト）**:
     - `BOOK:FULL`: 本のアイテム名
     - `NPC_:FULL`: NPC名
     - `NPC_:SHRT`: NPCの短い名前
     - `ARMO:FULL`: 防具名
     - `WEAP:FULL`: 武器名
     - `LCTN:FULL`: ロケーション名
     - `CELL:FULL`: セル名
     - `CONT:FULL`: コンテナ名
     - `MISC:FULL`: その他アイテム名
     - `ALCH:FULL`: 食料・ポーション等の錬金術アイテム名
     - `FURN:FULL`: 家具名
     - `DOOR:FULL`: ドア・扉名
     - `RACE:FULL`: 種族名
     - `INGR:FULL`: 錬金素材名
     - `FLOR:FULL`: 植物等の収穫物名
     - `SHOU:FULL`: シャウト名
   - **共通設定（Config）化による再利用**: 上記の「抽出対象のREC定義リスト」は、本Slice（XMLパーサー）の内部にハードコードするのではなく、**システム共通のConfig（設定情報定義）として切り出して定義し、DI等で注入**する。これにより、将来的にMod由来の辞書データを処理する別のコンテキストなど、コンテキストを超えて同一の定義・フィルタリングルールを再利用できるように設計する。
5. **ライブラリの選定**: 
   - XML解析: Go標準の `encoding/xml`（`xml.Decoder`を用いたストリーミングパース）
   - DBアクセス (PM側): `github.com/mattn/go-sqlite3` または標準 `database/sql`
   - 依存性注入: `github.com/google/wire`

## 関連ドキュメント
- [シーケンス図](./dictionary_sequence_diagram.md)
- [クラス図](./dictionary_class_diagram.md)
- [テスト設計](./dictionary_test_spec.md)

### Requirement: 辞書構築（Dictionary Builder）画面へのアクセス
ユーザーはシステム全体のサイドバーまたはナビゲーションを通じて、新規に実装される辞書構築（Dictionary Builder）画面にアクセスできなければならない（SHALL）。

#### Scenario: ナビゲーションからの遷移
- **WHEN** ユーザーがサイドバーの「Dictionary Builder」に該当するアイコンをクリックしたとき
- **THEN** メインコンテンツ領域に辞書構築の画面が表示されること

### Requirement: 辞書ソース一覧のテーブル表示
Dictionary Builder画面において、辞書の親データセット（辞書ソース一覧）が、TanStack Tableを用いたDataTable形式で一覧表示されなければならない（SHALL）。

#### Scenario: リストの描画
- **WHEN** 辞書構築画面にアクセスしたとき
- **THEN** 辞書ソースリストが適切にテーブル行として表示されていること

### Requirement: 辞書詳細表示用の連携UI（DetailPane）
テーブル内の特定の辞書を選択した際、コンポーネント `DetailPane.tsx` と連携して、その辞書の詳細（種別、登録数、プレビューなど）を表示するペインが開かなければならない（SHALL）。

#### Scenario: 詳細のトグル操作
- **WHEN** リスト内の特定の辞書行、または選択ボタンをクリックしたとき
- **THEN** DetailPaneがせり上がり、詳細情報が閲覧可能になること
- **THEN** 閉じるアクションを行うと非表示に戻ること

### Requirement: GridEditorによる辞書エントリー編集UI
詳細画面または一画面遷移を介して、辞書そのもののエントリー群（単語、訳語などのフィールド）を、インライン編集可能な表である `GridEditor` 上で管理できなければならない（SHALL）。

#### Scenario: GridEditorの操作
- **WHEN** ユーザーが辞書エントリー領域（GridEditor）にアクセスしたとき
- **THEN** 行の追加ボタン、および各セルのインライン編集が可能で、修正後は保存等のアクションUIが配置されていること

### Requirement: 複数ファイルのアップロードと選択状態の保持
ユーザーはインポート用の辞書ファイル（SSTXMLなど）を複数同時に選択・管理できなければならない（SHALL）。
Web標準の `<input type="file">` による絶対パス取得不可の制約を回避するため、ファイルの選択にはWailsのネイティブファイルダイアログ（`runtime.OpenMultipleFilesDialog`）をバックエンドから呼び出して利用しなければならない（SHALL）。

#### Scenario: ファイルの追加と重複排除
- **WHEN** ユーザーがファイル選択ボタンからファイルを追加したとき
- **THEN** 以前に選択されたファイル群は失われずに追加（スタック）されること
- **AND** すでに選択されているファイルと同じ名前のファイルは重複して登録されないこと
- **AND** 選択されたファイルのリストが表示され、個別にリストから除外できること

### Requirement: 辞書構築の実行制御と進捗表示
辞書インポートを実行する「辞書構築を開始」ボタンがインポートパネル設定下部に配置されなければならない（SHALL）。
フロントエンドの進捗表示は、バックエンドから発火される `dictionary:import_progress` イベントを購読し、通知される `CorrelationID` とメッセージに基づいて動的に描画・管理されなければならない（SHALL）。

#### Scenario: 構築の実行とUIロック
- **WHEN** ファイルが1つ以上選択されているとき
- **THEN** 「辞書構築を開始」ボタンが活性化し実行可能となること
- **WHEN** 「辞書構築を開始」ボタンが押下されたとき
- **THEN** 選択したファイルごとの個別の進捗プログレスバーが表示されること（押下前は非表示であること）
- **AND** ファイル選択ボタンが非活性化されること
- **AND** 既存の辞書ソース一覧（DataTable）上にローディング画面（オーバーレイ表示）がかぶさり、操作不能となること

### Requirement: データバインディングとリスト操作のUI要件
フロントエンド・バックエンド間の通信（WailsによるJSONシリアライズ）におけるプロパティ名（スネークケースやキャメルケース）の揺れを吸収するため、DTOをフロントエンド側に取り込む際は複数のキーパターンでフォールバックマッピングされなければならない（SHALL）。
また、各ソースデータの「削除」操作は、一覧行ごとのアクションとして実装し、確認モーダルを経由して個別の削除対象のバックエンド処理（`DictDeleteSource`）を呼び出すこと（SHALL）。

### Requirement: 最新DBスキーマの準拠 (dlc_sources / dlc_dictionary_entries)
`DictionaryStore` は、永続化レイヤーとして `dictionary.db` を使用し、`specs/database_erd.md` で定義されたテーブルを実装しなければならない。

#### シナリオ: スキーマ移行
- **WHEN** スライスが初期化されるとき
- **THEN** `dictionary.db` 内に `dlc_sources` および `dlc_dictionary_entries` テーブルが存在することを確認する。
- **AND** レガシーな `dictionary_entries` テーブルは無視するか、安全に削除すること。

### Requirement: 辞書ソースの CRUD
バックエンドは辞書ソースを管理するためのメソッドを提供しなければならない。

#### シナリオ: ソースの一覧表示
- **WHEN** `GetSources` が呼び出されたとき
- **THEN** メタデータ（id, file_name, status, entry_count 等）を含む `dlc_sources` の全レコードを返さなければならない。

#### シナリオ: ソースの削除
- **WHEN** `DeleteSource(id)` が呼び出されたとき
- **THEN** `dlc_sources` からそのソースレコードを削除しなければならない。
- **AND** `dlc_dictionary_entries` に関連付けられているすべてのエントリをカスケード削除しなければならない。

### Requirement: 辞書エントリの CRUD (GridEditor サポート)
UI でのインライン編集を可能にするため、バックエンドは個別のエントリ操作をサポートしなければならない。

#### シナリオ: ソースに紐づくエントリの取得
- **WHEN** `GetEntriesBySourceID(sourceID)` が呼び出されたとき
- **THEN** そのソースに関連付けられたすべての `dlc_dictionary_entries` を返さなければならない。

#### シナリオ: エントリの更新
- **WHEN** `UpdateEntry(term)` が呼び出されたとき
- **THEN** `dlc_dictionary_entries` 内の特定の ID に対して `source_text` または `dest_text` を更新しなければならない。

#### シナリオ: エントリの削除
- **WHEN** `DeleteEntry(id)` が呼び出されたとき
- **THEN** `dlc_dictionary_entries` から特定のエントリを削除しなければならない。

### Requirement: Wails サービスバインディング (DictionaryService)
内部のスライスロジックを Wails フロントエンドに橋渡しするための新しい `DictionaryService` を実装しなければならない。

#### シナリオ: フロントエンド連携
- **WHEN** Wails アプリが起動するとき
- **THEN** `DictionaryService` がバインディング用として登録されていなければならない。
- **AND** 上記で定義されたすべての CRUD メソッドが、タスク/UI レイヤーからアクセス可能でなければならない。

### Requirement: 進捗報告とメタデータを伴うインポート処理
`DictionaryImporter` は、メタデータを登録し、進捗を通知するよう調整されなければならない。

#### シナリオ: ファイルのインポート
- **WHEN** ファイルインポートが開始されたとき
- **THEN** `dlc_sources` レコードを `status: "IMPORTING"` で作成しなければならない。
- **AND** XML のトークンがパースされ、バッチ単位で保存される際、`pkg/infrastructure/progress`（または同等の通知機構）を介して進捗状況を送信しなければならない。
- **AND** 完了時に、`status` は `"COMPLETED"` になり、`entry_count` は実際にインポートされたレコード数に更新されなければならない。

---

## ログ出力・テスト共通規約

> 本スライスは `architecture.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
