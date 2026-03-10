## 1. 境界ルールと文書整備

- [ ] 1.1 `backend-quality-gates` と `slice-local-persistence-and-artifact-boundary` の spec 内容をレビューし、slice ローカル DB / artifact 境界の表現を確定する
- [ ] 1.2 `.golangci.yml` の `depguard` を 6 区分に対応する `files` ルールへ更新し、slice 配下の test code を含めて対象化する
- [ ] 1.3 `architecture.md` と `database_erd.md` の記述差分を洗い出し、artifact 用ストアと shared artifact 保存境界の追記内容を確定する

## 2. slice違反の分類と移行方針確定

- [ ] 2.1 `pkg/slice/**` の depguard 違反を `runtime` / `gateway` / `workflow` / `slice -> slice` に分類する
- [ ] 2.2 各違反について「slice ローカル永続化に残すもの」と「artifact へ移す共有データ」を仕分ける
- [ ] 2.3 `pkg/artifact/**` に共通検索契約を追加し、workflow が束ねる artifact 識別子を slice と 1:1 の単位で定義する

## 3. 実装修正

- [ ] 3.1 slice から `runtime` / `gateway` / `workflow` を直接 import している本番コードを、artifact 契約または workflow 経由の呼び出しへ置き換える
- [ ] 3.2 slice 間直接依存を持つコードを、共通検索契約を使う artifact 経由の共有データ参照へ置き換える
- [ ] 3.3 `database_erd.md` に artifact 用ストアを追記し、slice ローカル DB と shared artifact の保存境界を文書化する
- [ ] 3.4 slice 配下の test code を同じ境界へ合わせ、必要な結合検証を上位層または専用 integration test へ移す

## 4. 検証

- [ ] 4.1 変更対象ファイルに対して `npm run backend:lint:file -- <file...>` を繰り返し実行し、depguard を含む違反を段階的に解消する
- [ ] 4.2 `npm run lint:backend` を実行し、slice 境界違反が想定どおり検出・解消されていることを確認する
- [ ] 4.3 影響した slice の Go テストを実行し、artifact 経由への移行で回帰がないことを確認する
