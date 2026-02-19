# Data Loader Tasks

## 1. Contracts & Models

- [x] 1.1 `pkg/domain/models/models.go`: ExtractedData および関連構造体の定義 <!-- id: 1 -->
- [ ] 1.2 `pkg/contract/loader.go`: Loader インターフェース定義 <!-- id: 2 -->

## 2. Infrastructure (Loader)

- [ ] 2.1 `pkg/infrastructure/loader/encoding.go`: 文字コード自動判別ロジック実装 <!-- id: 3 -->
- [ ] 2.2 `pkg/infrastructure/loader/decoder.go`: `map[string]json.RawMessage` への一次デコード実装 <!-- id: 4 -->
- [ ] 2.3 `pkg/infrastructure/loader/parallel.go`: RawMessageからの並列Unmarshal & 正規化実装 <!-- id: 5 -->
- [ ] 2.4 `pkg/infrastructure/loader/loader.go`: パブリックAPI `LoadExtractedJSON` 実装 <!-- id: 6 -->

## 3. Verification

- [ ] 3.1 `pkg/infrastructure/loader/loader_test.go`: ユニットテスト作成 (正常系・異常系) <!-- id: 7 -->
- [ ] 3.2 `cmd/loader/main.go`: 動作確認用CLIツール作成 <!-- id: 8 -->
