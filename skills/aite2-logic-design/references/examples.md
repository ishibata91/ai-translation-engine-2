# Logic Design Examples

## 責務分解の例
- Controller: request を受け取り workflow へ渡す
- Workflow: ユースケース進行を制御し、必要な slice と runtime を呼ぶ
- Slice: 個別業務ロジックを実行する
- Artifact: 後続工程へ渡す共有成果物を保存する
- Runtime: 外部 I/O を伴う処理を実行する
- Gateway: 外部依頼口の契約を提供する

## Data Flow の例
- 入力: controller から受けた request
- 中間成果物: artifact に保存する handoff データ
- 出力: workflow が返す結果 DTO

## Persistence Boundary の例
- slice ローカルな保存は slice 側に留める
- 後続工程へ渡す共有データは artifact に出す
