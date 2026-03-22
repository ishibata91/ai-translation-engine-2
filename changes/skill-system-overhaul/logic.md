# Logic Design

## Scenario
`.codex/skills` 群の user-facing 入口、下流 workflow、template schema、agent 契約を再編し、`impl-workplan` と専用 `workplan_builder` agent を含む監査可能で壊れにくい skill system に整理する。

## Goal
skill 定義の正本を AGENTS.md と整合させつつ、direction / distill / workplan / work / review / sync の責務境界を明確にし、`tasks.md` を impl lane の正本 artifact として運用できるようにすること。

## Responsibility Split
- Controller:
  - AGENTS.md が user-facing 入口、使用可能 MCP、優先ルールを定義する。
  - skill 単体では global policy を上書きせず、必要なら handoff だけを返す。
- Workflow:
  - direction skill が lane 判定、chain 順序、review loop、docs handoff の管理を担う。
  - `plan-direction` `impl-direction` `fix-direction` `investigation-direction` だけを user-facing 入口として扱う。
  - UI polish は `impl-direction` 配下の specialized branch として扱い、独立入口とは見なさない。
- Slice:
  - distill / workplan / review / sync skill は packet schema に従った read-only もしくは限定 write を担う。
  - `impl-workplan` は `impl-distill` の implementation packet を読み、モジュール/契約単位の section plan と `tasks.md` を確定する。
  - `impl-frontend-work` `impl-backend-work` は section work order を受け取る正本 execution skill とする。
- Artifact:
  - `SKILL.md` は lane ルール、入力契約、出力契約、禁止事項だけを書く。
  - `references/templates.md` は packet schema の唯一の正本とし、direction と下流 skill の期待値を一致させる。
  - `review.md` は今回見つかった矛盾の記録と優先度を保持する。
  - `tasks.md` は `changes/<id>/tasks.md` に置く impl lane 専用の実装計画正本であり、section 分割、依存順、owner、契約、検証条件を保持する。
- Runtime:
  - `.codex/agents/*.toml` は実在する role と実行制約を定義する。
  - `workplan_builder` agent を新設し、`impl-workplan` 専用に section 分割と `tasks.md` 生成を担わせる。
  - `implementer` agent は `impl-frontend-work` または `impl-backend-work` を実行する container role として残す。
- Gateway:
  - server-filesystem, go-llm-lens, ts-lsp が skill の探索・参照・編集 I/O を提供する。
  - skill 文書は存在しない tool 名や旧名を使わない。

## Data Flow
- 入力
  - AGENTS.md の global policy
  - `.codex/skills/**/SKILL.md`
  - `.codex/skills/**/references/*.md`
  - `.codex/agents/*.toml`
- 中間成果物
  - 入口ルールの差分一覧
  - workflow 断絶一覧
  - schema mismatch 一覧
  - 改善 proposal 一覧
- 出力
  - overhaul 方針
  - 実施タスク順
  - impl lane へ渡せる修正 backlog

## Main Path
1. 監査で AGENTS.md を基準に user-facing 入口 skill の正当集合を確定する。
2. 各 direction skill が参照する下流 skill、agent、template を実在確認し、死んだ参照を抽出する。
3. packet schema を比較し、direction 側の判定条件と review / distill template の不一致を洗い出す。
4. impl lane では `impl-direction -> impl-distill -> impl-workplan -> impl-frontend-work / impl-backend-work -> impl-review` を正本 chain とし、`tasks.md` 生成主体を `impl-workplan` に固定する。
5. `impl-workplan` が section ごとの `owner / depends_on / shared_contract / owned_paths / forbidden_paths / validation / acceptance` を確定し、`workplan_builder` agent が `changes/<id>/tasks.md` を書き出す前提へ揃える。
6. logging、docs handoff のような lifecycle を持つ workflow について、追加・更新・削除の各責務を 1 skill 1 役割で割り当て直す。
7. 正本 workflow を `入口 -> distill -> workplan -> work/review -> sync/handoff` に単純化し、旧 workflow 名は削除または alias 廃止対象にする。
8. tool 利用規約を server-filesystem / go-llm-lens / ts-lsp に寄せ、旧 tool 記述を置換する。
9. 監査ルールを script または checklist として追加し、将来の skill 追加時に自動検証できるようにする。

## Key Branches
- 入口 skill の衝突:
  - `aite2-ui-polish-orchestrate` は user-facing 入口として書かれているが、AGENTS.md と `plan-direction` の routing 例に反する。
  - 解決策は `impl-direction` 配下の branch へ統合するか、internal helper に格下げするかの 2 択に限定する。
- impl workflow の断絶:
  - `impl-direction` は `implementer` 直実装を正本としている一方、`impl-distill` と worker skill 群は `impl-workplan` を前提にしている。
  - 解決策は `impl-workplan` と専用 `workplan_builder` agent を新設し、task 分割と `tasks.md` 生成を impl lane に統合することに固定する。
  - `impl-review` は section ごとではなく統合差分に 1 回だけ適用し、差し戻し時のみ affected section を再投入する。
- fix logging lifecycle の断絶:
  - `fix-direction` は最終 cleanup を要求するが、`fix-logging` は追加専用で削除契約を持たない。
  - `fix-logging` を add/remove 両対応に拡張するか、`fix-log-cleanup` を分離する必要がある。
- review schema の断絶:
  - `fix-direction` は `score` 判定を行うが、`fix-review` template には `score` がない。
  - review skill と template の schema を先に揃えないと loop 制御が不安定になる。
- tool 記述の断絶:
  - investigation skill 群には `find_by_name` `view_file` `grep` など現行ルールにない記述が残っている。
  - すべて server-filesystem / go-llm-lens / ts-lsp の表現へ置換する。

## Persistence Boundary
- AGENTS.md:
  - global policy の正本
- `.codex/skills/*/SKILL.md`:
  - lane / contract / procedure の正本
- `.codex/skills/*/references/*.md`:
  - packet schema と補助 checklist の正本
- `.codex/agents/*.toml`:
  - 実行 role と権限の正本
- `changes/skill-system-overhaul/*.md`:
  - 今回の差分設計と改善 backlog の正本
- `changes/<id>/tasks.md`:
  - `impl-workplan` が生成する impl lane 専用の実装計画正本

## Side Effects
- skill 定義の routing を削除または統合する
- `impl-workplan` と `workplan_builder` agent を追加する
- packet template を更新する
- agent と skill の責務説明を同期する
- 将来は整合性検査 script を追加する

## Risks
- 正本 workflow を決めないまま文言だけ更新すると、旧 workflow 参照が残って再度壊れる
- user-facing 入口を整理せずに specialized skill を増やすと、routing が AGENTS.md から逸脱する
- template だけ修正して SKILL.md を直さないと、再び schema mismatch が発生する
- `impl-workplan` が section owner と shared contract を曖昧にすると、worker が設計判断を持ち始めて impl lane が破綻する
- logging cleanup を曖昧にしたまま fix flow を回すと、一時ログが恒久残留する
- `ui.md` 不在時の impl 判定を単純化しすぎると、frontend 影響がある change を backend-only と誤分類する

## Improvement Proposals
- Proposal 1: 入口を 4 direction skill に固定し、specialized skill はすべて internal helper へ再分類する。
- Proposal 2: impl lane は `impl-workplan` と `workplan_builder` agent を新設し、`tasks.md` の生成責務を plan lane から impl lane へ完全移管する。
- Proposal 3: fix lane は `trace -> log-add -> analysis -> work -> review -> log-cleanup` の lifecycle を明文化し、prefix を `[fix-trace]` に統一する。
- Proposal 4: 全 skill に共通の section 順を導入する。`起動確認` `適用可否` `入力契約` `出力契約` `手順` `禁止事項` `参照` を必須にする。
- Proposal 5: template lint を追加し、存在しない skill 名、agent 名、tool 名、欠落 field を CI または手動チェックで検出する。
- Proposal 6: `impl-direction` の artifact readiness を `ui.md 有無` ではなく `変更対象が frontend を含むか` で判定する matrix に改める。
- Proposal 7: `implementer` は execution container role に縮退し、正本 execution skill は `impl-frontend-work` / `impl-backend-work` に戻す。

## Context Board Entry
```md
### Logic Design Handoff
- 確定した責務境界: AGENTS.md は global policy、direction skill は routing と loop 管理、`impl-workplan` は section planning と tasks 生成、work skill は section 実装、templates は schema 正本、agents は runtime role を担当する
- docs 昇格候補: なし。今回の change は `.codex/skills` 運用設計の差分として保持する
- review で見たい論点: UI polish skill の最終位置、`impl-workplan` の section schema 固定、fix logging cleanup の分離要否
- 未確定事項: specialized internal helper をどこまで残すか
```
