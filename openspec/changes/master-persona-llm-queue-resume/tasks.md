## 1. フェーズ1: MasterPersona request のキュー保存

- [x] 1.1 `pkg/task/master_persona_task.go` で `PreparePrompts` 結果を Queue 登録 DTO に変換する
- [x] 1.2 Queue 永続化層へ MasterPersona request state（pending/running/completed/failed/canceled）を保存する
- [x] 1.3 `task` メタデータへ再開に必要なキー（phase/resume_cursor）を保存する
- [x] 1.4 キューマネージャーから task/request 状態を取得できる API を追加する
- [x] 1.5 ユーザー動作テスト: 開始後に再起動しても request が失われないことを確認する

## 2. フェーズ2: MasterPersona LLM config 永続化

- [x] 2.1 `config` に `master_persona.llm` namespace の保存/読込 API を追加する
- [x] 2.2 `frontend/src/pages/MasterPersona.tsx` で初期ロード時に保存設定を読み込む
- [x] 2.3 `ModelSettings` 変更値を開始時に `master_persona.llm` として保存する
- [x] 2.4 未保存時の既定値フォールバック（空/既定パラメータ）を実装する
- [x] 2.5 ユーザー動作テスト: 設定保存後にアプリ再起動して値が復元されることを確認する

## 3. フェーズ3: LM Studio 経由の request 実行

- [x] 3.1 Queue worker の MasterPersona 経路を `provider=lmstudio` 限定で実装する
- [x] 3.2 `pkg/llm` 呼び出し時に再開時点の `config` から `provider/model` を再読込して適用する
- [x] 3.3 `task` 側で phase イベント（REQUEST_ENQUEUED / REQUEST_DISPATCHING / REQUEST_SAVING / REQUEST_COMPLETED）を決定し `progress` へ報告する
- [x] 3.4 キャンセル時に request state を保存し、再開 API で未完了 request のみ再送する
- [x] 3.5 ユーザー動作テスト: LM Studio 実行中にキャンセル→再開して途中から進行することを確認する

## 4. インフラ層の非依存性担保

- [x] 4.1 `pkg/progress` から MasterPersona 固有語彙を排除し、`task_type/phase/current/total` の汎用契約だけにする
- [x] 4.2 `pkg/queue` からスライス固有分岐を排除し、task 種別は `task_type` で受け取る
- [x] 4.3 `pkg/llm` からスライス名依存を排除し、Queue worker 由来コンテキストのみで実行する
- [x] 4.4 ユーザー動作テスト: MasterPersona 実行で機能を満たしつつ、インフラ層に固有実装が入っていないことを確認する

## 5. フェーズ4: ペルソナ保存と冪等再開

- [ ] 5.1 LLM レスポンスから persona 保存入力を生成し、既存保存層へ upsert する
- [ ] 5.2 再開時は completed request を再保存しない制御を追加する
- [ ] 5.3 保存成功/失敗件数を task summary と phase_completed イベントへ反映する
- [ ] 5.4 ユーザー動作テスト: 途中失敗を挟んでも最終的に重複なく persona が保存されることを確認する

## 6. 仕様・ドキュメント整合

- [ ] 6.1 実装に合わせて `openspec/specs/database_erd.md` の影響有無を確認し、必要なら更新する
- [ ] 6.2 `openspec/specs/task/spec.md` `queue/spec.md` `llm/spec.md` `progress/spec.md` `persona/spec.md` `config/spec.md` の本体 spec 反映計画を記録する
- [ ] 6.3 フェーズ別の手動テスト結果を change 配下に記録し、アーカイブ前の検証証跡を残す

## 7. 実装中に確定した追加仕様

- [x] 7.1 `provider=lmstudio` のモデルロードで `context_length` を設定可能にする（UI保存・再開時再読込を含む）
- [x] 7.2 `frontend/src/components/ModelSettings.tsx` で並列実行数（sync concurrency）を設定可能にし、`master_persona.llm` 名前空間へ保存する
- [x] 7.3 一時停止後の再起動/ダッシュボード再開時に進捗（current/total と progress%）を復元し、0から再開始しないようにする
