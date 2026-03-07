# Tasks: Refactor Frontend VSA Pattern A

## 1. Phase 1: Directory Setup & Type Extraction (Master Persona)
_Goal: フロントエンドの機能を「VSAスライス」として扱うための器を作り、`MasterPersona.tsx` 内にある固有の型を移動する。_
- [x] 1.1 `src/hooks/features/masterPersona/` ディレクトリを作成する。
- [x] 1.2 `src/hooks/features/masterPersona/types.ts` を作成する。
- [x] 1.3 `src/pages/MasterPersona.tsx` から、特定の型定義（`PersonaProgressEvent`, `PersonaNPCRecord`, `PersonaDialogueRecord` など）を `types.ts` に移動し、`export` する。
- [x] 1.4 `src/pages/MasterPersona.tsx` および `src/components/PersonaDetail.tsx` (必要であれば) で、分離した型定義を `src/hooks/features/masterPersona/types.ts` から import するように修正する。

## 2. Phase 2: Logic Extraction to Custom Hook (Master Persona)
_Goal: `MasterPersona.tsx` の肥大化の原因であるWails呼び出し、Zustand連携、ならびに複雑な副作用 (`useEffect`) をカスタムフックに隔離する。この段階ではまだUI側への組み込みは行わず、フック単体を完成させる。_
- [x] 2.1 `src/hooks/features/masterPersona/useMasterPersona.ts` を作成する。
- [x] 2.2 Wails API (`StartMasterPersonTask`, `ResumeTask`, `CancelTask`, `ListNPCs`, `ListDialoguesByPersonaID`, `ConfigGetAll`, `ConfigSet`) の import を Hook 側に移動する。
- [x] 2.3 `MasterPersona.tsx` 内にある、LLM設定（provider, config）の初期化と保存用 `useEffect` 群を `useMasterPersona` 内にまるごとコピーする。
- [x] 2.4 プロンプト設定の初期化と保存用の `useEffect`群、および各種ローカル state (`allNpcData`, `selectedRow`, `isGenerating`, `jsonPath`, `activeTaskId`, `progressCounts` など) を `useMasterPersona` 内にコピー・移植する。
- [x] 2.5 `npcPage` や `pagedNpcData` といったページネーションや検索フィルターの計算（`useMemo` を含む）ロジックを `useMasterPersona` に移植する。
- [x] 2.6 フックの戻り値として、UI 描画に必要な値(`pagedNpcData` 等)とコールバック関数(`handleStart`, `handlePause` 等)を全て `return` するようにインターフェースを定義する。

## 3. Phase 3: Wiring UI (Master Persona)
_Goal: `MasterPersona.tsx` (Pageコンポーネント) から不要なロジックを削ぎ落とし、Hookから受け取ったデータを配線するだけの純粋な Presentational Component に変える。_
- [x] 3.1 `MasterPersona.tsx` の全てのロジック群を削除し、純粋なJSX定義と `useMasterPersona` アダプターの呼び出しのみにする。
- [x] 3.2 `MasterPersona.tsx` から不要になった Wails インポートと Zustand 設定などを整理・削除する。
- [x] 3.3 アプリケーションを実行し、ペルソナ一覧表示、フィルター機能、ローカル設定の保存、タスク実行といった全ての機能がリファクタリング前と同様に動作していることを確認する。

## 4. Phase 4: Directory Setup & Type Extraction (Dictionary Builder)
_Goal: DictionaryBuilder 向けの「VSAスライス」の器を作り、固有の型を分離する。_
- [x] 4.1 新規ディレクトリ `src/hooks/features/dictionaryBuilder/` を作成する。
- [x] 4.2 `src/pages/DictionaryBuilder.tsx` から、Wails イベント（`DictionaryProgressEvent`）や特有のテーブル行のデータ型（`DictRow`, `DictItem` 等）を `src/hooks/features/dictionaryBuilder/types.ts` 等に移管、または切り分ける。
- [x] 4.3 移管した型を `DictionaryBuilder.tsx` 側で正しく import し直す。る。

## 5. Phase 5: Logic Extraction to Custom Hook (Dictionary Builder)
_Goal: WailsによるDBアクセスやイベント購読をカスタムフックに隔離する。_
- [x] 5.1 `src/hooks/features/dictionaryBuilder/useDictionaryBuilder.ts` を作成する。
- [x] 5.2 Wails API (`DictGetSources`, `DictStartImport`, `DictGetEntriesPaginated`, `DictUpdateEntry` 等) と `EventsOn` の import を Hook に移動する。
- [x] 5.3 `DictionaryBuilder.tsx` から、イベント購読 (`Events.EventsOn`) の設定およびクリーンアップ (`useEffect`) を Hook に移植する。
- [x] 5.4 辞書ソース一覧 (`sources`) の取得、エントリページネーション (`entries`, `crossEntries`) の取得・更新・検索ロジックを Hook に移植する。
- [x] 5.5 ファイル選択 (`SelectFiles`) や削除 (`DictDeleteSource`) を含む副作用を Hook に移植する。
- [x] 5.6 フックの戻り値として、UI のビュー状態（`view`, `crossEntries` 等）とハンドラー（`handleImport`, `handleEntriesSave` 等）を `return` する。

## 6. Phase 6: Wiring UI (Dictionary Builder)
_Goal: `DictionaryBuilder.tsx` を純粋な Presentational Component に変える。_
- [x] 6.1 `src/pages/DictionaryBuilder.tsx` にて、`useDictionaryBuilder` を呼び出し戻り値を展開する。
- [x] 6.2 展開した値で動作するように、既存の Wails API 呼び出し、`useState`、`useEffect` のコードを全て削除し、ビュー（JSX）の配線のみを残す。
- [x] 6.3 異なるビュー状態（list, entries, cross-search）が正しく切り替わり、テーブルコンポーネントが描画されることを確認する。
