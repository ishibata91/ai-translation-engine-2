# Logic Design

## Scenario
Terminology (単語翻訳) フローの単語対象レコードを、Dictionary import が辞書構築時に採用する REC 集合と一致させる。

## Goal
Dictionary import を正本とした REC 対象定義を `foundation` の共有定数として持ち、`translationinput` がその定義に従って Terminology 用の正規化済み入力を保存し、`terminology` slice が同じ入力をそのまま request 化できる状態にする。

## Responsibility Split
- Controller:
  変更しない。Terminology phase の起動・状態取得 API を提供するだけに留まる。
- Workflow:
  変更しない。taskID と terminology phase 実行を束ねる。
- Slice:
  `dictionary` は `foundation` の共有 REC 定数を使って import 対象を判定する。
  `terminology` は `translationinput` が返す正規化済み `[]TerminologyEntry` を受け取り、`RecordType + SourceText` 単位で request を構築する。
  NPC の FULL/SHRT 同時翻訳は terminology slice 内の業務ルールとして維持する。
- Artifact:
  `translationinput` は parser 出力から Terminology 対象を抽出し、階層なしの `[]TerminologyEntry` を task 単位で保存・返却する正本になる。
- Runtime:
  変更しない。LLM 実行と進捗通知に集中する。
- Gateway:
  変更しない。新しい config 保存責務は持たない。
- Foundation:
  Dictionary import 正本の REC 集合を共有定数として提供する。

## Data Flow
- 入力:
  parser 出力の各 row と `RecordType`
- 中間成果物:
  `translationinput` artifact に保存される `[]TerminologyEntry`
- 出力:
  terminology slice が生成する LLM request 群

## Main Path
1. `foundation` に Dictionary import 正本の REC 定数を置く。
2. `dictionary` import はその定数で XML の `REC` をフィルタする。
3. `translationinput` は parser 出力から、同じ定数に含まれる row だけを Terminology 対象として抽出する。
4. `translationinput` は Terminology 用入力をカテゴリ階層ではなく、正規化済みの `[]TerminologyEntry` として保存する。
5. `terminology` slice は `[]TerminologyEntry` を読み、`RecordType + SourceText` で重複を束ねて request を作る。
6. 翻訳結果保存時は、同じ重複統合キーに属する全 entry に同じ訳を反映する。
7. NPC は `NPC_:FULL` と `NPC_:SHRT` を `PairKey` で束ね、1 request で同時翻訳し、保存時に 2 レコードへ分配する。

## Normalized Terminology Input
`translationinput` は Terminology 用に以下の正規化済み配列を保持する。

```go
type TerminologyEntry struct {
    ID         string
    EditorID   string
    RecordType string
    SourceText string
    SourceFile string
    PairKey    string
    Variant    string
}
```

補足:
- `PairKey` は NPC FULL/SHRT を同一 request に束ねるためのキーであり、通常レコードでは空でよい。
- `Variant` は `full` / `short` / `single` を表し、NPC 以外は `single` を使う。
- 重複統合キーは `RecordType + SourceText` とし、`EditorID` は翻訳単位ではなく保存先識別に使う。

## Request Construction Rules
- terminology slice は `RecordType + SourceText` ごとに 1 request を作る。
- 同一キーに属する複数 entry は 1 回だけ翻訳する。
- 同一キーに属する全 entry には同じ訳を保存する。
- `NPC_:FULL` と `NPC_:SHRT` は `PairKey` が一致する場合に 1 request に束ねる。
- NPC 以外は PairKey を使わず、正規化済み entry をそのまま重複統合する。

## Persistence Boundary
- 永続化するのは `translationinput` artifact 上の `[]TerminologyEntry` と、従来どおり terminology の phase summary / mod term 保存結果のみ。
- REC 対象定義は固定定数なので DB 保存しない。
- 重複統合は terminology 実行時の処理であり、artifact に翻訳済みキャッシュとして保存しない。

## Side Effects
- 新規の外部 I/O は増やさない。
- LLM request 数は重複統合により減る可能性がある。
- 同一 `RecordType + SourceText` の複数 row は、1 回の翻訳結果を共有する。

## Key Branches
- NPC FULL のみ存在する場合は単独 request として扱う。
- NPC SHRT のみ存在する場合も単独 request として扱う。
- `PairKey` は一致しても `SourceText` が空の row は request 対象から除外する。
- Dictionary import 正本に含まれる REC でも、parser 出力から `SourceText` を構成できない row は Terminology 対象にしない。

## Risks
- 現在の `translationinput` はカテゴリ別 DTO なので、`[]TerminologyEntry` への置換または併設が必要になる。
- `terminology` builder はカテゴリ別入力前提のため、正規化済み配列前提へ組み替えが必要になる。
- `RecordType + SourceText` を重複統合キーとするため、同一表記でも本来訳し分けたいケースは吸収できない。今回はユーザー判断に従い、その単位を正とする。
- Dictionary import 正本の REC 集合に今後変更が入った場合は、`foundation` 定数の変更だけで両 slice に波及することを前提とする。
