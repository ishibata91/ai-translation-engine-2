# Spec 反映計画メモ (2026-03-07)

## 対象 change
- `master-persona-llm-queue-resume`

## 6.2 本体 spec 反映方針
- `openspec/specs/task/spec.md`
  - 既存 requirement (`task:updated` で状態通知、resume_cursor保持) で今回変更をカバー。
  - 追加更新は不要。運用上は completed を UI 側でクリーンアップ表示する。
- `openspec/specs/queue/spec.md`
  - 既存 requirement (`pending/running/completed/failed/canceled` と resume) で今回変更をカバー。
  - 追加更新は不要。
- `openspec/specs/llm/spec.md`
  - 既存 requirement (`provider=lmstudio` 制約、config 再読込) で今回変更をカバー。
  - 追加更新は不要。
- `openspec/specs/progress/spec.md`
  - 既存 requirement (`task` が phase/current/total 決定、progress は中継のみ) で今回変更をカバー。
  - 追加更新は不要。
- `openspec/specs/persona/spec.md`
  - 今回実装に合わせて DB 配置を `db/persona.db` に更新。
  - `npc_personas` スキーマを実装値 (speaker_id PK, updated_at) に合わせて調整。
- `openspec/specs/config/spec.md`
  - 既存 requirement (`master_persona.llm` namespace) で今回変更をカバー。
  - 追加更新は不要。

## 6.1 ERD 影響
- `openspec/specs/database_erd.md` の persona セクションを更新。
  - DB 名: `{PluginName}_persona.db` -> `persona.db`
  - 主キー: `id` -> `speaker_id`
  - カラム整合: `created_at` を削除し実装準拠へ統一。