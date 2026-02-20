# 辞書DB作成 テスト設計

## 1. ユニットテスト (Unit Tests)

### 1.1 `DictionaryImporter` (XML Parser)
*   **対象**: `ImportXML` メソッド
*   **テストケース**:
    *   正常系: 有効なxtranslator形式のXMLを渡し、正しく `DictTerm` のスライスが生成されること。(`EDID`, `REC`, `Source`, `Dest`, `Addon` が全て正確に抽出されるか)。
    *   異常系: 不正なXMLフォーマットを渡した場合、パースエラーを返すこと。
    *   エッジケース: `<Params>` の `Addon` が空の場合の挙動確認。

### 1.2 `DictionaryStore` (SQLite Adapter)
*   **対象**: `SaveTerms` メソッド
*   **テスト環境**: In-Memory SQLite (`file::memory:?cache=shared`) などを利用し、ディスクI/Oなしで高速テスト実行。
*   **テストケース**:
    *   正常系: 複数の `DictTerm` を渡し、エラーなくDBへ保存されること。
    *   正常系 (重複): すでに存在する `EDID` と `REC` の組み合わせに対し、UPSERT (ON CONFLICT REPLACE/UPDATE) 処理が正しく動作すること。
    *   異常系: DBコネクション切断状態でのエラーハンドリング確認。

## 2. 統合テスト (Integration Tests)

### 2.1 HTTP ハンドラー (`DictionaryHandler`)
*   **対象**: `HandleImport` メソッド
*   **テスト環境**: `net/http/httptest` を利用し、サーバーを立ち上げずにHTTPリクエストをモック実行。
*   **テストケース**:
    *   正常系: `multipart/form-data` 形式でダミーXMLファイルをPOSTし、ステータス `200 OK` およびインポートされた件数が返却されること。
    *   異常系: 未サポートのファイル形式（例：CSVやTXT）をPOSTした場合、ステータス `400 Bad Request` を返すこと。
    *   異常系: リクエストボディが空（ファイルなし）の場合のエラーハンドリング確認。

## 3. UI動作テスト (Manual / E2E)

*   **フロントエンド**: React UI（構築予定）より、ローカルにある `Dawnguard_english_japanese.xml` などの実XMLファイルを複数指定。
*   **検証項目**:
    *   複数ファイルの一括送信が正しく行えるか。
    *   大量のデータ（例: Skyrim本体のXML、約25MB）でもサーバーがOOM(Out Of Memory)にならずに処理できるか、もしくはストリーミング処理・分割処理が適切に働いているか。
    *   UI上へのプログレスバー表示や結果のフィードバックが適切か。
