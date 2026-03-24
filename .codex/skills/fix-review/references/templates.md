# Fix Review Templates

`references/templates.md` は `fix-review` の review feedback field schema を定義する唯一の正本。

## review feedback
```json
{
  "score": 0.0,
  "severity": "",
  "location": [],
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

- 正本は `changes/<id>/context_board/fix-review.feedback.json` とし、validation は同じ場所の `fix-review.feedback.validation.json` を使う。
- field は 7 個で固定し、追加 field を生やさない。
- `required_delta` には少なくとも `scope_failures` `external_validation_noise` `known_pre_existing_issue` の見出しを入れる。
- `recheck` には再実行コマンドと residual risk を分けて書く。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、validator fail 時は 1 回だけ自己再試行する。
