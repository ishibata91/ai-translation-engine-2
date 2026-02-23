## 1. 仕様書 (Specs) のメンテナンス

- [x] 1.1 `specs/` フォルダのリネーム（`git mv`）
  - `config-store` -> `config`
  - `loader` -> `parser`
  - `context-engine` -> `lore`
  - `persona-gen` -> `persona`
  - `dictionary-builder` -> `dictionary`
  - `term-translator` -> `terminology`
  - `pass2-translator` -> `translator`
  - `summary-generator` -> `summary`
  - `process-manager` -> `pipeline`
- [x] 1.2 各 `spec.md` 内の古い名称を置換
- [x] 1.3 `specs/refactoring_strategy.md` を `specs/architecture.md` にリネーム（参照箇所も全て更新）
- [x] 1.4 `requirements.md`, `database_erd.md`, `config.yaml` 内の参照を更新

## 2. ソースコード (pkg) のリネーム

- [x] 2.1 `pkg/` フォルダのリネーム（`git mv`）
  - `pkg/config_store` -> `pkg/config`
  - `pkg/loader_slice` -> `pkg/parser`
  - `pkg/context_engine` -> `pkg/lore`
  - `pkg/persona_gen` -> `pkg/persona`
  - `pkg/dictionary_builder` -> `pkg/dictionary`
  - `pkg/term_translator` -> `pkg/terminology`
  - `pkg/pass2_translator` -> `pkg/translator`
  - `pkg/summary_generator` -> `pkg/summary`
  - `pkg/process_manager` -> `pkg/pipeline`
- [x] 2.2 各ファイルの `package` 宣言を新しい名前に更新
- [x] 2.3 `infrastructure` パッケージ内などのパッケージ参照を修正

## 3. 依存関係とDIの更新

- [x] 3.1 全ファイルの `import` パスを一括置換
- [x] 3.2 `cmd/` 下のメイン関数やプロバイダ登録を修正
- [x] 3.3 `wire.go` を更新し、`wire gen` を実行して `wire_gen.go` を再生成
- [x] 3.4 テストコード内のパッケージ名・ディレクトリ参照を修正

## 4. 検証

- [x] 4.1 全パッケージのビルド確認 (`go build ./...`)
- [x] 4.2 全テストの実行 (`go test ./...`)
- [x] 4.3 `openspec status` で正常にステータスが取れるか確認
