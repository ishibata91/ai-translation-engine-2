# Design Sequence Checklist

## まず見ること
- 既存見た目修正か
- 新規 UI 設計か
- シナリオ未確定か
- ロジック境界未確定か
- docs 同期フェーズか

## 基本順序
- UI を含む新規変更: `plan-ui` -> `plan-scenario` -> `plan-logic`
- UI を含まない変更: `plan-scenario` -> `plan-logic`
- 既存見た目修正: `aite2-ui-polish-orchestrate`
- docs 反映: `plan-sync`

## 注意点
- 一度に複数 skill を深掘りしない
- 次の skill へ渡す未確定事項を明確にする
- 深い設計内容は各専用 skill に委譲する
