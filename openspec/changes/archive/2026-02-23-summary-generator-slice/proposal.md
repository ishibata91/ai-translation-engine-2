# Proposal: summary-generator-slice

## Motivation
現行のPython版ツール（v1.x）では、クエスト要約と会話要約をサイドカーJSONファイルにキャッシュしているため、信頼性・検索性が低く、並列処理時の競合リスクがある。
Go v2では要件定義書 §3.2「クエスト要約」に基づき、以下の課題を解決する独立スライスとして実装する。

1. **キャッシュの信頼性向上**: サイドカーJSONからSQLiteへ移行し、ACID保証・インデックス検索を実現する。
2. **会話コンテキストの品質向上**: `PreviousID` チェーンを辿った会話フロー全体をLLMで要約することで、Pass 2翻訳時の文脈精度を向上させる。
3. **クエスト要約の累積的生成**: ステージを時系列順に処理し、「これまでのあらすじ」を累積的にビルドする仕組みをSQLiteキャッシュと組み合わせて実現する。
4. **新アーキテクチャへの適合**: Vertical Slice Architecture・Consumer-Driven Contracts DTO Style・Interface-First AIDD v2 に準拠した完全自律スライスとして設計する。

## Capabilities

### New Capabilities
- `SummaryGeneratorSlice`: 以下の責務を持つ独立したスライス。
  - **2フェーズモデル（Propose/Save）の提供**:
    - `ProposeJobs`: 会話/クエストのデータを受け取り、キャッシュ判定を行い、未キャッシュ分をLLMリクエスト（ジョブ）として提案する。
    - `SaveResults`: LLMの実行結果を受け取り、SQLiteキャッシュに永続化する。
  - **会話要約の生成**: `DialogueGroup` 単位で、`PreviousID` チェーンを含む会話フローをプロンプト化する。
  - **クエスト要約の生成**: `Quest.Stages` をIndex昇順で処理し、過去ステージの記述を累積したプロンプトを構築する。
  - **キャッシュヒット判定**: レコードIDとSHA-256ハッシュに基づくキャッシュキーでSQLiteを検索する。
  - **SQLite永続化（ソースファイル単位）**: ソースプラグインごとに独立した `_summary_cache.db` ファイルを作成し、`summaries` テーブルへUPSERTで保存する。
  - **独立性の確保**: 本スライス専用の入力DTO（`SummaryGeneratorInput`）を `contract.go` 内に定義し、他スライスのデータ構造に依存しない。

### Modified Capabilities

## Impact
- **アーキテクチャ**: 新規パッケージ `pkg/summary_gen` が追加される。完全な Vertical Slice として実装され、外部DTOには依存しない。
- **データベース**: ソースプラグインごとの SQLite ファイル（`{plugin_name}_summary_cache.db`）に `summaries` テーブルを定義・作成する。テーブル作成はスライス内にカプセル化される。
- **後続処理 (Pass 2)**: Pass 2 翻訳スライスが `DialogueGroup.ID` または `Quest.ID` をキーに `summaries` テーブルから要約を取得し、LLMプロンプトのコンテキストに挿入できるようになる。
- **依存性注入**: `google/wire` を用いてWire Providerを定義し、`*sql.DB` および `LLMClient` をDIで受け取る。
