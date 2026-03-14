## 1. Spec と配置方針の確定

- [x] 1.1 `openspec/changes/per-32-reorganize-format-parser-exporter/specs/format/spec.md` の要求に合わせて、`pkg/format` を外部フォーマット境界として扱う実装方針を確定する
- [x] 1.2 `openspec/changes/per-32-reorganize-format-parser-exporter/specs/export/spec.md` と `openspec/changes/per-32-reorganize-format-parser-exporter/specs/slice/parser/spec.md` に沿って、`Parser` / `Exporter` 契約名を維持したまま移設する対象ファイルを棚卸しする

## 2. Parser の format 配下への移設

- [x] 2.1 `pkg/slice/parser` の実装を `pkg/format/parser/skyrim` 相当へ移設し、Skyrim 固有 parser として package 構成を再編する
- [x] 2.2 workflow / controller / provider からの parser import を新 package へ切り替え、`Parser` 契約名のまま解決できるようにする
- [x] 2.3 旧配置前提の [cmd/parser](/F:/ai%20translation%20engine%202/cmd/parser) を削除し、不要な互換 import を残さない
- [x] 2.4 parser 関連の変更ファイルに対して `npm run backend:lint:file -- <file...>` を実行し、指摘を解消して再実行する

## 3. Exporter の format 配下への移設

- [x] 3.1 `pkg/gateway/export` の実装を `pkg/format/exporter/xtranslator` 相当へ移設し、xTranslator 固有 exporter として package 構成を再編する
- [x] 3.2 workflow / provider からの exporter import を新 package へ切り替え、`Exporter` 契約名のまま解決できるようにする
- [x] 3.3 exporter の既存テスト資産を format 配下へ移し、責務名に合わせてテスト名と fixture 名を更新する
- [x] 3.4 exporter 関連の変更ファイルに対して `npm run backend:lint:file -- <file...>` を実行し、指摘を解消して再実行する

## 4. 統合確認と仕上げ

- [x] 4.1 旧 package 参照が残っていないことを確認し、必要なら関連ドキュメントやコメントを新配置に合わせて更新する
- [x] 4.2 関連する Go テストを実行して、Skyrim 抽出 JSON と xTranslator XML の互換を維持できていることを確認する
- [x] 4.3 `npm run lint:backend` を実行してバックエンド品質ゲートを通す
- [x] 4.4 変更結果を `review.md` に記録し、未実行の品質ゲートがあれば理由を明記する
- [x] 4.5 OpenSpec の既存 spec 配置を見直し、`openspec/specs/slice/parser` を `openspec/specs/format/parser` へ、`openspec/specs/export` を `openspec/specs/format/export` へ移して format 境界に揃える
