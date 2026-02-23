# 辞書DB作成機能 (Dictionary Slice) 仕様書

## 概要
xtranslator形式のXMLファイルから用語と翻訳データを読み込み、SQLiteベースの辞書DBへ登録する機能である。
当機能は Interface-First AIDD v2 アーキテクチャに則り、**完全な自律性を持つ Vertical Slice** として設計される。
AIDDにおける決定的なコード再生成の確実性を担保するため、あえてDRY原則（データ構造やDB操作の共通化）を捨て、**本Slice自身が「辞書テーブルのスキーマ定義」「DTO」「SQL発行・永続化ロジック」の全ての責務を負う。** 外部機能には一切依存せず、単一の明確なコンテキストとして自己完結する。

## 要件
1. **独立したUI**: ユーザーはWeb UI上から複数のxtranslator XMLファイルを指定し、一括でインポート処理を実行できる。
2. **XML解析**: `SSTXMLRessources > Content > String` 階層から `EDID`, `REC`, `Source`, `Dest` を抽出する。
3. **カプセル化された永続化**: プロセスマネージャーから `*sql.DB` などの**「DBのプーリング・接続管理のためだけのインフラモジュール」**のみをDIで受け取り、本Slice内の `DictionaryStore` が辞書テーブルに対するすべての操作（テーブル生成・INSERT/UPSERT等）を単独で完結させる。
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

---

## ログ出力・テスト共通規約

> 本スライスは `architecture.md` セクション 6（テスト戦略）・セクション 7（構造化ログ基盤）に準拠する。

### 実装時の義務

1.  **パラメタライズドテスト**: テストは Table-Driven Test で網羅的に行い、細粒度のユニットテストは作成しない（セクション 6.1）。
2.  **Entry/Exit ログ**: 全 Contract メソッドおよび主要内部関数で `slog.DebugContext(ctx, ...)` による入口・出口ログを出力する（セクション 6.2 ①）。
3.  **TraceID 伝播**: 公開メソッドは第一引数に `ctx context.Context` を受け取り、OpenTelemetry TraceID を全ログに自動付与する（セクション 7.3）。
4.  **ログファイル出力**: 実行単位ごとに `logs/{timestamp}_{slice_name}.jsonl` へ debug 全量を記録する（セクション 6.2 ③）。
5.  **AI デバッグプロンプト**: 障害時は定型プロンプト（セクション 6.2 ④）でログと仕様書をAIに渡し修正させる。
