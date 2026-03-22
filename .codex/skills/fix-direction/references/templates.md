# Bug Fix Templates

## 下流スキル起動
```md
### Skill Invocation
- invoked_skill: <起動する下流スキル名 (例: fix-distill)>
- invoked_by: fix-direction
- agent: <起動するエージェント名 (例: ctx_loader)>
- operation: <`fix-logging` を起動する場合のみ `add` または `remove`>
- 対象 change:
- 入力:
- focus:
```

> 下流スキルのサブエージェントを立ち上げるときは必ずこのテンプレートを使い、
> `invoked_skill` と `invoked_by` を明示すること。
> サブエージェントは自分がどのスキルで起動されたかをこのフィールドで確認する。
> `fix-logging` を起動する場合は、このテンプレートに加えて `../fix-logging/references/templates.md` の add/remove packet を併記し、`[fix-trace]` prefix と `log_additions` / `log_removals` 契約を明示する。

## 再現メモ
```md
### Repro
- 前提:
- 手順:
- 実挙動:
- 再現率:
```

## 期待挙動 / 実挙動
```md
### Expected
- 

### Actual
- 
```

## 仕様照合メモ
```md
### Spec Check
- 参照 spec:
- 関連箇所:
- spec の期待:
- 実装との差分:
```

## 原因切り分けメモ
```md
### Triage
- 実装不備:
- 仕様が古い:
- 仕様が曖昧:
- 仕様同士の矛盾:
- 次の 1 手:
```

## 修正方針
```md
### Fix Plan
- 変更対象:
- 最小修正方針:
- 仕様修正の有無:
- 回帰確認方法:
```

## 状態要約
```md
### State Summary
- current_stage:
- reproduction_status:
- known_facts:
- unknowns:
- active_logs:
- current_hypothesis:
- current_scope:
- latest_review_result:
- next_action:
```

> `State Summary` は full history の代替として `fix-direction` が持つ正本。
> 再現待ち、review loop、cleanup 状態をこの形式で引き継ぐ。

## 回帰確認
```md
### Verification
- 同じ再現手順で確認:
- 関連ケース:
- 実行した品質ゲート:
```

## review 判定
```md
### Review Decision
- review_score:
- unresolved_fix_scope:
- external_validation_noise:
- residual_risks:
- next_action:
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
