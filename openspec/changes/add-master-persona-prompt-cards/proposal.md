## Why

MasterPersona 画面にはペルソナ生成時のプロンプト構成をユーザーが確認・調整する導線がなく、システム補完値とユーザー入力の責務も混在している。プロンプトの役割を UI と永続化仕様の両方で明確化し、再起動後も一貫した設定を再利用できる状態にする必要がある。

## What Changes

- `MasterPersona` 画面に、ユーザーが編集できる「ユーザープロンプト入力カード」を追加する。
- 既存ユーザープロンプト内のうち、システムが自動注入している固定値・補完値をシステムプロンプト側へ移し、責務を分離する。
- システムプロンプトは UI 上で読み取り専用カードとして表示し、ユーザーには参照のみ許可する。
- MasterPersona で使用するユーザープロンプト／システムプロンプトを `config` 仕様に基づいて永続化し、画面再表示時に復元する。

## Capabilities

### New Capabilities

- なし

### Modified Capabilities

- `config`: MasterPersona のユーザープロンプトと読み取り専用システムプロンプトを永続化・再読込できるよう requirement を拡張する
- `persona`: MasterPersona 画面でユーザープロンプトを編集し、システムプロンプトを読み取り専用表示する UI/プロンプト構成ルールへ requirement を拡張する

## Impact

- 影響コード: [MasterPersona.tsx](C:/Users/shiba/.codex/worktrees/186b/ai translation engine 2/frontend/src/pages/MasterPersona.tsx) を中心とした MasterPersona UI、関連する Wails 設定読み書き経路、ペルソナ用プロンプト組み立て処理
- 影響仕様: [spec.md](C:/Users/shiba/.codex/worktrees/186b/ai translation engine 2/openspec/specs/config/spec.md)、[spec.md](C:/Users/shiba/.codex/worktrees/186b/ai translation engine 2/openspec/specs/persona/spec.md)
- API/依存関係: 既存の `ConfigGetAll` / `ConfigSet` を継続利用し、新規ライブラリ追加は行わない
