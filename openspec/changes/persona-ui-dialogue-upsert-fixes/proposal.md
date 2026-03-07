## Why

MasterPersona 周辺で、詳細画面のセリフ表示欠落・同一 `speaker_id` の衝突・モックUIの件数不整合が同時に発生し、検証と運用の信頼性を下げている。`source_plugin` を考慮した識別と上書き制御を明示し、UI表示と保存挙動を一致させる必要がある。

## What Changes

- `PersonaDetail` の「セリフ一覧」を `persona.npc_dialogues` 由来データで確実に表示し、表示項目は原文セリフのみとする（訳文は表示・保存対象にしない）。
- `npc_personas` の識別キーを `speaker_id` 単独から `source_plugin + speaker_id` の複合一意制約へ拡張し、同一 `speaker_id` が複数プラグインに存在しても衝突しないようにする。
- JSONインポート開始時（MasterPersona）と生成カードUIに「重複時に上書き」チェックボックスを追加し、ON時は既存行を更新、OFF時は安全側（既存保持）で処理する。
- リクエスト生成完了（`REQUEST_GENERATED`）後は、ユーザーの手動操作を待たずに同一タスクIDでキュー実行（`ResumeTask`）へ自動遷移する。
- 翻訳処理のたびに `npc_dialogues` の訳文を更新しない運用に合わせ、ペルソナ用途の会話保存は原文中心の最小項目に限定する。
- DBスキーマ変更に合わせて `openspec/specs/database_erd.md` の persona セクションを更新する。

## Capabilities

### New Capabilities

- なし

### Modified Capabilities

- `persona`: `npc_personas` の一意性要件を `source_plugin + speaker_id` に変更し、重複時上書き可否を入力オプションとして扱う要件を追加する。
- `persona-request-preview`: MasterPersona 開始UIに「重複時上書き」指定を追加し、タスク開始パラメータへ反映する要件を追加する。
- `task`: `StartMasterPersonTask` の入力メタデータに上書き指定を保持し、再開時も同じ上書き方針を維持する要件を追加する。

## Impact

- 影響コード:
  - `frontend/src/pages/MasterPersona.tsx`
  - `frontend/src/components/PersonaDetail.tsx`
  - `pkg/persona/*`（スキーマ・一覧取得・保存ロジック）
  - `pkg/task/*`（開始入力・メタデータ）
- 影響データ:
  - `db/persona.db` の `npc_personas` 制約変更（必要に応じた移行処理）
  - `openspec/specs/database_erd.md` のER定義更新
- 依存関係:
  - 新規ライブラリ追加は不要（既存 React / Wails / SQLite 実装で対応）
