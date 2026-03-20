# Implementation Routing Checklist

## frontend 実装へ送る条件
- 主変更が `frontend/src` 配下
- `ui.md` または frontend spec が主判断材料
- 品質ゲートが `typecheck / lint:frontend / Playwright`

## backend 実装へ送る条件
- 主変更が `pkg/` や controller / workflow / slice / runtime / artifact / gateway
- `logic.md` や architecture spec が主判断材料
- 品質ゲートが `backend:lint:file / lint:backend`

## 混在時
- 先に task を frontend / backend へ分割する
- 1 つの skill に両方の品質ゲートを背負わせない
