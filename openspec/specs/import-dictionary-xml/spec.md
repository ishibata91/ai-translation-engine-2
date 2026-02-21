# 辞書 XML のインポート (Import Dictionary XML)

## 目的
xTranslator 形式の XML ファイルをインポートし、用語辞書を構築する。

## 要件

### 要件: xTranslator XML のインポート
システムは、1 つ以上の xTranslator XML ファイルを読み込み、設定可能な許可リストに基づいて名詞レコードを抽出できなければならない (MUST)。

#### シナリオ: 複数のレコードタイプが混在する有効な XML のインポート
- **WHEN** ユーザーが `BOOK:FULL`, `NPC_:FULL`, `INFO` レコードを含む有効な xTranslator XML ファイルを提供した場合
- **AND** 設定で `BOOK:FULL` と `NPC_:FULL` が許可タイプとして指定されている場合
- **THEN** システムは `BOOK:FULL` と `NPC_:FULL` レコードのみを抽出する
- **AND** `INFO` レコードは無視する
- **AND** 抽出されたレコードを SQLite 辞書データベースに UPSERT（追加・更新）する

#### シナリオ: 非常にサイズの大きい XML ファイルの処理
- **WHEN** ユーザーが利用可能なメモリを超えるサイズの XML ファイルを提供した場合
- **THEN** システムはストリーミングアプローチ (`encoding/xml.Decoder`) を用いてファイルを解析し、メモリ不足 (OOM) エラーを発生させずに処理しなければならない

#### シナリオ: データベーススキーマが存在しない場合
- **WHEN** アプリケーションが起動し `DictionaryStore` を初期化する場合
- **THEN** システムは、インポートが実行される前に辞書テーブルが存在することを保証するため、自動的に `CREATE TABLE IF NOT EXISTS` 文を実行する

#### シナリオ: 重複レコードのインポート
- **WHEN** ユーザーが、データベースに既に存在する EDID を持つレコードを含む XML ファイルをインポートした場合
- **THEN** システムは UPSERT (`ON CONFLICT`) 操作を実行する
- **AND** 重複行を作成することなく、既存のレコードを新しい `Source` および `Dest` の値で更新する

