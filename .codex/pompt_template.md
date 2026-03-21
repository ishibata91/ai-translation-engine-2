[$fix-direction](F:\ai translation engine 2\.codex\skills\fix-direction\SKILL.md)

## 不具合概要
- **画面/機能**: 翻訳プロジェクト/単語翻訳フェーズ
- **現象**: 

## 再現手順
前提: 
1. 
2. 
3. 

再現率: 毎回

## 期待挙動

## 実際の挙動

## 補足
- ログ: 
- 波及の可能性: 
- エラー表示: 
- スクリーンショット: 

---

[$plan-direction](F:\ai translation engine 2\.codex\skills\plan-direction\SKILL.md)

## 設計・仕様策定の概要
- **対象機能/画面**: 翻訳プロジェクト/単語翻訳フェーズ
- **目的・背景**: 単語翻訳フェーズの実行時､辞書構築で作成したアーティファクトから､参考単語を抽出しシステムプロンプトに含める

## 要求事項
1. 辞書構築で作成したマスター辞書から参考単語を抽出できる｡
2. 完全一致がある場合は辞書の日本語で確定で置き換えること｡
3. 人名部分一致の場合は参考単語としてプロンプトに含めること｡
4. 文字列が既に日本語の場合､単語翻訳対象として含めず､Requestの構築､辞書索引､翻訳を行わないこと｡


## 制約・前提条件
- 辞書構築実施済み

## 補足資料
- 関連資料/issue: 
- docs\slice\terminology 既存仕様
- 

---

[$impl-direction](F:\ai translation engine 2\.codex\skills\impl-direction\SKILL.md)

## 実装対象の概要
- **対象の変更文書 (changes)**: 
- **目的**: 

## 実装スコープ
- **対象領域**: [Frontend / Backend / Mixed]

## 前提・制約事項
- **共有コントラクト**: 
- **所有範囲 (owned paths)**: 
- **変更禁止範囲 (forbidden paths)**: 

## 補足・懸念事項
- 
