# Dictionary Builder Slice Design

## Context

v2.0のコンテキストベース翻訳を正確に行うためには、翻訳前にゲーム内の固有名詞（NPCやアイテム等）の対訳辞書を構築しておく必要がある。この辞書データは、既存のxTranslator形式のXMLファイルから生成する。
本機能（Dictionary Builder Slice）は、XMLファイルを入力とし、指定された対象レコードのみを抽出してSQLiteベースの辞書DBへ保存する責務を持つ。

## Goals / Non-Goals

**Goals:**
- xTranslator形式のXMLファイルから、指定された名詞レコードを抽出する。
- 抽出条件（対象とするRECの種類）を、スライス内にハードコードせず、外部（Config等）からDIで注入可能にする。
- 抽出したデータをSQLiteベースの辞書DBに対してUPSERTで確実に永続化する。
- DBのテーブル作成からレコード更新までの永続化ロジックを本スライス内で単独完結させる（DRY原則の無効化、Vertical Slice Architectureの徹底）。
- 巨大なXMLファイルに対応するため、ストリーミングパース(`encoding/xml.Decoder`)を採用してメモリ消費を抑える。

**Non-Goals:**
- xTranslator XMLの完全な解析や、名詞以外の汎用レコード（会話文など）の抽出・翻訳処理。
- 他ドメイン（例: Translation Engine Sliceなど）とのDBスキーマやロジックの共有。

## Decisions

### Decision 1: ストリーミングパースの採用
巨大なXMLファイルをメモリに一括ロードすることを避けるため、Go標準の `encoding/xml.Decoder` を用いたストリーミングパースを採用する。`SSTXMLRessources > Content > String` 階層の要素を逐次読み込み、処理する。

### Decision 2: 抽出フィルタの外部注入 (Config化)
抽出対象となるレコードタイプ（`BOOK:FULL`, `NPC_:FULL`等）は、本モジュール内にハードコードせず、DIコンテナ（Google Wire）を通じて注入する。これにより、将来的な仕様変更や、別機能（例: ユーザー独自辞書）での再利用を容易にする。

### Decision 3: 自己完結型の永続化 (DictionaryStore)
データベースアクセスにおいて、共通のORMや汎用DAOは使用しない。本スライス専用の `DictionaryStore` を実装し、初期化時のテーブル生成 (`CREATE TABLE IF NOT EXISTS`) から、データ挿入・更新 (`INSERT ... ON CONFLICT`) に至るまでのSQLを直接記述する。外部依存は `*sql.DB` インスタンスのみとする。

### Decision 4: トランザクショナルな一括処理
多数のレコードを効率的にDBへ保存するため、XMLファイル1つのパースおよび永続化処理を1つのトランザクション内で実行する、もしくは一定件数ごとのバッチINSERTを行うことで、I/Oパフォーマンスを最適化する。

## Risks / Trade-offs

- **DRY原則の放棄**: 辞書DBのスキーマ定義やSQL操作をこのスライス内に持つため、他のスライスで同様のテーブルを参照する場合、コードの重複が発生する可能性がある。しかし、これはVertical Slice Architectureにおける「コンテキスト境界の自律性」を優先するため、許容されるトレードオフである。
- **メモリ消費 vs パフォーマンス**: ストリーミングパースは省メモリだが、DOMパースに比べて処理速度が若干劣る可能性がある。しかし、巨大なXMLを安定して処理する要件を満たすためには不可欠な選択である。
