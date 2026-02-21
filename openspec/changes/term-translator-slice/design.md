# Design: term-translator-slice

## Context
本機能（Term Translator Slice）は、Skyrim Mod自動翻訳エンジン「translate_with_local_ai.py (v2.0)」の 2-Pass System における **Pass 1: 用語翻訳 (Term Translation)** の中核を担う。
抽出されたModデータ(`ExtractedData`)からNPC名やアイテム名などの固有名詞を抽出し、事前構築済みの辞書DB（Dictionary Builder Sliceによる出力）を参照しながら、LLMを用いて翻訳を行う。
また、翻訳結果はMod専用のSQLiteデータベース（Mod用語DB）に保存され、後続の Pass 2（本文翻訳）でコンテキストとして参照される。
アーキテクチャとしては、Interface-First AIDD v2 に則り、他機能から独立した自律的な Vertical Slice として `pkg/term_translator` パッケージ内に実装する。

## Goals / Non-Goals

**Goals:**
- `ExtractedData` から指定されたレコードタイプ（設定で動的変更可能）の名詞類を抽出し、翻訳リクエストを生成する。
- 既存の辞書DB（公式DLC辞書等）から、対象テキストに適合する参照用語を貪欲最長一致やNPC部分一致などの高度な検索手法を用いて取得する。
- 英語の複数形や所有格に対応するため、Snowball Stemmerを用いたステミング検索を実装する。
- LLMクライアントを呼び出し並行翻訳を実行し、翻訳済みの結果をModごとに個別の用語SQLiteDBに保存（FTS5対応）する。
- DI (Google Wire) を用いてインターフェース依存関係を解決し、網羅的パラメタライズドテストと `slog` + OpenTelemetry による構造化デバッグログをサポートする。

**Non-Goals:**
- xTranslator XMLからの辞書DB自体の構築（別スライスの責務）。
- 会話ツリーの解析や会話本文（`INFO` レコード等）の翻訳（別スライスの責務）。
- SQLite以外のDBへの永続化対応。
- このスライス外のデータモデルとのDRY（コード共有）。モデルやDB永続化コードはスライス内に完全にカプセル化（WETアーキテクチャ）する。

## Decisions

- **Vertical Slice Architecture の徹底**: `ModTermStore` などのDBアクセス層や、`TermTranslationRequest` などのDTOは全て `term_translator` パッケージ内に専用に定義し、外部と共有しない。
- **SQLite + FTS5 のカプセル化**: Mod用語DBの保存において、`mod_terms` テーブルと `mod_terms_fts` (FTS5) 仮想テーブルをスライス内部の初期化ロジックで構築し、別スライスへは完成したDBのパスを渡す形とする。
- **Snowball Stemmer による検索強化**: `github.com/kljensen/snowball` を導入し、辞書検索時のキーワードマッチング率を向上させる。
- **NPC名のペア翻訳機能**: `NPC_:FULL` と `NPC_:SHRT` は個別に翻訳すると不整合が起きるため、`TermRequestBuilder` でペアとして結合し、1つのプロンプトで同時に翻訳要求を出して分解保存する。
- **並行処理と進捗通知**: Goの `sync` パッケージや `errgroup` を活用して複数のLLMリクエストを並行実行し、`ProgressNotifier` インターフェースを通じて進行状況を逐次プロセスマネージャーに通知する。
- **テスト設計**: 細粒度のユニットテストは排除し、スライス全体の入力（`ExtractedData`）から出力（用語DBのレコード）までを検証する網羅的パラメタライズドテストを採用する。また、DBには `:memory:` を用いる。

## Risks / Trade-offs

- **[Risk] LLMプロンプトのコンテキスト制限制限**: 大量の参照用語がコンテキストに含まれた場合、LLMのトークン制限を超過する可能性がある。
  - **Mitigation**: 貪欲最長一致アルゴリズムによって不要な部分一致ノイズをフィルタリングし、本当に必要な用語のみを `reference_terms` に含めるよう制御する。
- **[Risk] WETな設計によるボイラープレート増加**: 似たようなDBアクセスクラスを複数スライスで定義する必要がある。
  - **Mitigation**: 意図的な設計方針（Interface-First AIDD）であり、AIがコンテキストを迷わず自己完結してコード生成できるようになるメリットの方が大きいため許容する。
