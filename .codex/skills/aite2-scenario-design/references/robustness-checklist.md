# Robustness Checklist

## 最低限分けるもの
- Actor
- Boundary
- Control
- Entity

## 描くときの確認
- 1 図 1 シナリオになっているか
- Main Flow が中心になっているか
- Actor に内部コンポーネントを置いていないか
- Boundary は画面やダイアログなど UI 境界になっているか
- Control は判断や進行制御の責務になっているか
- Entity は共有成果物や業務上意味のある入出力になっているか

## 避けること
- UI の見た目詳細を書く
- 実装アルゴリズムや if 分岐の全列挙を書く
- runtime / gateway の詳細を広げすぎる
- `logic.md` の責務まで先に書く
