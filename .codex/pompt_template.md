件数が10件から先に進まない
ページングが機能してない

[$fix-direction](F:\ai translation engine 2\.codex\skills\fix-direction\SKILL.md)

## 不具合概要
- **画面/機能**: 翻訳プロジェクト/単語翻訳フェーズ
- **現象**: 自動置き換え時､失敗として扱われる｡

## 再現手順
前提: データロード済み
1. 単語翻訳画面に遷移する
2. Gemini 同期実行をモデルに指定して翻訳実行
3. 失敗件数を参照

再現率: 毎回

## 期待挙動
12件中12件成功､失敗0と表示される
## 実際の挙動
3件失敗となる｡いずれもマスター辞書構築にて､完全一致置き換えがされたもの｡
## 補足
- ログ: 
- 波及の可能性:なし 
- エラー表示: なし
- スクリーンショット:なし 

---

[$plan-direction](F:\ai translation engine 2\.codex\skills\plan-direction\SKILL.md)

## 設計・仕様策定の概要
- **対象機能/画面**: 翻訳プロジェクト/ペルソナ生成フェーズ
- **目的・背景**: 翻訳対象に登場するNPCの口調などの､ペルソナを生成する｡

## 要求事項
1. 翻訳対象に登場するNPCのペルソナを生成できること
2. 既存のマスターペルソナに含まれているNPCは生成対象としないこと


## 制約・前提条件


## 補足資料
- 関連資料/issue: 
- docs\slice\persona 既存仕様
- 

---

[$impl-direction](F:\ai translation engine 2\.codex\skills\impl-direction\SKILL.md)

## 実装対象の概要
- **対象の変更文書 (changes)**: translation-flow-terminology-dictionary-reference-rules
- **目的**: ペルソナ生成フェーズを実装する

## 実装スコープ
- **対象領域**: [Frontend / Backend / Mixed]
Mixed
## 前提・制約事項
- **共有コントラクト**: 
- **所有範囲 (owned paths)**: 
- **変更禁止範囲 (forbidden paths)**: 

## 補足・懸念事項
- 
