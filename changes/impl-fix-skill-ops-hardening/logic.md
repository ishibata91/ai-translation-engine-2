# Logic

## Goal

`impl` / `fix` 系 skill の orchestrator と worker/fixer 間の受け渡しを、長い履歴依存から state summary と condensed packet 中心へ移し、停止・再開・reroute・cleanup の判断を安定化する。

## Inputs

- [2026-03-22-impl-skill-operation-report.md](/F:/ai translation engine 2/.codex/reports/2026-03-22-impl-skill-operation-report.md)
- 現行の `.codex/skills/impl-*` 契約
- 現行の `.codex/skills/fix-*` 契約

## Outputs

- `impl-direction` / `impl-workplan` / worker skill の新しい packet 契約
- `fix-direction` / `fix-*` skill の state summary 契約
- 実装時の編集対象と優先順位

## Primary Logic

### 1. 共通運用原則

#### 1.1 State Summary First

orchestrator は長い履歴全文ではなく、次の短い state summary を正本として持つ。

- current_stage
- completed_units
- blocked_units
- unresolved_contracts
- active_temp_assets
- latest_review_result
- next_action

`impl` lane の `completed_units` は section 単位、`fix` lane の `completed_units` は distill / trace / analysis / work / review の stage 単位で扱う。

#### 1.2 Condensed Packet First

下流 skill へ渡す packet は「追加探索を始めるための資料一覧」ではなく、「この時点で確定済みの判断材料を圧縮した本文」を主とする。

packet は以下を固定で持つ。

- scope_summary
- required_delta または goal
- fixed_contracts
- non_goals
- known_blockers
- validation_baseline
- must_read

`must_read` は補助導線であり、本文の代替ではない。

#### 1.3 Outcome Classification

worker / fixer の返却は最低でも次を分離する。

- result: `completed` | `blocked`
- scope_failure
- external_validation_noise
- known_pre_existing_issue
- completed_scope
- remaining_gap

これにより orchestrator は reroute と residual risk を分離できる。

#### 1.4 Lifecycle Cleanup

orchestrator は一時資産の生成と破棄を state summary 上で管理する。

- `impl-direction`: subagent lifecycle
- `fix-direction`: temporary log lifecycle

完了条件には cleanup 完了を含める。

### 2. impl lane

#### 2.1 `impl-workplan` の責務追加

`impl-workplan` は section plan と `tasks.md` に加えて、section ごとの condensed brief を生成する。

section brief は少なくとも次を含む。

- section goal summary
- why_now
- fixed shared contract
- non_goals
- known blockers
- validation baseline
- carry_over notes

また `tasks.md` には status snapshot を持たせる。

- `pending`
- `in_progress`
- `completed`
- `blocked`
- `completed_with_noise`

#### 2.2 `impl-direction` の責務変更

`impl-direction` は `tasks.md` を impl lane の進捗正本として扱う。

主な追加責務:

- 再開時 reconciliation
- section 結果に応じた `tasks.md` 自動更新
- reroute packet の状態要約化
- completed agent の close

`tasks.md` と review / reroute 結果が矛盾する場合は、worker 再投入ではなく `impl-workplan` 再実行へ戻す。

#### 2.3 `impl-backend-work` / `impl-frontend-work` の責務変更

worker は 1 section 完了または blocked で停止する原則を維持した上で、返却 schema を厳密化する。

必須返却項目:

- section_id
- result
- completed_scope
- remaining_gap
- changed_paths
- validation_result
- noise_classification
- reroute_hint

`owned_paths` 外が必要になった場合は、`reroute_hint` に「次 section で解消可能か」「workplan 再実行が必要か」を含める。

### 3. fix lane

#### 3.1 `fix-direction` の責務追加

`fix-direction` は bugfix flow 全体の state summary を持つ。

必須 summary 項目:

- reproduction_status
- known_facts
- unknowns
- active_logs
- current_hypothesis
- current_scope
- latest_review_result
- next_action

これにより、再現待ちや review loop をまたいでも full history を持ち続ける必要を減らす。

#### 3.2 `fix-distill` / `fix-trace` / `fix-analysis` の packet 見直し

- `fix-distill` は関連資料の列挙だけでなく、再現条件と観測ギャップの condensed brief を返す
- `fix-trace` は原因仮説、観測ポイント、ログ必要性、scope 未確定点を summary 更新可能な形で返す
- `fix-analysis` はログ全文ではなく、時系列と重要イベントの圧縮結果を返す

この 3 skill は原因推定・恒久修正・次工程起動の責務を増やさず、summary を更新しやすい packet へ寄せる。

#### 3.3 `fix-work` / `fix-review` の outcome 体系化

`fix-work` は `fix scope` 内でどこまで完了したかと、検証ノイズ分類を返す。
`fix-review` は既存の 7 field schema を維持しつつ、`required_delta` と `recheck` の内容で次を区別できるようにする。

- unresolved_fix_scope
- external_validation_noise
- residual_risks
- docs_sync_needed

`fix-direction` は review feedback と state summary を突き合わせ、`score < 0.85` の場合でも未解消が scope failure ではなく外部ノイズだけなら、loop 継続ではなく residual risk 管理へ落とせる。

#### 3.4 `fix-logging` lifecycle 固定

`fix-direction` は `fix-logging add` で受け取った `log_additions` を state summary に保持し、`score >= 0.85` 後の `remove` に必ず引き渡す。

cleanup 完了前に flow を閉じてはならない。

## Responsibilities

### Orchestrator

- stage 進行判断
- summary 更新
- cleanup 管理
- reroute / handoff 判定

### Distill / Trace / Analysis

- 事実圧縮
- 観測ギャップ整理
- summary 更新用 packet 返却

### Work / Review

- 単位作業の完了判定
- scope failure と外部ノイズの切り分け
- reroute / residual risk の材料返却

## Persistence Boundary

- `impl` lane の進捗正本は `changes/<id>/tasks.md`
- `fix` lane の進行正本は direction が保持する state summary packet
- review 結果は全文保持ではなく、summary に必要な要素だけを昇格させる

## Risks / Open Questions

- `fix` lane の state summary をどこに保存するかは未確定であり、まずは skill 応答 packet 内の明示契約として導入する
- `completed_with_noise` のような status 拡張を `tasks.md` format 正本へどう反映するかは `impl-workplan` テンプレート更新時に要確認
- `fix-review` は 7 field 固定契約を維持する前提のため、未解消 scope と外部ノイズの詳細は `required_delta` / `recheck` の構造化記述か summary 側の補助 field へ寄せる必要がある

## Docs Promotion Candidates

- `impl-direction` / `impl-workplan` / worker skill の packet 契約見直し
- `fix-direction` / `fix-*` の state summary 契約と cleanup 完了条件
- blocked / validation noise の標準分類
