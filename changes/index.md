# Changes

このディレクトリは進行中の仕様差分をブラウズするための入口です。`changes/<id>/` 配下に Markdown が追加されると、`docs-site` の Starlight ビルドで自動的にページ化され、`changes/<id>/index.md` も自動生成できます。

## ルール

- 仕様差分は `changes/<id>/` に置く
- 少なくとも `ui.md` `scenarios.md` `logic.md` `tasks.md` を候補として扱う
- `changes/<id>/index.md` は入口ページとして自動生成する
- 正本は `docs/`、`changes/` は差分と設計途中の置き場

## 読み方

- 進行中の change は sidebar の `Changes` グループから辿る
- `changes/<id>/tasks.md` は実装進行の入口として扱う
