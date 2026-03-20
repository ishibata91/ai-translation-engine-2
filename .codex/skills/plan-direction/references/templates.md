# Design Orchestrator Templates

## 下流スキル起動
```md
### Skill Invocation
- invoked_skill: <起動する下流スキル名 (例: plan-distill)>
- invoked_by: plan-direction
- agent: <起動するエージェント名 (例: ctx_loader)>
- 対象 change:
- 入力:
- focus:
```

> 下流スキルのサブエージェントを立ち上げるときは必ずこのテンプレートを使い、
> `invoked_skill` と `invoked_by` を明示すること。
> サブエージェントは自分がどのスキルで起動されたかをこのフィールドで確認する。

## context packet 依頼
```md
### Context Request
- 対象 change:
- 現在の依頼:
- 読むべき changes:
- 読むべき docs:
- 読むべきコード:
- 今回の焦点:
```

## handoff 更新
```md
### Handoff Update
- 現在の担当 role:
- 次に起動する skill:
- board で読むべきファイル:
- 完了条件:
- 未確定事項:
```

## chain 実行更新
```md
### Chain Update
- current_stage:
- completed_stages:
- remaining_stages:
- current_artifacts:
- blockers:
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
