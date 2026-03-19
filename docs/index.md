# AI Translation Engine 2 ドキュメント

このディレクトリは AI Translation Engine 2 の spec 正本です。`docs-site/` の Starlight サイトはこの `docs/` を直接読み込み、重複コンテンツを持ちません。

## 読み始める順番

1. [Governance](/governance/)
2. [Frontend](/frontend/)
3. [Workflow](/workflow/)
4. [Slice](/slice/)
5. [Runtime](/runtime/)
6. [Gateway](/gateway/)
7. [Artifact](/artifact/)
8. [Controller](/controller/)
9. [Foundation](/foundation/)
10. [Changes](/changes/)
11. [Skills](/skills/)

## Zone 一覧

- [Governance](/governance/): repo 全体の基準、品質ゲート、責務境界
- [Frontend](/frontend/): UI 構造、画面責務、コーディング標準
- [Controller](/controller/): Wails binding など外部入力境界
- [Workflow](/workflow/): phase 管理、orchestration、resume/cancel
- [Slice](/slice/): ユースケース固有の振る舞いと契約
- [Runtime](/runtime/): queue、task、実行制御基盤
- [Artifact](/artifact/): shared handoff と中間成果物
- [Gateway](/gateway/): config、datastore、llm など外部依頼口
- [Foundation](/foundation/): progress、telemetry など横断基盤
- [Changes](/changes/): 進行中の仕様差分、設計メモ、tasks
- [Skills](/skills/): repo ローカル skill の定義とテンプレート

## 参考資料

- [Gemini Batch Reference](/references/gemini-batch-ref/)
- [xAI Batch Reference](/references/xai-batch-reference/)
- [Gemini OpenAPI JSON](/reference/gemini_complete_openapi.json)
- [xAI OpenAPI JSON](/reference/xai_openapi_full.json)
