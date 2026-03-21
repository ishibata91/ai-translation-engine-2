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

## task 分割（旧 work order）
impl-distill の packet 返却後、worker を起動する前に必ずこのテンプレートで記録する。

```md
### Task Split
- change:
- routing: frontend | backend | mixed
- shared_contract:          # mixed 時のみ。型定義・API 契約を worker 起動前に確定する
- workers:
  - owner: frontend | backend
    target_task:
    owned_paths:            # このワーカーが書いてよいパス
    forbidden_paths:        # このワーカーが触れてはいけないパス
    required_reading:       # 読む必要のある artifact
    validation_commands:    # typecheck / lint:frontend など
  - owner: ...              # 必要なだけ繰り返す
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
