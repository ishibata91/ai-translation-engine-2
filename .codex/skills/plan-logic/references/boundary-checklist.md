# Logic Boundary Checklist

## 粒度
- 1 文書 1 主シナリオになっているか
- 画面全体や機能全体を一度に扱っていないか
- 関数単位まで細かくしすぎていないか

## 責務境界
- controller が進行決定していないか
- workflow が orchestration に集中しているか
- runtime が外部 I/O 実行に集中しているか
- slice が個別業務ロジックに集中しているか
- artifact が handoff 保存境界に留まっているか

## 避けること
- メソッド分解の詳細を書く
- DTO フィールド一覧の詳細を書く
- docs 正本へ同期すべき長期仕様まで書く
- 具象依存を無造作に広げる
