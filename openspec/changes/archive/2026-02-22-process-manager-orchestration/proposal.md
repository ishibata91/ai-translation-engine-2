# Proposal: Process Manager Orchestration

## Problem & Motivation
Vertical Slice Architecture (VSA) では、各スライス（TermTranslator, PersonaGen, SummaryGeneratorなど）は完全に独立しており、別スライスのDTOや状態を知ることは許されません。
しかし、実際のアプリケーションでは、「Loaderでデータを読み込む」→「用語を抽出する」→「要約を生成する」といった一連のワークフローが存在します。
また、時間を要するバッチAPIのリクエスト群を一時的にJobQueueインフラへ委譲し、完了したJobQueueの結果を元のスライスへ戻す（コールバックする）というハブの役割を持つコンポーネントが欠如しています。

## Proposal
`ProcessManager` スライスを正式な「オーケストレーター」として実装・拡充します。
ProcessManager は各ドメインスライスの純粋な関数（プロンプト生成関数など）を呼び出し、その結果（LLMリクエスト）をプロセスIDと紐付けて JobQueue インフラへ登録します。その後、JobQueueから完了連絡（ポーリング結果等）を受け取ると、結果データを該当するスライスの「保存関数」へ引き渡し、UIへ処理全体の進行状態（Translation State）を連携する役割を担います。

## Capabilities

### New Capabilities
- `ProcessManager`: VSAのスライス間連携、UIへの状態伝達、およびインフラ（JobQueue）へのジョブ委譲と結果回収のオーケストレーション。

## Impact
- `pkg/process_manager` の拡充および各スライスに対する調整層の確立。
- UI側とバックエンド間の進行状態(State)の不整合解消。
