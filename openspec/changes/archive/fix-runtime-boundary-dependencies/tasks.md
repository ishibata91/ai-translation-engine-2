## 1. 境界ルールと対象整理

- [x] 1.1 `runtime-execution-boundary` と `backend-quality-gates` の spec 内容をレビューし、runtime が保持すべき責務と外へ出す責務を確定する
- [x] 1.2 `.golangci.yml` の runtime 向け `depguard` ルールを確認し、`pkg/runtime/**` の本番コードと test code に適用される状態を整える
- [x] 1.3 `pkg/runtime/**` の違反を `workflow` / `slice` / `artifact` 依存に分類する

## 2. 設計移行

- [x] 2.1 workflow との連携を共通 executor 契約と中立 DTO へ寄せる対象を洗い出す
- [x] 2.2 runtime が持つ slice 固有判断を workflow 側へ戻す対象を洗い出す
- [x] 2.3 runtime が gateway だけへ依存する構成に必要な contract と DTO の粒度を確定する

## 3. 実装修正

- [x] 3.1 runtime 本番コードの workflow / slice / artifact 直接依存を、中立契約と gateway 利用へ置き換える
- [x] 3.2 runtime 配下の test code を同じ境界へ合わせ、必要な結合検証を上位層へ移す
- [x] 3.3 変更後の runtime 入出力 DTO が技術的結果だけを返す構成に保たれていることを確認する

## 4. 検証

- [x] 4.1 変更対象ファイルに対して `npm run backend:lint:file -- <file...>` を繰り返し実行し、runtime 境界違反を段階的に解消する
- [x] 4.2 `npm run lint:backend` を実行し、runtime 境界違反が想定どおり検出・解消されていることを確認する
- [x] 4.3 影響した runtime の Go テストを実行し、実行制御経路の回帰がないことを確認する
