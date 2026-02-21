## Why

v2.0のコンテキスト解析および翻訳を行うためには、固有名詞（NPC名、アイテム名、場所名など）があらかじめ正確に翻訳・統一されている必要があります。そのため、既存のxTranslator用XML辞書から必要な名詞レコードのみを抽出し、SQLiteデータベースに永続化する「辞書DB作成機能（Dictionary Builder Slice）」が必要です。

## What Changes

- xTranslator形式のXMLを読み込み、指定された対象レコード（`BOOK:FULL`, `NPC_:FULL`等）のみをストリーミングパースする処理の実装。
- 抽出条件（許可リスト）を本スライスにハードコードせず、システム共通のConfigからDIで注入可能にする仕組みの導入。
- 抽出したデータを辞書DBへ確実に保存（UPSERT）する永続化ロジック（`DictionaryStore`）の実装。
- この機能をその他のドメインから完全に独立させたVertical Sliceとして構築。

## Capabilities

### New Capabilities
- `import-dictionary-xml`: Web UIなどから複数のxTranslator形式のXMLファイルを受け取り、名詞データを抽出して用語辞書DB（SQLite）へ一括でインポートする機能。

## Impact

- `pkg/dictionary_builder_slice/`: 辞書構築機能のVertical Sliceとして、パーサー、DTO、ストアモデル等が新規追加されます。
