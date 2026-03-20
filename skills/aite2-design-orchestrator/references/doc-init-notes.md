# Design Doc Init Notes

## `init-change-design-docs.ps1`
`scripts/init-change-design-docs.ps1` は `changes/<id>/` に `ui.md` `scenarios.md` `logic.md` の雛形を配置し、`context_board/` を初期化し、`index.md` を自動生成するための補助スクリプト。

## 使うタイミング
- change 文書がまだ無い
- 設計を `changes/<id>/` で順に進めたい
- `ui.md` `scenarios.md` `logic.md` の土台を先に用意したい
- context board を最初から作って design handoff を回したい
- `changes/<id>/` を Starlight の入口として見せたい

## 使わなくてよいケース
- 既存文書の軽微修正だけで済む
- docs 同期だけが目的
