## 概要

- 

## 確認事項

- [ ] `npm run backend:check` を実行し、整形差分なしと `go test ./pkg/...` の通過を確認した
- [ ] `npm run backend:lint` を実行し、新規違反を持ち込んでいないことを確認した
- [ ] 開発中はローカルで上記コマンドを適宜回し、最後に再実行した
- [ ] 公開型・公開関数・公開メソッドに doc コメントを付与した
- [ ] `context.Context` を公開入口から内部処理まで伝播させた
- [ ] `error` に呼び出し文脈を付与し、握り潰しをしていない
- [ ] ログは `slog.*Context` を使い、解析に必要な構造化キーを付与した
- [ ] Contract / DTO / Mapper の責務境界が `openspec/specs/architecture.md` に一致している
- [ ] Table-Driven Test を基本とし、必要な並行処理テストに `goleak` を適用した
- [ ] `govulncheck` が必要な変更の場合、`npm run backend:vuln` の実行有無を説明した

## レビュー観点メモ

- MUST 違反がある場合は理由、影響範囲、解消予定を明記すること
- SHOULD 違反を見送る場合はレビューコメントまたは関連タスクに判断理由を残すこと
