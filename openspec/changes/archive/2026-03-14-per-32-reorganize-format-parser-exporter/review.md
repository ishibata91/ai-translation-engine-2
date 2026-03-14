# Review: per-32-reorganize-format-parser-exporter

## 1. ユーザが出した完了条件

- 完了条件 1: parser 実装を `pkg/format/parser/skyrim` へ移設し、`Parser` 契約名を維持する
- 完了条件 2: exporter 実装を `pkg/format/exporter/xtranslator` へ移設し、`Exporter` 契約名を維持する
- 完了条件 3: 旧配置依存を解消し、品質ゲートと OpenSpec の spec 配置を更新する

## 2. 品質ゲート確認

### Backend

- [x] 変更中ファイルに対して `npm run backend:lint:file -- <file...>` を逐次実行した
- [x] `backend:lint:file -> 修正 -> 再実行 -> 最後に lint:backend` の順で進めた
- [x] 作業中または完了前に `npm run lint:backend` を実行した
- [ ] 必要に応じて `npm run backend:check` または `npm run backend:watch` で品質確認した

### Frontend

- [ ] 変更中ファイルに対して `npm run lint:file -- <file...>` を逐次実行した
- [ ] `lint:file -> 修正 -> 再実行 -> 最後に lint:frontend` の順で進めた
- [ ] 作業完了前に `npm run lint:frontend` を実行した

## 3. 実行メモ

- 実行したコマンド:
  - `go test ./pkg/...`
  - `go test ./tests/...`
  - `go test ./...`
  - `npm run backend:fmt`
  - `npm run backend:lint:file -- pkg/format/parser/skyrim/contract.go pkg/format/parser/skyrim/dto.go pkg/format/parser/skyrim/decoder.go pkg/format/parser/skyrim/encoding.go pkg/format/parser/skyrim/loader.go pkg/format/parser/skyrim/parallel.go pkg/format/parser/skyrim/provider.go pkg/format/parser/skyrim/test/loader_test.go pkg/workflow/master_persona_service.go pkg/workflow/pipeline/mapper.go`
  - `npm run backend:lint:file -- pkg/format/exporter/xtranslator/contract.go pkg/format/exporter/xtranslator/dto.go pkg/format/exporter/xtranslator/exporter.go pkg/format/exporter/xtranslator/provider.go pkg/format/exporter/xtranslator/exporter_test.go`
  - `npm run lint:backend`
- 未実行の品質ゲートと理由:
  - `backend:check` / `backend:watch`: 任意導線のため未実行
  - フロントエンド品質ゲート: フロントエンド変更なしのため未実行
- レビュー時の補足:
  - `pkg/slice/parser` と `pkg/gateway/export` の参照はコード上から除去済み
  - `cmd/parser` の Go ファイルは削除済み（ディレクトリのみ空で残る）
  - OpenSpec 本体の spec を `openspec/specs/format/parser` と `openspec/specs/format/export` へ移設済み
