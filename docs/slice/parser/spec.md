# パーサースライス アーキテクチャ
> Interface-First AIDD: Parser 垂直スライス

## 目的
TBD: ローダーモジュールに対する依存性注入の提供、および内部データ処理の隠蔽（カプセル化）。

## 要件

### 要件: Parser出力の独自DTO定義とグローバルドメイン依存の排除
**理由**: VSA（Vertical Slice Architecture）の原則に従い、システム全体で共有するデータモデル（`pkg/domain` 等）を排除し、各スライスが自身の入出力構造を独自定義する設計（Anti-Corruption Layer）へ移行するため。
**移行**: これまでパース結果を格納するために利用していた `pkg/domain` などの外部依存モデルを破棄し、本スライスの `contract.go` パッケージ内に独自定義した出力専用DTOを返すようにインターフェースおよび内部実装を修正する。

#### シナリオ: 独自定義DTOによるパース結果の返却
- **WHEN** Modファイル群のロードおよびパース処理が完了した場合
- **THEN** 外部パッケージのモデルに一切依存することなく、本スライス内部で定義された独自DTOに全データを格納して返却できること

### 要件: Parser スライス DI プロバイダー
Skyrim 抽出 JSON を扱う parser 機能は、`pkg/format/parser/skyrim` 配下の format adapter として実装されなければならない。システムは、Google Wire または同等の composition root を介して依存性注入プロバイダー関数を提供し、workflow からは `Parser` 契約として解決できなければならない。

#### シナリオ: DI 初期化
- **WHEN** アプリケーションが抽出データ読み込み用の parser を初期化する場合
- **THEN** 内部構造体を直接インスタンス化することなく、`pkg/format/parser/skyrim` の provider を通じて `Parser` 契約を解決しなければならない
- **AND** workflow は interface 名を変更せずに新配置の実装を利用できなければならない

### 要件: 内部プロセスのカプセル化
Skyrim parser における順次ロードと並列デコードの手順は、`pkg/format/parser/skyrim` の `Parser` 実装内にカプセル化され、workflow に公開されてはならない。workflow は最終的な `ExtractedData` またはエラーのみを受け取らなければならない。

#### シナリオ: ロードプロセスのオーケストレーション
- **WHEN** workflow が `LoadExtractedJSON` を呼び出す場合
- **THEN** `pkg/format/parser/skyrim` の実装はファイルの読み込み、デコード、構造化を内部で調整し、最終的な `ExtractedData` またはエラーのみを返さなければならない
- **AND** workflow は format 配下の内部処理手順へ依存してはならない

### 要件: 階層的コンテキストのための parser DTO 拡張
Parser スライスは、完全な翻訳コンテキストと将来のエクスポート処理に必要な階層的データ関係をキャプチャする、堅牢なデータ転送オブジェクト（DTO）を定義・入力しなければならない。

#### シナリオ: 階層的なクエスト項目の DTO への入力
- **WHEN** 抽出スクリプトから DTO にデータをロードする場合
- **THEN** Parser スライスは、`QuestStage` と `QuestObjective` の構造を拡張し、親クエストの `ID` と `EditorID` を保存できるようにする
- **AND** データロードプロセス中にこれらのフィールドに値を入力する

### 要件: データ抽出の前提条件と厳密なパース
Parser スライスは、入力される JSON ペイロード（`extractData.pas` によって生成される）に、インデックスの衝突とレスポンスの順序に関する修正が含まれていることを前提としなければならない。

#### シナリオ: 抽出された JSON の正しいパース
- **WHEN** Pascal 抽出スクリプトから JSON ペイロードをロードする場合
- **THEN** Parser スライスは、衝突を防ぐために `stage_index` と `log_index` を別のフィールドとしてパースする
- **AND** 正確なツリー探索と順序を保証するために、会話レスポンスの明示的な `order`（または `index`）フィールドをパースする

---

## ログ出力・テスト共通規約

> 本スライスは `standard_test_spec.md` と `log-guide.md` に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（`standard_test_spec.md` 参照）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（`log-guide.md` 参照）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（`log-guide.md` 参照）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（`log-guide.md` 参照）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（`log-guide.md` の AI デバッグ運用参照）でログと仕様書をAIに渡し修正させる。
