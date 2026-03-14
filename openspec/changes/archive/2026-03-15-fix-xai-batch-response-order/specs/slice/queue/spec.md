## ADDED Requirements

### Requirement: Queue worker は batch 結果を相関 ID で適用しなければならない
Queue worker は batch 実行結果を `results[i] -> jobs[i]` の配列順対応で保存してはならない。結果保存時は `Response.Metadata.queue_job_id` を第一優先の相関キーとして使用し、`job.ID` へ 1:1 で対応付けて状態更新しなければならない。相関不能な結果は誤保存せず failed として扱わなければならない。

#### Scenario: provider 返却順が submit 順と異なる場合でも正しく保存する
- **WHEN** `GetBatchResults` が submit 順と異なる順序で結果を返す
- **THEN** worker は `queue_job_id` で対象 job を特定して保存しなければならない
- **AND** 配列インデックスのみで job を選択してはならない

#### Scenario: Resume 後に pending job 列挙順が変わっても誤保存しない
- **WHEN** 再開実行で pending jobs の取得順が初回 submit 時と異なる
- **THEN** worker は `queue_job_id` で対応付けを行い、初回順序に依存せず保存しなければならない
- **AND** 既存 completed job の結果を別 job へ上書きしてはならない

#### Scenario: 未知または重複した相関 ID は失敗として扱う
- **WHEN** 結果に `queue_job_id` が欠落・未知・重複のいずれかが含まれる
- **THEN** worker は当該 request を failed として記録しなければならない
- **AND** 正常に相関できる request の保存は継続できなければならない

#### Scenario: Gemini metadata 経路で受け取った相関 ID を利用できる
- **WHEN** Gemini Batch 結果に `inlinedResponse.metadata.queue_job_id` が含まれる
- **THEN** worker は `Response.Metadata.queue_job_id` として受け取り、同一方式で保存しなければならない
- **AND** LLM 出力本文から speaker_id 等を抽出して対応付けしてはならない