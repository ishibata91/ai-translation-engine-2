# Design Sequence Checklist

## まず見ること
- 既存見た目修正か
- 新規 UI 設計か
- シナリオ未確定か
- ロジック境界未確定か
- docs 同期フェーズか

## 基本順序
- UI を含む新規変更: `plan-distill` -> `plan-ui` -> `plan-scenario` -> `plan-logic` -> `plan-review`
- UI を含まない変更: `plan-distill` -> `plan-scenario` -> `plan-logic` -> `plan-review`
- 既存見た目修正: conflict として `impl-direction` へ handoff
- docs 反映: `plan-distill` -> `plan-sync`

## 注意点
- 各 stage は順番に実行し、packet や artifact が揃う前に次の stage を飛ばさない
- stage ごとの確定事項と未確定事項を保持し、review 前に取りこぼしを残さない
- 深い設計内容は各専用 skill に委譲するが、`plan-direction` 自体は chain を最後まで完走させる
