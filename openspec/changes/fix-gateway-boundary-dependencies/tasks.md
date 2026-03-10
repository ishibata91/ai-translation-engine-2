## 1. 境界ルールと対象整理

- [ ] 1.1 `gateway-io-boundary` と `backend-quality-gates` の spec 内容をレビューし、gateway が保持すべき責務と外へ出す責務を確定する
- [ ] 1.2 `.golangci.yml` の gateway 向け `depguard` ルールを確認し、`pkg/gateway/**` の本番コードと test code に適用される状態を整える
- [ ] 1.3 `pkg/gateway/**` の違反を `workflow` / `runtime` / `slice` / `artifact` 依存に分類する

## 2. 設計移行

- [ ] 2.1 gateway が持つ上位層知識を runtime または controller 側へ戻す対象を洗い出す
- [ ] 2.2 gateway 返却値を provider 横断の最小中立 DTO へ揃え、workflow が mapper で最終 DTO へ変換する責務分担を整理する
- [ ] 2.3 gateway を純技術接続実装へ寄せるため、provider ごとの helper と contract の境界を確定する

### 設計メモ

- `pkg/gateway/configstore`: SQLite 実装、migration、store contract を置く
- `pkg/runtime/configaccess`: TypedAccessor や実行時読取補助を置く
- `pkg/workflow/promptsettings` または `pkg/workflow/personasettings`: workflow 固有の default / 解釈を置く

## 3. 実装修正

- [ ] 3.1 gateway 本番コードの workflow / runtime / slice / artifact 直接依存を、純技術接続実装へ置き換える
- [ ] 3.2 gateway 配下の integration test を削除し、gateway 配下には境界内テストだけを残す
- [ ] 3.3 変更後の gateway 返却型が中立 DTO に保たれていることを確認する

## 4. 検証

- [ ] 4.1 変更対象ファイルに対して `npm run backend:lint:file -- <file...>` を繰り返し実行し、gateway 境界違反を段階的に解消する
- [ ] 4.2 `npm run lint:backend` を実行し、gateway 境界違反が想定どおり検出・解消されていることを確認する
- [ ] 4.3 影響した gateway の Go テストを実行し、外部接続経路の回帰がないことを確認する
- [ ] 4.4 後続の API テスト整備 change で削除した integration test を置き換える前提が review / task に残っていることを確認する
