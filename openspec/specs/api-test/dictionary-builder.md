# Dictionary Builder API Test Scope

`api-test` capability における Dictionary Builder 系 controller API テストの対象と観点を定義する。

## 対象 controller

- `pkg/controller/dictionary_controller.go`

## 対象公開メソッド

- `DictGetSources`
- `DictDeleteSource`
- `DictGetEntries`
- `DictGetEntriesPaginated`
- `DictSearchAllEntriesPaginated`
- `DictUpdateEntry`
- `DictDeleteEntry`
- `DictStartImport`

## テスト観点

### 正常系

- ソース一覧取得が service 戻り値をそのまま返せる
- エントリ一覧、ページング検索、横断検索が入力パラメータを保持して service へ渡る
- エントリ更新、削除、ソース削除、インポート開始が成功時に error なく完了する
- `DictStartImport` が生成された task ID を返せる

### 主要異常系

- 各 service 呼び出しエラーが controller からそのまま返る
- 検索条件、ページ番号、source ID などが境界値でも panic しない
- update / delete / import の失敗が error として上位へ返る

### 境界責務

- `SetContext` で注入した context が service 呼び出しへ伝播する
- trace ID 付き context が `telemetry.WithTraceID` を通じて維持される
- controller 側で不要な DTO 変換や workflow 判断を持ち込まない

## 優先ケース

| ケースID | 対象 | 目的 |
| :-- | :-- | :-- |
| DBAPI-01 | `DictGetSources` | 一覧取得の正常系を固定する |
| DBAPI-02 | `DictStartImport` | Dictionary Builder の開始導線を固定する |
| DBAPI-03 | `DictGetEntriesPaginated` | query / filters / page 引数伝播を確認する |
| DBAPI-04 | `DictSearchAllEntriesPaginated` | 横断検索の入力伝播を確認する |
| DBAPI-05 | `DictUpdateEntry` | 更新失敗が error で返ることを確認する |
| DBAPI-06 | `DictDeleteSource` | 削除失敗が error で返ることを確認する |
