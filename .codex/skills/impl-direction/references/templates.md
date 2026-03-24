# Impl Direction Templates

## packet 正本

- `impl-distill` 正本: `changes/<id>/context_board/impl-distill.packet.json`
- `impl-distill` validation: `changes/<id>/context_board/impl-distill.packet.validation.json`
- `impl-workplan` 正本: `changes/<id>/context_board/impl-workplan.packet.json`
- `impl-workplan` validation: `changes/<id>/context_board/impl-workplan.packet.validation.json`
- `impl-review` 正本: `changes/<id>/context_board/impl-review.feedback.json`
- `impl-review` validation: `changes/<id>/context_board/impl-review.feedback.validation.json`

> downstream packet の正本は会話本文ではなく JSON artifact とし、validation artifact が `valid: true` でない packet は不採用にする。

## 下流スキル起動
```md
### Skill Invocation
- invoked_skill: <起動する下流スキル名 (例: impl-distill)>
- invoked_by: impl-direction
- agent: <起動するエージェント名 (例: ctx_loader)>
- 対象 change:
- task:
- 入力:
- focus:
```

> 下流スキルのサブエージェントを立ち上げるときは必ずこのテンプレートを使い、
> `invoked_skill` と `invoked_by` を明示すること。
> サブエージェントは自分がどのスキルで起動されたかをこのフィールドで確認する。

## artifact 充足確認
```md
### Artifact Readiness
- change:
- ui.md:
- scenarios.md:
- logic.md:
- tasks.md: optional before first workplan / source of truth after generated
- issues:
- next_action:
```

## impl-workplan 起動
```md
Use $impl-workplan at skills/impl-workplan with `workplan_builder`.

Inputs:
- invoked_skill: impl-workplan
- invoked_by: impl-direction
- change:
- task:
- implementation_packet:
- focus:
```

## section plan 受領
`impl-workplan` の packet 返却後、worker を起動する前に必ずこのテンプレートで記録する。

```md
### Workplan Summary
- change:
- tasks_path:
- progress_snapshot:
- unresolved:
- shared_contracts:
- dispatch_order:
- sections:
  - section_id:
    title:
    owner: frontend | backend
    status:
    goal:
    depends_on:
    shared_contract:
    condensed_brief:
      why_now:
      fixed_contracts:
      non_goals:
      known_blockers:
      validation_baseline:
      carry_over_notes:
    required_reading:
    owned_paths:
    forbidden_paths:
    validation_commands:
    acceptance:
```

> `Workplan Summary` の `sections` は `impl-workplan` の `Section Plan` schema をそのまま写す。
> `title` `shared_contract` `required_reading` `validation_commands` `acceptance` を省略しないこと。

## impl-distill 起動
```md
Use $impl-distill at skills/impl-distill with `ctx_loader`.

Inputs:
- invoked_skill: impl-distill
- invoked_by: impl-direction
- change:
- task:
- must_check_artifacts:
- focus:
```

## section dispatch
```md
### Section Dispatch
- change:
- section_id:
- title:
- owner: frontend | backend
- goal:
- depends_on:
- progress_snapshot:
- invoked_skill:
- agent: implementer
- shared_contract:
- condensed_brief:
  why_now:
  fixed_contracts:
  non_goals:
  known_blockers:
  validation_baseline:
  carry_over_notes:
- required_reading:
- owned_paths:
- forbidden_paths:
- validation_commands:
- acceptance:
```

> `Section Dispatch` は `impl-workplan` の `Work Order` schema を包む dispatch envelope。
> `title` `goal` `depends_on` `shared_contract` `condensed_brief` `owned_paths` `forbidden_paths` `required_reading`
> `validation_commands` `acceptance` `progress_snapshot` を workplan 返却から書き換えずに引き継ぐこと。

## section 結果
```md
### Section Result
- change:
- section_id:
- result: completed | blocked
- completed_scope:
- remaining_gap:
- changed_paths:
- validation_result:
- noise_classification: none | scope_failure | external_validation_noise | known_pre_existing_issue
- reroute_hint:
- unverified:
```

> `Section Result` は worker が 1 section 完了または blocked で停止した時点の返却正本。
> `completed_scope` と `noise_classification` を省略しないこと。

## review 差し戻し
```md
### Review Reroute
- change:
- review_score:
- progress_snapshot:
- affected_sections:
  - section_id:
    title:
    owner: frontend | backend
    status:
    goal:
    depends_on:
    shared_contract:
    condensed_brief:
    required_reading:
    owned_paths:
    forbidden_paths:
    validation_commands:
    acceptance:
- required_delta:
- carry_over_contracts:
- recheck:
```

> `affected_sections` に含まれる section だけを再 dispatch し、元の full work order を崩さない。
> `carry_over_contracts` には `title` `goal` `depends_on` `shared_contract` `required_reading` `owned_paths` `forbidden_paths` `validation_commands` `acceptance` `condensed_brief` の再利用方針を書く。

## plan へ差し戻し
```md
### Plan Handoff
- reason:
- missing_or_stale_artifacts:
- required_plan_skill:
- must_read:
```

## plan-sync handoff
```md
### Docs Sync Handoff
- change:
- docs_sync_needed:
- promotion_candidates:
- must_read:
```

## conflict 応答
```md
### Conflict Handoff
- reason:
- detected_intent:
- why_this_direction_is_wrong:
- recommended_direction:
- next_prompt: `Use $<direction-skill> at skills/<direction-skill> to continue this request.`
```
