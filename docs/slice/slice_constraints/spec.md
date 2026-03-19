# スライス制約

usecase slice の責務境界を固定し、slice が自律的な業務ロジックとローカル永続化に集中できるようにする。

## Requirements

### Requirement: sliceローカル永続化はslice内部に閉じなければならない
システムは、SQLite を含む slice 固有のローカル永続化を各 slice 内に保持しなければならない。slice ローカル永続化は当該 slice の業務ロジックと永続化契約だけが参照でき、他 slice から直接参照されてはならない。

#### Scenario: slice固有DBを当該slice内で保持する
- **WHEN** 開発者が translator や persona のような slice 固有の SQLite 永続化を実装する
- **THEN** 永続化実装とその契約は当該 slice 配下に置かれなければならない
- **AND** 他 slice はその SQLite 実装や保存物を直接 import してはならない

### Requirement: sliceはartifact以外の他区分へ直接依存してはならない
システムは、slice 間受け渡しを artifact 経由に統一するため、slice から `workflow`、`runtime`、`gateway`、他 slice への直接 import を禁止しなければならない。

#### Scenario: slice境界違反の依存を追加する
- **WHEN** 開発者が slice から `workflow`、`runtime`、`gateway`、他 slice への import を追加する
- **THEN** 品質ゲートは当該依存を違反として報告しなければならない
- **AND** slice から `artifact` への依存だけを許可しなければならない
