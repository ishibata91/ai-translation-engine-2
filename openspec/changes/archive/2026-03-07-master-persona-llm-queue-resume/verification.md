# 検証ログ (2026-03-07)

## 自動テスト
- 実行: `go test ./pkg/...`
- 結果: PASS

## フロントエンドビルド
- 実行: `npm run build` (`frontend/`)
- 結果: FAIL
- 失敗理由: `frontend/src/wailsjs/**` が未生成のため TypeScript 解決不可
  - 例: `Cannot find module '../wailsjs/go/task/Bridge'`
- 備考: `wails dev` または `wails generate module` 相当の生成が必要

## 手動確認 (5.4 対応観点)
- 目的: 途中失敗を挟んだ再開時に重複保存がないこと
- 本セッション結果:
  - コード観点:
    - `saved_request_ids` により保存済み request を再保存しない制御を確認
    - `npc_personas` は `speaker_id` PK の UPSERT を確認
  - 実機手動試験:
    - 未実施（LM Studio 実行環境と UI 操作が必要）

## 今回の追加確認観点
- `persona.db` が `db/persona.db` に生成されること
- 保存完了後に `MasterPersona` テーブルへレスポンス由来行が反映されること
- completed タスクが Dashboard 一覧から除外されること
- `MasterPersona` 画面が completed 後に `Running` 表示のまま残らないこと