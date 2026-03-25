件数が10件から先に進まない
ページングが機能してない

[$fix-direction](F:\ai translation engine 2\.codex\skills\fix-direction\SKILL.md)

## 不具合概要
- **画面/機能**: 翻訳プロジェクト/ペルソナ生成フェーズ
- **現象**: ペルソナ生成が始まらない

## 再現手順
前提: 単語翻訳済み
1. モデルをLMstudioにする
2. 同期でペルソナ生成開始

再現率: 毎回

## 期待挙動
ペルソナ生成が始まり、終了後にNPC一覧に反映される
## 実際の挙動
処理中にはなるが、モデルはロードされず、何も始まらない。
## 補足
- ログ: 
- 波及の可能性:なし 
- エラー表示: なし
- スクリーンショット:なし 

---

[$plan-direction](F:\ai translation engine 2\.codex\skills\plan-direction\SKILL.md)

## 設計・仕様策定の概要
- **対象機能/画面**: 翻訳プロジェクト/ペルソナ生成フェーズ
- **目的・背景**: ペルソナ生成フェーズのNPC一覧をマスターペルソナ生成の一覧と同じにする

## 要求事項
1. ペルソナ生成フェーズのNPC一覧とマスターペルソナ生成の一覧が同じ見た目になること
2. 同じコンポーネントを共有すること

## 制約・前提条件


## 補足資料
- 関連資料/issue: 
- docs\slice\persona 既存仕様
- 

---

[$impl-direction](F:\ai translation engine 2\.codex\skills\impl-direction\SKILL.md)

## 実装対象の概要
- **対象の変更文書 (changes)**: changes/persona-persona-phase-shared-npc-list
- **目的**: ペルソナ生成フェーズのNPC一覧をマスターペルソナ生成の一覧と同じにする
skillに厳格に従って進める事｡
## 実装スコープ
- **対象領域**: Frontend
Mixed
## 前提・制約事項
- **共有コントラクト**: 
- **所有範囲 (owned paths)**: 
- **変更禁止範囲 (forbidden paths)**: 

## 補足・懸念事項
- 
