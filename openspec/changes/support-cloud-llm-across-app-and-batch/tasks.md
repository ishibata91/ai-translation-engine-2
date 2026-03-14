## 1. Capability Boundary Cleanup

- [x] 1.1 legacy な `master-persona-lmstudio-resume-flow` を `master-persona-execution-flow` へ寄せる方針を確定する
- [x] 1.2 spec 名変更で影響を受ける main specs / change specs / review 観点を洗い出す
- [x] 1.3 `master-persona-execution-flow` として provider 非依存の spec 名へ main specs を移行する
- [x] 1.4 change 配下の `master-persona-execution-flow` spec 名と tasks / design / proposal 内の参照を新しい spec 名へ揃える

## 2. Gemini Batch Gateway

- [x] 2.1 `docs/gemini batch ref.md` から submit / status / results の必要フィールドを実装メモへ落とす
- [x] 2.2 `pkg/gateway/llm` に Gemini BatchClient の submit 処理を追加する
- [x] 2.3 Gemini の status 取得処理を追加し、`BatchState` と `batchStats` を読めるようにする
- [x] 2.4 Gemini の結果取得処理を追加し、inlined response を共通 response へ変換する
- [x] 2.5 Gemini の生状態を `queued / running / completed / partial_failed / failed / cancelled` へ正規化する

## 3. xAI Batch Gateway Alignment

- [x] 3.1 `docs/xAI batch reference.md` と既存 xAI 実装の差分を洗い出す
- [x] 3.2 xAI status 取得を `num_pending / num_success / num_error / num_cancelled` 前提で見直す
- [x] 3.3 xAI results pagination を吸収して全件取得できるようにする
- [x] 3.4 xAI の生状態を共通 `BatchStatus.State` へ正規化する

## 4. Shared LLM Capability Model

- [x] 4.1 `BatchStatus.State` の共通 enum 相当を gateway 層で確定する
- [x] 4.2 既存 sync プロバイダ実装を capability DTO 前提へ寄せる
- [x] 4.3 モデル単位の `supports_batch` 相当 capability を `ModelInfo` / modelcatalog DTO に追加する
- [x] 4.4 provider ごとの batch 可否判定が UI や workflow に漏れていないことを確認する

## 5. Config Persistence

- [x] 5.1 `master_persona.llm.<provider>` 配下に `bulk_strategy` を保存する key 設計を確定する
- [x] 5.2 config 読み取り処理を拡張して `bulk_strategy` を復元できるようにする
- [x] 5.3 `bulk_strategy` 未保存時の既定値を `sync` にする後方互換を入れる
- [x] 5.4 provider 切替時に `bulk_strategy` を独立保持できることを確認する

## 6. Runtime Queue Resume

- [x] 6.1 batch 実行中 job の `batch_job_id` 保持経路を確認し、再接続条件を明確化する
- [x] 6.2 `pkg/runtime/queue` で `batch_job_id` がある job を再 submit しない resume 処理へ変更する
- [x] 6.3 batch job への再接続時に polling から再開できるようにする
- [x] 6.4 `partial_failed` でも成功分結果を後段へ渡せるように queue 更新を調整する

## 7. Workflow Integration

- [x] 7.1 `pkg/workflow/master_persona_service.go` で execution profile を扱う入力 / metadata を整理する
- [x] 7.2 provider 固有文言を廃し、共通 phase 名で進捗通知するように変更する
- [x] 7.3 batch submit 後の再開を「再投入」ではなく「既存 job 再接続」として扱う
- [x] 7.4 workflow が capability 上未対応の execution profile を開始前に拒否するようにする

## 8. Model Catalog / Controller

- [x] 8.1 `modelcatalog` が provider 由来 capability を返す DTO を設計する
- [x] 8.2 modelcatalog 実装で `supports_batch` 相当を返せるようにする
- [x] 8.3 Wails controller 境界で新 DTO を frontend が読める形へ公開する

## 9. Frontend State / Hooks

- [x] 9.1 `frontend/src/types/masterPersona.ts` に execution profile / bulk strategy / model capability の型を追加する
- [x] 9.2 `useModelSettings` を capability DTO 消費型へ変更する
- [x] 9.3 `useMasterPersona` に provider 名直書きではなく execution profile ベースの状態を持たせる
- [x] 9.4 `openai` を Master Persona 画面の選択肢から消すロジックを feature hook 側へ閉じ込める
- [x] 9.5 model capability の参照を `capability.supports_batch` に統一し、互換目的のトップレベル `supports_batch` を DTO から削除する

## 10. Frontend Presentation

- [x] 10.1 `ModelSettings.tsx` に `AIプロバイダ / モデル / 実行方式` を主入力として表示する UI を追加する
- [x] 10.2 capability DTO に従って `同期並列数` などの詳細項目を条件表示する
- [x] 10.3 モデルごとの `Batch API 対応` / `同期実行のみ対応` 補助文を表示する
- [x] 10.4 `MasterPersona.tsx` の進捗表示を remote progress 優先の文言へ更新する
- [x] 10.5 provider progress が取れない場合の不定 progress 表示を追加する

## 11. Backend Verification

- [ ] 11.1 変更する Go ファイルごとに `npm run backend:lint:file -- <file...>` を実行する
- [ ] 11.2 lint 指摘を反映した後、対象ファイルで `npm run backend:lint:file -- <file...>` を再実行する
- [ ] 11.3 batch gateway / queue / workflow の必要な `go test ./pkg/...` を実行する
- [ ] 11.4 最終確認として `npm run lint:backend` を実行する

## 12. Frontend Verification

- [x] 12.1 変更するフロントファイルごとに `npm run lint:file -- <file...>` を実行する
- [x] 12.2 lint 指摘を反映した後、対象ファイルで `npm run lint:file -- <file...>` を再実行する
- [x] 12.3 最終確認として `npm run typecheck` を実行する
- [x] 12.4 最終確認として `npm run lint:frontend` を実行する
- [x] 12.5 最終確認として Playwright E2E を実行する
