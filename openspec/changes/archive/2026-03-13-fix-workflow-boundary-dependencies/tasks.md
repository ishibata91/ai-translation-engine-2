## 1. 境界ルールと対象整理

- [x] 1.1 `workflow-orchestration-boundary` と `backend-quality-gates` の spec 内容をレビューし、workflow が保持すべき責務と外へ出す責務を確定する
- [x] 1.2 `.golangci.yml` の workflow 向け `depguard` ルールを確認し、`pkg/workflow/**` の本番コードと test code に適用される状態を整える
- [x] 1.3 `pkg/workflow/**` の違反を `gateway` / `controller` / `artifact` 依存に分類する

## 2. 設計移行

- [x] 2.1 workflow が直接持つ gateway 依存を runtime 契約経由へ移す対象を洗い出す
- [x] 2.2 workflow が直接持つ UI 境界依存を controller 側へ戻す対象を洗い出す
- [x] 2.3 workflow が共有データ本体を保持している箇所を、artifact 識別子・検索条件管理へ置き換える方針を確定する

## 3. 実装修正

- [x] 3.1 workflow 本番コードの gateway / controller / artifact 直接依存を、runtime 契約・controller 境界・artifact ref 管理へ置き換える
- [x] 3.2 workflow 配下の test code を責務ごとに整理し、必要な結合検証を上位層または専用 integration test へ移す
- [x] 3.3 変更後の workflow 入出力 DTO と contract が orchestration 専用に保たれていることを確認する

## 4. 検証

- [x] 4.1 変更対象ファイルに対して `npm run backend:lint:file -- <file...>` を繰り返し実行し、workflow 境界違反を段階的に解消する
- [x] 4.2 `npm run lint:backend` を実行し、workflow 境界違反が想定どおり検出・解消されていることを確認する
- [x] 4.3 影響した workflow の Go テストを実行し、orchestration 経路の回帰がないことを確認する
