## ADDED Requirements

### Requirement: sliceローカル永続化はslice内部に閉じなければならない
システムは、SQLite を含む slice 固有のローカル永続化を各 slice 内に保持しなければならない。slice ローカル永続化は当該 slice の業務ロジックと永続化契約だけが参照でき、他 slice から直接参照されてはならない。

#### Scenario: slice固有DBを当該slice内で保持する
- **WHEN** 開発者が translator や persona のような slice 固有の SQLite 永続化を実装する
- **THEN** 永続化実装とその契約は当該 slice 配下に置かれなければならない
- **AND** 他 slice はその SQLite 実装や保存物を直接 import してはならない

### Requirement: slice間共有データはartifactへ配置しなければならない
システムは、複数 slice から参照される共有データ、中間成果物、resume 用状態を `artifact` に配置しなければならない。ある slice の内部保存物を、後続 slice が直接参照してはならない。

#### Scenario: 後続sliceが前段sliceの成果物を利用する
- **WHEN** ある slice が後続 slice へ渡す共有データを保存する必要がある
- **THEN** 共有データは `artifact` の保存・検索契約へ格納されなければならない
- **AND** 後続 slice は前段 slice の内部 DB や内部 DTO を直接参照してはならない

### Requirement: slice間受け渡しはworkflowがartifact境界で束ねなければならない
システムは、slice 間の受け渡しを `workflow` が `artifact` 識別子、検索条件、batch / page / cursor を用いて束ねなければならない。slice は他 slice や runtime / gateway / workflow を直接 import せず、自 slice の契約と artifact 契約に集中しなければならない。

#### Scenario: slice間連携を実装する
- **WHEN** 開発者が parser の出力を persona や translator へ受け渡す処理を実装する
- **THEN** `workflow` が artifact 上の識別子または検索条件を束ねて後続 slice を呼び出さなければならない
- **AND** slice から `workflow`、`runtime`、`gateway`、他 slice への直接 import は追加してはならない
