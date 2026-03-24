# Impl Review Templates

## review feedback
```json
{
  "score": 0.0,
  "severity": "",
  "location": [],
  "affected_sections": [],
  "violated_contract": [],
  "required_delta": {
    "scope_failures": [],
    "external_validation_noise": [],
    "known_pre_existing_issue": []
  },
  "recheck": {
    "commands": [],
    "residual_risk": []
  },
  "docs_sync_needed": false
}
```

## Rules

- 正本は `changes/<id>/context_board/impl-review.feedback.json` とし、validation は同じ場所の `impl-review.feedback.validation.json` を使う。
- field は `score` `severity` `location` `affected_sections` `violated_contract` `required_delta` `recheck` `docs_sync_needed` の 8 個で固定する。
- `reviewer_actions` のような追加 field を生やさない。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、validator fail 時は 1 回だけ自己再試行する。
- frontend 差分がある場合、`recheck.commands` には `npm --prefix frontend run build` を含める。
