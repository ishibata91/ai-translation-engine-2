# Impl Direction Templates

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
- tasks.md:
- issues:
- next_action:
```

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
