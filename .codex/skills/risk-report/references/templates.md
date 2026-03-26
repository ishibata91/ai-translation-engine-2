# Risk Report Templates

## 下流スキル起動
```md
### Skill Invocation
- invoked_skill: risk-report
- invoked_by: <impl-direction | fix-direction>
- agent: review_cycler
- change_id:
- lane: <impl | fix>
- diff_range:
- focus:
```

## Risk Report
```md
# Risk Report (<lane>)

## Metadata
- change_id:
- generated_at:
- diff_range:
- total_changed_files:

## Overall
- summary:
- risk_level: low | medium | high

## Risk Items
- id: R1
  category: 仕様逸脱 | 回帰可能性 | テスト不足 | 運用影響
  risk:
  diff_evidence:
    - `<path>:<line-or-hunk>`
  impact:
  likelihood: low | medium | high
  mitigation:

## Open Questions
- question:
  needed_evidence:

## Recommended Follow-up
- owner:
- action:
- due_hint:
```
