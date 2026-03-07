# Proposal: Refactor Frontend VSA Pattern A

## Motivation
現在のフロントエンドにおける「肥大化したページコンポーネント（Fat Page）」の問題を解消し、バックエンドのInterface-First AIDD / VSAアーキテクチャに調和した保守性の高い構造へ移行するため。特に、ビジネスロジック（Wails API呼び出しやZustandストア連携）とUI描画が密結合している現状を改善し、将来的な機能拡張やAIによるコード生成・改修を容易にすることを目的とします。

## New Capabilities
- `frontend-headless-architecture`: UIの見た目（コンポーネント）とロジック（Wails呼出・状態管理）を完全に分離するHeadless Architecture（Custom Hooks偏重）の基礎構造を確立する。

## Modified Capabilities
- `master-persona-ui`: MasterPersona関連のページコンポーネント（約51KB）を分割し、UIとロジックを分離する形にリファクタリングする。
- `dictionary-builder-ui`: DictionaryBuilder関連のページコンポーネント（約33KB）を分割し、UIとロジックを分離する形にリファクタリングする。

## Impact
- `src/pages/MasterPersona.tsx` および `src/pages/DictionaryBuilder.tsx` のコードを大幅に削減し、純粋なPresentational層（UIの配線のみ）へ変更。
- 各機能特有の複雑なロジックをカプセル化したCustom Hook（例：`useMasterPersona`、`useDictionaryBuilder`）を導入。
- 共通利用される「見た目」のUIパーツを切り出し・集約する。
- Wails (`wailsjs/go/...`) のAPI呼び出しおよびZustand (`store/`) との通信が全てCustom Hooks内に隠蔽される。
- UIの見た目やユーザー機能自体には変更を加えない（純粋なリファクタリング）。
