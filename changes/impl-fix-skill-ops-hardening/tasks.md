# Tasks

## 1. impl lane の正本管理を固める

- [ ] 1.1 `impl-workplan` の `tasks.md` format に status snapshot と condensed brief 参照を追加する
- [ ] 1.2 `impl-direction` に resume reconciliation 手順と `tasks.md` 自動更新手順を追加する
- [ ] 1.3 completed subagent を close する lifecycle を `impl-direction` に追加する

## 2. impl worker の返却契約を標準化する

- [ ] 2.1 `impl-backend-work` の result / validation / noise classification schema を固定する
- [ ] 2.2 `impl-frontend-work` の result / validation / noise classification schema を固定する
- [ ] 2.3 reroute packet を履歴型から状態要約型へ変更する

## 3. fix lane の state summary を導入する

- [ ] 3.1 `fix-direction` に reproduction / hypothesis / active_logs / review を含む state summary 契約を追加する
- [ ] 3.2 `fix-distill` `fix-trace` `fix-analysis` を summary 更新しやすい condensed packet へ寄せる
- [ ] 3.3 `fix-logging` add / remove の lifecycle を summary 管理前提へ更新する

## 4. fix 実装と review のノイズ分類を固める

- [ ] 4.1 `fix-work` に completed scope と external noise 分類の返却を追加する
- [ ] 4.2 `fix-review` の 7 field schema を維持したまま、`required_delta` / `recheck` で unresolved scope と residual risk を分離する運用を追加する
- [ ] 4.3 docs-only 乖離と code-fix 対象の分岐条件を `fix-direction` で再確認する

## 5. 正本テンプレートを同期する

- [ ] 5.1 `impl` / `fix` 各 skill の `references/templates.md` を schema 正本として更新する
- [ ] 5.2 必要なら `.codex/reports/` の運用レポート雛形にも summary / noise classification 観点を追加する
- [ ] 5.3 実装後に skill chain を dry-run し、resume / reroute / cleanup の各観点を検証する
