# Tasks: Rename Slices and Infrastructure

## 1. Specification Artifact Renaming

- [x] 1.1 `specs/export-slice` を `specs/export` に名称変更
- [x] 1.2 `specs/llm-client` を `specs/llm` に名称変更
- [x] 1.3 `specs` 内の各ドキュメントにおける相対リンクおよび名称の更新

## 2. Infrastructure Specification Completion

- [x] 2.1 `specs/datastore/spec.md` の作成
- [x] 2.2 `specs/queue/spec.md` の作成
- [x] 2.3 `specs/telemetry/spec.md` の作成
- [x] 2.4 `specs/progress/spec.md` の作成

## 3. Package and Directory Renaming (Implementation Phase)

- [x] 3.1 `pkg/export_slice` を `pkg/export` に名称変更
- [x] 3.2 `pkg/infrastructure/database` を `pkg/infrastructure/datastore` に名称変更
- [x] 3.3 `pkg/infrastructure/job_queue` を `pkg/infrastructure/queue` に名称変更
- [x] 3.4 `pkg/infrastructure/logger` を `pkg/infrastructure/telemetry` に名称変更
- [x] 3.5 `pkg/infrastructure/llm_client` と `llm_manager` の統合および `pkg/infrastructure/llm` への名称変更

## 4. Code Reference Updates

- [x] 4.1 全 Go ファイルの import パス更新
- [x] 4.2 `package` 宣言の更新
- [x] 4.3 `wire.go` およびその他の DI 構成の更新
- [x] 4.4 テストコード内のパスおよび名称の更新

## 5. Verification

- [x] 5.1 プロジェクト全体のビルド確認 (`go build ./...`)
- [x] 5.2 全ユニットテストの実行 (`go test ./...`)
