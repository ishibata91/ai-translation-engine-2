## 1. 設定仕様と保存経路の整備

- [ ] 1.1 MasterPersona prompt 用 namespace と既定値の扱いを `config` 実装へ反映する
- [ ] 1.2 `MasterPersona.tsx` で prompt 設定を hydrate/save する state と差分保存処理を追加する

## 2. MasterPersona UI の拡張

- [ ] 2.1 `MasterPersona.tsx` に編集可能なユーザープロンプト入力カードを追加する
- [ ] 2.2 `MasterPersona.tsx` に読み取り専用のシステムプロンプト表示カードを追加する

## 3. Prompt 責務分離の実装

- [ ] 3.1 既存ユーザープロンプト内の固定補完値を洗い出し、system prompt 側へ移設する
- [ ] 3.2 画面表示中の user/system prompt と実際の送信 prompt の責務が一致するよう組み立て処理を調整する

## 4. 検証

- [ ] 4.1 MasterPersona 画面で prompt の保存・再読込・readOnly 表示を確認する
- [ ] 4.2 必要なテストまたは動作確認を追加し、既存 LLM 設定保存との回帰がないことを確認する
