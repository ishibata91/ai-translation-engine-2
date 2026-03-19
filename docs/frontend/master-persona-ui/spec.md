# Master Persona 画面UI

## Purpose
Master Persona 画面における UI 表示責務と feature hook 境界を定義し、実行方式選択と進捗表示を provider 非依存で扱う。

## Requirements

### Requirement: Master Persona UI Components Integrity
Master Persona UI components MUST preserve existing UX while extracting page logic into a feature hook.
- 既存のUIの見た目 (Tailwind CSS/daisyUI を利用したスタイル) や UX プロセスの流れは維持する。
- `frontend/src/components/ModelSettings.tsx` は `AIプロバイダ`、`モデル`、`実行方式` を主要入力として表示し、feature hook から受け取る capability DTO に基づいて必要な詳細項目だけを条件表示しなければならない。
- Master Persona 画面は provider 個別ルールを直接持たず、feature hook が公開する「表示可能な execution profile」だけを描画しなければならない。

#### Scenario: Existing Feature Parity
- **WHEN** ユーザーがパーソナ生成の開始、一時停止、ログの閲覧など既存の操作を行う
- **THEN** 従来と同様に Wails API 呼び出しが行われ、UI が即座に反応する

#### Scenario: 実行方式の入力が capability DTO に従って制御される
- **WHEN** feature hook が provider / model ごとの execution capability を返す
- **THEN** UI はその capability に含まれる実行方式だけを選択肢として表示しなければならない
- **AND** page コンポーネントは provider 名ごとの分岐を直接持ってはならない

### Requirement: Master Persona のモデル設定は前者の情報優先で表示しなければならない
Master Persona の `ModelSettings` は、まず `AIプロバイダ`、`モデル`、`実行方式`、`Temperature`、`対応状況メモ` を表示し、その後に capability DTO が要求する詳細設定を条件表示しなければならない。`syncConcurrency` のような sync 専用項目は sync 選択時だけ表示し、local provider 固有の詳細項目は capability DTO が要求した場合だけ表示しなければならない。

#### Scenario: sync 実行では並列数を編集できる
- **WHEN** ユーザーが capability DTO 上で sync を選択する
- **THEN** UI は `同期並列数` を表示しなければならない
- **AND** `クラウドBatch` を選んだときは同項目を表示してはならない

#### Scenario: モデルごとの batch 可否を補助文で示す
- **WHEN** UI が modelCatalog から選択中モデルの capability を受け取る
- **THEN** UI は `このモデルは Batch API に対応しています` または `このモデルは同期実行のみ対応です` の補助文を表示しなければならない
- **AND** batch 非対応モデルでは `クラウドBatch` を選択できてはならない

### Requirement: Master Persona の batch 進捗表示は remote provider progress を優先しなければならない
Master Persona の進捗表示は、batch 実行中に remote provider が返す reported progress を主進捗バーへ反映しなければならない。ローカル保存進捗は主進捗へ使わず、補助文言または最終 phase として表現しなければならない。

#### Scenario: batch 実行中はクラウド側進捗を主バーへ表示する
- **WHEN** capability DTO 上で batch 実行が進行中である
- **THEN** UI は reported progress を主進捗バーへ表示しなければならない
- **AND** ステータス文言は `Batch ジョブを送信しました`、`クラウド側で処理中です`、`結果を取り込んでいます` のような provider 非依存文言を使わなければならない

#### Scenario: provider progress を取得できない場合は不定表示へ落とす
- **WHEN** provider が割合進捗を返さない
- **THEN** UI は不定 progress 表示へ切り替えなければならない
- **AND** 状態文言だけは継続更新されなければならない
