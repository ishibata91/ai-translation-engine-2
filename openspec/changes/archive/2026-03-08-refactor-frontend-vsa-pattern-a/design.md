# Design: Refactor Frontend VSA Pattern A

## Context
現在、フロントエンドの主要ページである `MasterPersona.tsx` や `DictionaryBuilder.tsx` は、WailsのAPI呼び出し、Zustandからの状態取得、およびUIのDOM構築（JSX）を一つの巨大なファイル内で全て処理しています（Fat Page現象）。
これにより、見通しが悪く、UIデザインの微修正を行う場合でも複雑なロジックを読み解く必要があり、AIによる自動生成や改修を行う際の精度低下・デグレの温床となっています。
一方、バックエンドでは既にVertical Slice Architecture（VSA）とInterface-First AIDDが採用されており、これと軌を一にするフロントエンドアーキテクチャの導入が求められています。

## Goals / Non-Goals
**Goals:**
- UI（Presentational）とロジック（Container/Custom Hooks）の責任境界を明確化する。
- ページコンポーネントのコード量を劇的に削減し、UIの配置に特化させる。
- 各ページの固有ロジックは対応する専用のCustom Hookに隠蔽する（Headless Architecture）。
- 複数ページで共通利用される汎用的なUIパーツ（必要であれば）を `src/components/ui` 等に集約・共通化する。

**Non-Goals:**
- **UIデザイン・見た目の変更:** 本件は純粋な内部リファクタリングであり、ユーザーの目に触れる変更は行わない。
- **全機能の完全なFSD化:** 過剰な設計（機能ごとのフォルダ完全分離など）は行わず、手痛い部分である「巨大コンポーネントのロジック分離」にとどめる。
- **Zustandストア自体の構造変更:** あくまでHookからZustandを呼び出すように変更するのみで、ストアそのものの設計変更は含まない。

## Decisions
- **Headless Component Pattern（Pattern A）の採用:** 
  UIコンポーネントに持たせる状態やロジックを極力減らし、すべて `useMasterPersona()` のような形にまとめ、コンポーネントからは戻り値に含まれるデータや関数を利用して描画のみを行う形式にします。

### リファクタリングの具体的手順 (What to move and how)

**1. Master Persona 画面のリファクタリング**
- **新規作成:** `src/hooks/features/masterPersona/useMasterPersona.ts`
  - *移動する処理:*
    - Wails呼び出し (`StartMasterPersonTask`, `ResumeTask`, `CancelTask`, `ListNPCs`, `ListDialoguesByPersonaID`, `ConfigGetAll`, `ConfigSet`)
    - LLM設定、プロンプト設定の永続化管理 (`useEffect` の群れ)
    - ページネーション状態処理 (`allNpcData`, `npcPage`等の計算)
    - ローカル側のステータス管理 (`isGenerating`, `progressPercent`, `statusMessage`) 等
  - *インターフェース (戻り値):*
    - UI側に必要な変数（`npcData`, `pagedNpcData`, `isGenerating`, `llmConfig` 等）
    - UI側から呼び出す関数（`handleStart`, `handlePause`, `handleRowSelect`, `applyFilters` 等）
- **修正:** `src/pages/MasterPersona.tsx` (51KB -> 大幅削減)
  - `const { ... } = useMasterPersona()` を呼び出し、内部の長大な `useState` と `useEffect` を全て削除。
  - JSX (DOM構築部分) はそのまま維持し、見た目の崩れが出ないようにする。

**2. Dictionary Builder 画面のリファクタリング**
- **新規作成:** `src/hooks/features/dictionaryBuilder/useDictionaryBuilder.ts`
  - *移動する処理:*
    - Wails呼び出し (`DictGetSources`, `DictStartImport`, `DictGetEntriesPaginated`, `DictUpdateEntry`, `DictDeleteSource`, etc.)
    - イベント購読 (`Events.EventsOn('dictionary:import_progress')`) と進捗状態の管理 (`importMessages`, `isImporting`)
    - 横断検索ロジックとページネーション (`entries`, `crossEntries`, `fetchCrossSearch`)
  - *インターフェース (戻り値):*
    - 辞書ソース一覧、個別エントリ、インポート状況などの状態
    - 操作関数（`handleImport`, `handleEntriesSave`, `handleDeleteSource`, etc.）
- **修正:** `src/pages/DictionaryBuilder.tsx` (33KB -> 大幅削減)
  - 同様にフックから値と関数を取得し、View の切り替え (`list`, `entries`, `cross-search`) と UI コンポーネントの配線に徹する。

**3. 型定義の機能単位への集約**
- 肥大化した `types/` ディレクトリ（またはコンポーネント内の一部の型）のうち、その機能しか使わない型（例：`PersonaNPCRecord` や `DictSourceRow` など）は、必要に応じて `src/features/.../types.ts` 等へ移動、もしくは分離が容易な場所で管理し、モジュール間のグローバルな密結合を防ぎます。

## Risks / Trade-offs
- *Risk:* 既存コードの分割時に、不要な再レンダリングが発生する可能性がある。
  - *Mitigation:* 依存配列の管理や、必要に応じてコールバック (`useCallback`) および算出処理 (`useMemo`) を Custom Hook 内で適切に設定する。
- *Trade-off:* ファイル数が増加する。
  - *Mitigation:* `hooks/features/` ディレクトリの中に機能ごとにフォルダを設け、名前空間を整理することで直感的に理解しやすい構成を維持する。
