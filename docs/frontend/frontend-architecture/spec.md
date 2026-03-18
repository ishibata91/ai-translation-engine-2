# フロントエンド アーキテクチャ (UI Architecture)

> **Frontend Headless Architecture v1 準拠**
> Wails + React 環境における、ロジックとUI表示の完全分離戦略

---

## 1. アーキテクチャの目的 (Objectives)

現在のフロントエンドにおける「肥大化したページコンポーネント（Fat Page）」の問題を解消し、バックエンドの **Interface-First AIDD / Vertical Slice Architecture (VSA)** に調和した保守性の高い構造へ移行する。

*   **肥大化の解消**: WailsのAPI呼び出し、Zustandからの状態取得、およびUIのDOM構築（JSX）が一つのファイルに混在することを防ぐ。
*   **AI生成・改修の最適化**: UI（見た目）とロジック（動作）の責任境界を明確化し、リファクタリングや機能追加時のデグレ（意図しない破壊）を防ぐ。
*   **モジュール結合の粗結合化**: 異なる画面間の状態・型の密結合を排除し、VSA的な「スライスごとの完結（State, Action, API Binding）」を実現する。

---

## 2. コア・パターン：Headless Component Architecture (Pattern A)

本プロジェクトのフロントエンドでは、**「Headless Architecture（Custom Hooks偏重型）」** を採用する。

### 2.1 UI層（Presentational）とロジック層（Container/Hooks）の完全分離
コンポーネント（特にPageコンポーネント）には「Wailsとの通信」や「複雑なローカルステートの管理」を一切持たせず、すべての振る舞いと状態管理を **機能ごとの Custom Hook** に隠蔽する。

*   **ロジック層 (`src/hooks/features/...`)**:
    Wails (`wailsjs/go/...`) の API 呼び出し、Zustand ストアとの通信、イベント購読などの副作用をすべて一手に引き受ける。
*   **UI層 (`src/pages/...`)**:
    Hook から返却されるデータ（状態）とコールバック関数のみを利用して DOM 構造 (JSX) を組み立てる。純粋な「配線（Wiring）」と「描画」のみに徹する。

### 2.2 依存関係のルール (Dependency Rules)
1.  **PageはHookにのみ依存する**: `MasterPersona.tsx` のようなページコンポーネントは、直接 Wails API を `import` してはならない。必ず `useMasterPersona` を経由する。
2.  **型（DTO）の局所化**: その機能（スライス）でしか使わない型インターフェース（APIレスポンスの型など）は、グローバルな `src/types/` ではなく `src/hooks/features/[featureName]/types.ts` 等へ移動し、他機能からの再利用を意識しない（バックエンドの WET 原則に準ずる）。

---

## 3. ディレクトリ構造とスライス (Directory Structure)

バックエンドの各機能（スライス）に対応するように、フロントエンドでも機能単位のロジックを集約する。

```text
src/
 ├─ components/             <- 全画面で使い回す汎用UIパーツ (Button, DataTable, Modal など)
 │   └─ ui/                 <- 汎用的な見た目のみを定義する
 │
 ├─ hooks/
 │   └─ features/           <- 🌟 バックエンドのSliceに対応するロジック群
 │       ├─ masterPersona/
 │       │   ├─ useMasterPersona.ts  <- 通信・状態管理などのすべて
 │       │   └─ types.ts             <- MasterPersona専用の型定義
 │       │
 │       └─ dictionaryBuilder/
 │           ├─ useDictionaryBuilder.ts
 │           └─ types.ts
 │
 ├─ store/                  <- グローバルなUI状態のみ (サイドバー開閉など)
 │   └─ uiStore.ts
 │
 └─ pages/                  <- ルーティングに乗る「ページ（画面）」
     ├─ MasterPersona.tsx         <- 非常に薄い(JSX配線のみの)コンポーネント
     └─ DictionaryBuilder.tsx     <- 同上
```

---

## 4. Hook のインターフェース (Contract as Hook)

バックエンドが「Interface-First」であるように、フロントエンドにおいては **Custom Hook が返すオブジェクトの型（戻り値）が、UIに対する「Contract（契約）」** となる。AIが実装・リファクタを行う際は、まずこの「Hookが何を返すか」を設計の第一歩とする。

### 実装例のイメージ
```typescript
// src/hooks/features/masterPersona/useMasterPersona.ts
export const useMasterPersona = () => {
    // ... 複雑な Wails API 呼び出し、Zustand 監視、useEffect ...
    
    return {
        // UIに描画させるデータ
        pagedNpcData: [...],
        isGenerating: false,
        progressPercent: 45,
        statusMessage: "生成中...",
        
        // UIから発火させるアクション
        handleStart: () => { ... },
        handlePause: () => { ... },
        handlePageChange: (page) => { ... }
    };
};
```

このように設計することで、「見た目をTailwindで派手にしたい（`MasterPersona.tsx` の修正）」タスクと、「Wailsの新しいAPIを生やして呼び出したい（`useMasterPersona.ts` の修正）」タスクが物理的に分離され、AI駆動開発における精度の向上とコンフリクトの回避を実現する。
