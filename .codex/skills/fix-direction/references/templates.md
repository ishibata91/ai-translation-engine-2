# Bug Fix Templates

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

## 回帰確認
```md
### Verification
- 同じ再現手順で確認:
- 関連ケース:
- 実行した品質ゲート:
```

## Context Board Entry
```md
### Bugfix Handoff
- 固定した再現条件:
- 現在の原因仮説:
- 確定した観測事実:
- 次に起動する skill:
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
