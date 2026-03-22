# Frontend Implementation Quality Checklist

## 必読
- `docs/frontend/frontend-architecture/spec.md`
- `docs/frontend/frontend-coding-standards/spec.md`

## 品質ゲート
- `lint:file -> 修正 -> 再実行`
- `typecheck`
- `lint:frontend`
- `Playwright`

## 停止条件
- 品質ゲートを通すために `owned_paths` 外の修正が必要なら、その修正は行わず blocked を返す
- 別 section の責務や未固定 contract が見えた場合は、品質ゲート完遂より先に停止して `impl-direction` へ返す
