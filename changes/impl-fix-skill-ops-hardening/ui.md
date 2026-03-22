# UI

## 方針

この change はアプリの画面 UI を直接変更しない。
変更対象は `.codex/skills/` 配下の運用契約と packet schema であり、観測可能な差分は Markdown artifact と skill 応答形式に限定する。

## Operator-visible Facts

- `impl` lane では `changes/<id>/tasks.md` が section 進捗の正本として更新される
- `impl` lane の reroute / resume 判断は、長い履歴ではなく progress snapshot と condensed brief を基準に行う
- `fix` lane では再現状態、観測ログ状態、fix scope、review 状態を短い state summary で引き継ぐ
- `fix` lane の一時ログは add / remove lifecycle が state summary 上で追跡される

## Non-goals

- Wails / React の表示追加
- skill 状態を表示する専用ダッシュボード追加
- 実装コードの自動修正ロジック追加
