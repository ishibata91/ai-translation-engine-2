# Scenarios

## Goal

`impl` lane と `fix` lane の orchestrator が、停止・再開・reroute・cleanup を短い状態要約で扱い、完了済み作業の再実行と不要な再読を減らせるようにする。

## Scenario 1: impl-direction が中断後に section 進捗を再開する

### Trigger

- 長時間 change の途中で orchestrator を再開する
- review reroute 後に dispatch 候補を決め直す

### Preconditions

- `changes/<id>/tasks.md` が存在する
- `impl-workplan` が progress snapshot を持つ tasks 形式を出力済みである
- 直近 section の結果が `completed` `blocked` `in_progress` のいずれかで記録されている

### Main Flow

1. `impl-direction` は `tasks.md` を読み、section status snapshot を取得する
2. `impl-direction` は直近 review / reroute 結果を progress summary へ圧縮する
3. `impl-direction` は status が `pending` または reroute 指定された section だけを dispatch 候補にする
4. `impl-direction` は dispatch 時に full history ではなく condensed section brief と current snapshot を渡す
5. section 完了または blocked 返却後、`impl-direction` は `tasks.md` の該当 section を更新する
6. `impl-direction` は完了済み subagent を close し、state summary の `next_dispatch` を更新する

### Alternate Flow

- `tasks.md` と review 結果が矛盾する場合、`impl-direction` は実装を始めず `impl-workplan` 再実行を要求する
- section が `blocked` で、原因が shared contract 未固定なら `impl-direction` は worker 再投入を行わず workplan 再実行へ戻す
- section が `blocked` で、原因が external validation noise だけなら status を `completed_with_noise` 相当の注記付きで保持し reroute しない

### Acceptance Criteria

- 再開時に完了済み section が再 dispatch されない
- dispatch packet が full history ではなく condensed brief だけで着手可能である
- `tasks.md` が orchestrator 記憶より優先される正本になる
- 完了済み agent が close され、agent 上限で詰まりにくい

## Scenario 2: impl worker が section 完了結果を標準形式で返す

### Trigger

- `impl-backend-work` または `impl-frontend-work` が section 実装を終える
- owned paths 外の変更が必要になり blocked になる

### Main Flow

1. worker は `completed` または `blocked` のどちらかで必ず停止する
2. worker は `completed_scope` `remaining_gap` `validation_result` `noise_classification` を結果に含める
3. `impl-direction` は結果を読み、`tasks.md` の `実装` `検証` を更新する
4. `impl-direction` は `noise_classification` が `external_validation_noise` または `known_pre_existing_issue` の場合、reroute 対象から除外する

### Acceptance Criteria

- blocked 返却だけではなく、section 内でどこまで終わったかが残る
- validation 失敗が section failure か外部ノイズかを判別できる
- reroute は必要 section に限定される

## Scenario 3: impl-review reroute が状態要約で再投入される

### Trigger

- `impl-review` が `score < 0.85` で `required_delta` を返す

### Main Flow

1. `impl-direction` は review 全文を保持し続けず、`required_delta` `affected_sections` `progress_snapshot` `carry_over_contracts` を reroute packet に圧縮する
2. `impl-direction` は affected section の元 work order に reroute packet を付与して再 dispatch する
3. worker は変更理由を理解できるが、過去の経緯全文は読まなくてよい

### Acceptance Criteria

- reroute packet が状態要約で閉じる
- 非 affected section は再投入されない
- orchestrator の context 膨張が抑えられる

## Scenario 4: fix-direction が調査状態を短い summary で引き継ぐ

### Trigger

- bugfix flow が distill から trace、analysis、fix-work、review へ進む
- 追加再現待ちをまたいで fix flow を再開する

### Preconditions

- bugfix packet が存在する
- `fix-direction` が state summary を持つ

### Main Flow

1. `fix-direction` は distill 結果から `reproduction_status` `known_facts` `unknowns` `active_logs` `current_scope` を state summary に落とす
2. `fix-trace` は full log ではなく bugfix packet と summary を基に仮説と観測計画を返す
3. `fix-logging` add 後、`fix-direction` は `active_logs` と `reproduce_steps` を summary に保持して再現待ちへ入る
4. 再現後、`fix-analysis` はログ全文ではなく重要イベントの圧縮結果を返す
5. `fix-direction` は summary を更新し、scope 確定後だけ `fix-work` を起動する
6. `fix-review` 後は `required_delta` と `docs_sync_needed` を summary に反映する
7. accept 到達時だけ `fix-logging remove` を起動し、cleanup 完了を summary に反映する

### Alternate Flow

- docs-only の仕様乖離に収束した場合、`fix-direction` は code fix を起動せず `plan-direction` へ handoff する
- 追加観測が不要な場合、`fix-analysis` は省略し summary を直接 `fix-work` へ渡す
- cleanup 未完了のまま完了しそうな場合、`fix-direction` は終了せず remove を必須で起動する

### Acceptance Criteria

- 再現待ちをまたいでも active log と現在の調査段階が失われない
- fix scope 確定前に `fix-work` が起動されない
- cleanup 漏れが完了条件違反として検出される

## Scenario 5: fix-work と fix-review が外部ノイズを区別する

### Trigger

- fix 実装後の品質ゲートで unrelated failure が混ざる
- review が未解消リスクと既知前提を区別する必要がある

### Main Flow

1. `fix-work` は検証結果を `scope_failure` `external_validation_noise` `known_pre_existing_issue` に分類して返す
2. `fix-review` は 7 field の既存 schema を維持したまま、`required_delta` と `recheck` の内容で未解消 scope と外部ノイズを区別できるように返す
3. `fix-direction` は review feedback と state summary を照合し、`score < 0.85` でも外部ノイズのみなら review loop に戻さず residual risk として扱う

### Acceptance Criteria

- bugfix 本体の失敗と外部ノイズが分離される
- review loop が不要な再作業を生みにくくなる
- 最終 handoff に未解消 scope と外部ノイズが別枠で残る
