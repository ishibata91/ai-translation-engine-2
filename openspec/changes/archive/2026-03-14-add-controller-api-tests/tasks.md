## 1. 共通テスト基盤の拡張

- [x] 1.1 `pkg/controller` の公開メソッド棚卸しを行い、controller ごとの API テスト対象と不足テストを実装対象へ反映する
- [x] 1.2 `pkg/tests/api_tests/testenv` の共通責務を維持しつつ、controller 別 builder の追加方針に合わせて必要最小限の共通基盤を整える
- [x] 1.3 `FileDialogController` を API テスト可能にする最小限の seam 方針を決め、必要な差し替えポイントを実装する

## 2. Dictionary Builder 系 API テストの追加

- [x] 2.1 `pkg/tests/api_tests/dictionarycontroller` 配下に builder / fake を追加し、`DictionaryController` の依存を組み立てられるようにする
- [x] 2.2 `pkg/controller/dictionary_controller_test.go` を追加し、`specs/api-test/dictionary-builder.md` に沿って取得系 API の table-driven test を実装する
- [x] 2.3 `pkg/controller/dictionary_controller_test.go` に更新系 API と `DictStartImport` の正常系・主要異常系・context 伝播確認を追加する

## 3. Master Persona 系 API テストの追加

- [x] 3.1 `pkg/tests/api_tests/personataskcontroller` 配下に builder / fake workflow / fake manager store を追加する
- [x] 3.2 `pkg/controller/persona_task_controller_test.go` を追加し、開始・再開 API の入力写像と主要異常系を table-driven test で実装する
- [x] 3.3 `pkg/controller/persona_task_controller_test.go` に状態取得・cancel 委譲・workflow 未設定時 guard の検証を追加する

## 4. 残り controller API テストの整備

- [x] 4.1 `ModelCatalogController`、`PersonaController`、`TaskController` 向け builder / fake を追加し、公開メソッドの正常系・主要異常系をテスト化する
- [x] 4.2 `FileDialogController` の API テストを追加し、フィルタ設定、戻り値、error wrap を seam 経由で検証する
- [x] 4.3 既存 `ConfigController` テストとの整合を見直し、controller API テスト群の命名・構成・table-driven 方針を揃える

## 5. 品質ゲートと最終確認

- [x] 5.1 変更中の Go ファイルに対して `npm run backend:lint:file -- <file...>` を逐次実行し、違反を解消する
- [x] 5.2 `npm run lint:backend` を実行し、controller API テスト追加後もバックエンド lint を通す
- [x] 5.3 `go test ./pkg/...` と必要な backend 確認導線を実行し、追加した API テストが既存品質導線に含まれることを確認する
