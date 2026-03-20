# Impl Direction Templates

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
