# Investigation Direction Templates

## 下流スキル起動
```md
### Skill Invocation
- invoked_skill: <起動する下流スキル名 (例: investigation-distill)>
- invoked_by: investigation-direction
- agent: <起動するエージェント名 (例: ctx_loader)>
- task:
- 入力:
- focus:
```

> 下流スキルのサブエージェントを立ち上げるときは必ずこのテンプレートを使い、
> `invoked_skill` と `invoked_by` を明示すること。
> サブエージェントは自分がどのスキルで起動されたかをこのフィールドで確認する。

## conflict 応答
```md
### Conflict Handoff
- reason:
- detected_intent:
- why_this_direction_is_wrong:
- recommended_direction:
- next_prompt: `Use $<direction-skill> at skills/<direction-skill> to continue this request.`
```
