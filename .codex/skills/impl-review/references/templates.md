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
- レビュー前に、仕様適合、差分の危険箇所、テスト不足、例外・失敗時、既存設計との整合性、セキュリティ・性能の 6 観点を必ず確認する。
- `score` は「欠陥の重さ」を表す離散バンドとし、`1.00 / 0.90 / 0.85 / 0.75 / 0.50` のいずれかで返す。
- `score >= 0.85` を通過条件とし、`critical` `medium` `required verification 不足` は `0.75` 以下に落とす。
- `low` は件数帯で評価し、1-2 件は `0.90`、3-4 件は `0.85`、5 件以上は `0.75` とする。
- `external_validation_noise` または `known_pre_existing_issue` だけが残る場合の score 上限は `0.90` とする。
- `docs_sync_needed` は score に影響させず、docs handoff 判断専用で使う。
- packet 生成後は `.codex/skills/scripts/validate-packet-contracts.ps1` を実行し、validator fail 時は 1 回だけ自己再試行する。
- frontend 差分がある場合、`recheck.commands` には `npm --prefix frontend run build` を含める。
- frontend 差分があるのに build 結果が無い、または失敗している場合は最低でも `medium` として扱い、`score` は `0.75` 以下にする。
