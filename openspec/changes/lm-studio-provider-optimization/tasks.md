## 1. 契約と設定の更新

- [ ] 1.1 `pkg/llm` の共通契約に `ListModels(ctx)` と `GenerateStructured(...)` の必須契約（未実装は `ErrStructuredOutputNotSupported`）を追加する
- [ ] 1.2 モデル未指定時はエラーを返す契約へ変更し、自動選択ロジックを削除する
- [ ] 1.3 呼び出し側が `pkg/config` 経由で指定モデルIDを保持・参照するための設定キーと読み出し処理を更新する
- [ ] 1.4 既存 `local` / `local-llm` 設定を `lmstudio` へマップする互換レイヤーを実装する

## 2. LM Studio API連携の実装

- [ ] 2.1 `GET /api/v1/models` のクライアントを実装し、`models[]` を `ModelInfo` に正規化する
- [ ] 2.2 Queueジョブ開始時に `POST /api/v1/models/load` を呼び、`instance_id` をジョブ実行コンテキストに保持する
- [ ] 2.3 Queueジョブ完了時と中断時に `POST /api/v1/models/unload` を1回だけ呼ぶ処理を実装する（レスポンス単位で呼ばない）
- [ ] 2.4 `POST /v1/chat/completions` で `response_format.type=json_schema` / `json_schema.strict=true` を付与する Structured Output 呼び出しを実装する
- [ ] 2.5 モデルロード失敗時は再試行せず即時失敗にするエラーハンドリングを実装する

## 3. queue再開性との統合

- [ ] 3.1 ジョブメタデータに `provider`、`model`、`request_fingerprint`、`structured_output_schema_version` を追加する
- [ ] 3.2 Resume時に保存済み `provider/model` を復元してロードし、同一条件で再実行するフローを実装する
- [ ] 3.3 メタデータ欠損時は再開不可として失敗終了するガードを追加する

## 4. テストと検証

- [ ] 4.1 `GET /api/v1/models` 正規化、`/api/v1/models/load`、`/api/v1/models/unload` のHTTPモックを使ったTable-Drivenテストを追加する
- [ ] 4.2 ジョブ単位load/unload（完了時・キャンセル時）と「load再試行なし」を検証するテストを追加する
- [ ] 4.3 Structured Output の `json_schema` リクエスト整形と未実装プロバイダ時エラー契約を検証するテストを追加する
- [ ] 4.4 queue resume 復元（同一 provider/model）とメタデータ欠損時失敗を検証するテストを追加する

## 5. 仕様・ドキュメント整合

- [ ] 5.1 `specs/llm` に models/load/unload/chat completions のAPI仕様が実装と一致していることを確認・更新する
- [ ] 5.2 `specs/queue/spec.md` の再開要件（ジョブ単位ライフサイクル、load再試行なし）と実装の差分を解消する
- [ ] 5.3 必要に応じて `specs/requirements.md` のLLM/queue関連記述を更新し、契約変更（モデル未指定エラー）を反映する
