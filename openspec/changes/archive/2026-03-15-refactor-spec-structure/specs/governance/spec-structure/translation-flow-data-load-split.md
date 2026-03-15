## REMOVED Requirements

### Requirement: 翻訳フローはデータロードフェーズから開始しなければならない
**Reason**: フェーズ順序と進行制御は `workflow/translation-flow-data-load` で扱い、画面表示は `frontend/translation-flow-data-load-ui` で扱うため、混在 capability を廃止する。
**Migration**: フェーズ順序と後続フェーズへの進行は `workflow/translation-flow-data-load/spec.md` を参照し、画面表示は `frontend/translation-flow-data-load-ui/spec.md` を参照する。

### Requirement: データロードフェーズは複数ファイル選択を受け付けなければならない
**Reason**: 複数ファイル選択は画面入力責務であり、`frontend/translation-flow-data-load-ui` へ移すため。
**Migration**: データロード画面の入力要件は `frontend/translation-flow-data-load-ui/spec.md` を参照する。

### Requirement: データロードフェーズは artifact に保存された全 section の翻訳対象をファイル単位で表示しなければならない
**Reason**: artifact preview の画面表示責務は frontend 区分で定義し、artifact handoff の進行責務は workflow 区分で定義するため、1 capability への混在をやめる。
**Migration**: preview table の UI 要件は `frontend/translation-flow-data-load-ui/spec.md`、artifact handoff の進行要件は `workflow/translation-flow-data-load/spec.md` を参照する。

### Requirement: ファイルごとの翻訳対象テーブルは 50 行単位でページングできなければならない
**Reason**: ページングは画面コンポーネント責務であり、frontend 区分へ移すため。
**Migration**: ページング要件は `frontend/translation-flow-data-load-ui/spec.md` を参照する。

### Requirement: ファイルごとの翻訳対象テーブルは折りたたみ可能でなければならない
**Reason**: 折りたたみ UI も画面責務であり、frontend 区分へ移すため。
**Migration**: 折りたたみ要件は `frontend/translation-flow-data-load-ui/spec.md` を参照する。
