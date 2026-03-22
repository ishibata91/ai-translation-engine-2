# Fix Review Templates

`references/templates.md` は `fix-review` の review feedback field schema を定義する唯一の正本。

## review feedback
```md
### Review Feedback
- score:
- severity:
- location:
- violated_contract:
- required_delta:
- recheck:
- docs_sync_needed:
```

## Rules

- field は 7 個で固定し、追加 field を生やさない。
- `required_delta` には少なくとも `scope_failures` `external_validation_noise` `known_pre_existing_issue` の見出しを入れる。
- `recheck` には再実行コマンドと residual risk を分けて書く。
