# フロントエンド コーディング規約

## 1. 目的

- フロントエンドの品質を安定化し、リファクタ時のデグレを抑える。
- `openspec/specs/architecture.md` と `openspec/specs/frontend_architecture.md` の方針に沿って、責務分離と保守性を担保する。

## 2. 適用範囲

- `frontend/src` 配下の TypeScript / React コード全体。
- 対象: `pages`, `hooks`, `components`, `store`, `types`, `wailsjs` 境界利用箇所。

## 3. 一般品質ルール

### 3.1 型定義と境界の厳格化（Schema-Driven）

- TypeScript は `strict` 前提とし、`any` の使用を禁止する。
- 外部 API や Wails 層からの境界値・戻り値は `unknown` で受け、`zod` または `valibot` によるランタイムパース（型ガード）を行う。
- 型定義（`type` / `interface`）はコンポーネントから分離し、データ構造を把握しやすい単位で集約する。

### 3.2 状態管理と非同期処理

- 非同期処理のフェーズ（`idle / loading / success / error`）は明示的な状態として管理する。
- `catch` 内でのエラー握りつぶしを禁止する。必ずグローバルハンドラ、または上位 UI 層へエラーオブジェクトを伝播させる。
- グローバル状態への依存を最小化し、コンポーネントは可能な限り Props 駆動で設計する。

### 3.3 副作用と制御フロー

- 副作用は `useEffect` に閉じ、イベント購読・タイマー処理は必ずクリーンアップ関数を提供する。
- `useEffect` の依存配列による多段発火（副作用連鎖）を禁止する。ロジックは純粋関数へ分離し、`useEffect` は状態同期に限定する。
- Early Return を徹底し、ネストを浅く保つ。

### 3.4 命名規則と自己文書化

- 命名規則を統一する。
- Hook: `useXxx`
- イベントハンドラ: `handleXxx`（Props コールバックは `onXxx`）
- 真偽値: `isXxx` / `hasXxx` / `shouldXxx` / `canXxx`
- 主要な `interface`、関数、コンポーネントには TSDoc/JSDoc コメントを記述する。

### 3.5 責務の分離とファイル分割

- UI コンポーネントは表示責務に限定し、データ取得、Wails 呼び出し、ビジネスロジックを直接記述しない。
- 1 関数 1 責務（SRP）を原則とし、オーケストレーター以外は読解負荷を最小化する。

## 4. このリポジトリ固有ルール

### 4.1 Headless Architecture 準拠

- `pages` は `hooks/features/*` が返す state/action を利用して描画に専念する。
- Wails API (`wailsjs/go/...`) は `hooks/features/*` からのみ呼び出す。
- 機能専用型は `src/types` ではなく `src/hooks/features/<feature>/types.ts` に配置する。

### 4.2 インポート境界ルール

- `pages` から `wailsjs` の直接 import を禁止する。
- `pages` から `store` の直接 import を禁止する（必要な場合は feature hook 経由）。
- 境界ルールは ESLint で機械検証する。
- 第一候補: `eslint-plugin-import` の `import/no-restricted-paths`
- 必要に応じて: `eslint-plugin-boundaries`

### 4.3 状態管理ルール

- グローバル UI 状態のみ Zustand (`store`) を使う。
- 機能固有の一時状態は feature hook 内の `useState` で管理する。
- Hook の戻り値は「state / action / selector」に論理分割する。

### 4.4 Wails 境界ルール

- Wailsレスポンスの整形（snake_case/camelCase変換、日付整形、ステータス変換）は adapter 関数で一元化する。
- UI層で `as any` による直接吸収は禁止する。

## 5. 推奨ツール（デファクト）

- Lint: `eslint`, `typescript-eslint`, `eslint-plugin-import`, `eslint-plugin-react-hooks`, `eslint-plugin-react-refresh`, `eslint-plugin-tsdoc`, `eslint-plugin-jsdoc`
- Format: `prettier`, `eslint-config-prettier`
- Test: `vitest`, `@testing-library/react`, `@testing-library/user-event`, `msw`
- Runtime Validation: `zod`（Wails境界の入力/出力検証）

## 6. lint 化ポリシー

### 6.1 lint で強制する項目

- `pages` から `wailsjs` / `store` への直接 import 禁止
- `any` の使用禁止
- 公開 Hook、feature 型、ページコンポーネントへの TSDoc 必須化
- TSDoc 構文の妥当性
- 変更対象ファイル内での不要な `export` 検出
- 既存の feature 単位 override で有効化されている `react-hooks/exhaustive-deps`、`max-depth`、`no-else-return` などの厳格ルール

### 6.2 AI が逐次補完する項目

- 1 関数 1 責務（SRP）になっているか
- Hook の戻り値が `state / action / selector` の論理分割になっているか
- Wails adapter の責務が UI 層へ漏れていないか
- 読解負荷が高すぎる巨大関数や巨大 JSX を放置していないか
- lint だけでは判定し切れない公開範囲の妥当性

### 6.3 実行フロー

- 変更中ファイルの確認は `npm run lint:file -- <file...>` を使う
- `lint:file` は JSON 形式で結果を返し、AI がその場で修正に利用する
- フロント変更時の標準フローは `lint:file -> 修正 -> 再実行 -> 最後に lint:frontend`
- `npm run lint:frontend` は作業完了前に必ず実行する

## 7. 品質ゲート

- ローカル実行で以下を必須化する。
- `npm run typecheck`
- `npm run lint`
- `npm run lint:file -- <file...>`（変更中ファイルの逐次確認）
- `npm run test`
- `npm run build`

## 8. リファクタ開始時の優先順位

1. インポート境界の lint 導入（`pages` -> `wailsjs/store` 禁止）
2. `DictionaryBuilder` のページ責務をさらに薄くし、Hook 戻り値を整理
3. `useMasterPersona` の責務分割（設定永続化、タスク監視、データ取得）
4. Wails adapter 関数の共通化
5. テスト基盤導入と主要 Hook の振る舞いテスト作成

## 9. 例外許容箇所

- 現時点の暫定許容箇所は **なし**（2026-03-08 時点）。
