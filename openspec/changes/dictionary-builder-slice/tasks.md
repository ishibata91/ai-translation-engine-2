# Implementation Tasks

## 1. Core Config & DI
- [x] 1.1 `pkg/dictionary_builder/config.go` を作成し、抽出対象のRECタイプ（`BOOK:FULL`, `NPC_:FULL`等）の許可リストを定義するConfig構造体を実装する
- [x] 1.2 `pkg/dictionary_builder/provider.go` を作成し、依存関係（Config、`*sql.DB`）を注入するための Google Wire プロバイダ関数を定義する

## 2. DTO & Models
- [x] 2.1 `pkg/dictionary_builder/dto.go` を作成し、XMLから抽出するデータ（EDID, REC, Source, Dest）を保持するDTO構造体を定義する

## 3. Storage Layer (`DictionaryStore`)
- [x] 3.1 `pkg/dictionary_builder/store.go` を作成し、`DictionaryStore` インターフェースと実装を定義する
- [x] 3.2 辞書テーブル作成ロジック（`CREATE TABLE IF NOT EXISTS dictionary_entries (...)`）を実装する
- [x] 3.3 レコードのUPSERTロジック（`INSERT ... ON CONFLICT(edid) DO UPDATE SET ...`）を実装する

## 4. Parser & Importer Logic
- [x] 4.1 `pkg/dictionary_builder/importer.go` を作成する
- [x] 4.2 `encoding/xml.Decoder` を用いて、指定されたファイルパスのXMLの `SSTXMLRessources > Content > String` 階層逐次読み込む処理を実装する
- [x] 4.3 読み込んだ要素の `REC` 属性がConfigの許可リストに含まれるかチェックするフィルタリング処理を実装する
- [x] 4.4 許可された要素を抽出し、`DictionaryStore` を用いてDBへ保存する（トランザクションまたはバッチ処理を考慮）ロジックを実装する

## 5. Entrypoint & Integration
- [x] 5.1 `cmd/dictionary_builder/main.go` または適切なエントリーポイントコマンドを作成し、コマンドライン引数でXMLファイルパスを受け取るようにする
- [x] 5.2 コマンド内で Wire を用いて依存関係を解決し、インポート処理を呼び出す
- [x] 5.3 （必要に応じて）`wire.go` を更新して Wire provider の登録を行う

## 6. Verification
- [x] 6.1 `pkg/dictionary_builder/test/importer_test.go` を作成（あるいは実装中のログ等で動作確認）し、ダミーのxTranslator XMLを用いて正常に許可リストの要素のみがDBに登録されるか確認する
- [x] 6.2 OOM（メモリ超過）が発生しないようにストリーミングパースの挙動を確認する
- [x] 6.3 refactoring_strategy.md セクション 6・7 に準拠し、slogを用いたEntry/Exitの構造化デバッグログ（TraceID付き）がファイル出力されることを確認する
