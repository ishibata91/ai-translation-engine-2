# 辞書DB作成機能 (Dictionary Builder Slice) 仕様書

## 概要
xtranslator形式のXMLファイルから用語と翻訳データを読み込み、構造化データとして抽出したのち、プロセスマネージャー（呼び出し元）へ返却する機能である。
当機能は Interface-First AIDD v2 アーキテクチャに則り、純粋なデータ変換を担う自律した Vertical Slice として設計される。
**データの永続化（データベースへの登録）は本Slice内では行わず、本Sliceから返却されたデータを受け取ったプロセスマネージャーが共通の永続化層を通じて実行する。**

## 要件
1. **独立したUI**: ユーザーはWeb UI上から複数のxtranslator XMLファイルを指定し、一括でインポート処理を実行できる。
2. **XML解析**: `SSTXMLRessources > Content > String` 階層から `EDID`, `REC`, `Source`, `Dest` を抽出し、`Params > Addon` の情報を付加して `[]DictTerm` として返却する。
3. **プロセスマネージャーによる永続化**: 本Slice（XML解析層）から返却された用語データを、プロセスマネージャーが全社/全体共通の永続化層（Repository等）を通じてSQLiteベースの辞書DBへ保存する機能を調整・統合する。
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
- [シーケンス図](./dictionary_builder_sequence_diagram.md)
- [クラス図](./dictionary_builder_class_diagram.md)
- [テスト設計](./dictionary_builder_test_spec.md)
