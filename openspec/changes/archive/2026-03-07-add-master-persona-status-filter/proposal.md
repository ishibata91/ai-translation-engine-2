## Why

MasterPersona のペルソナ一覧は現在すべての行を完了扱いで表示しており、下書き保存だけされた未生成レコードと、実際に生成済みのレコードを区別できない。ペルソナの再開運用と一覧確認の精度を上げるため、保存状態をDBで明示し、一覧側で状態別にフィルタできるようにする必要がある。

## What Changes

- `npc_personas` に英語値で管理するステータスを追加し、下書き保存時と生成完了時で状態を切り替えられるようにする。
- MasterPersona 一覧で `下書き` と `生成済み` を表示し、ステータスで絞り込みできる UI を追加する。
- ペルソナ一覧取得DTOとサービス応答にステータスを含め、フロントエンドの固定 `完了` 表示を廃止する。
- `specs/database_erd.md` の persona コンテキストを更新し、`npc_personas.status` の追加を反映する。

## Capabilities

### New Capabilities
- なし

### Modified Capabilities
- `persona`: MasterPersona 一覧で下書きと生成済みを区別できる状態管理、および状態フィルタ要件を追加する

## Impact

- バックエンド: `pkg/persona` の SQLite スキーマ、一覧取得DTO、保存フェーズの状態更新処理
- フロントエンド: [frontend/src/pages/MasterPersona.tsx](C:/Users/shiba/.codex/worktrees/75c7/ai%20translation%20engine%202/frontend/src/pages/MasterPersona.tsx) と [frontend/src/types/npc.ts](C:/Users/shiba/.codex/worktrees/75c7/ai%20translation%20engine%202/frontend/src/types/npc.ts) の一覧表示・フィルタUI
- 仕様書: [openspec/specs/persona/spec.md](C:/Users/shiba/.codex/worktrees/75c7/ai%20translation%20engine%202/openspec/specs/persona/spec.md) と [openspec/specs/database_erd.md](C:/Users/shiba/.codex/worktrees/75c7/ai%20translation%20engine%202/openspec/specs/database_erd.md)
- 依存追加は想定しない。既存の React / Wails / SQLite 構成の範囲で対応する
