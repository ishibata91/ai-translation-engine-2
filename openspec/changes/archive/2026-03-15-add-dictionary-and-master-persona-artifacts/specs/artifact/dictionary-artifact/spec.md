## ADDED Requirements

### Requirement: 共有辞書成果物は artifact の正本として保存されなければならない
システムは、Dictionary Builder が管理する辞書ソースと辞書エントリを `pkg/artifact/dictionary_artifact` の契約を通じて `artifact` に保存しなければならない。translation flow など後続機能が再利用する辞書データを、slice ローカル DB の複製や別経路の正本として保持してはならない。

#### Scenario: 辞書ソースを artifact に作成する
- **WHEN** ユーザーが Dictionary Builder で新しい辞書ソースを作成する
- **THEN** システムは `artifact` に辞書ソース行を保存しなければならない
- **AND** source 名、format、作成日時を後続処理が参照できる形で保持しなければならない

#### Scenario: 辞書エントリを source 単位で保存する
- **WHEN** ユーザーが特定 source に属する辞書エントリを追加または更新する
- **THEN** システムは `artifact` 上の当該 source に対して entry を保存しなければならない
- **AND** `record_type`、`edid`、`source_text`、`dest_text` を一貫した正本として更新しなければならない

#### Scenario: 辞書 source を削除すると配下 entry も artifact から削除される
- **WHEN** ユーザーが辞書 source を削除する
- **THEN** システムは当該 source と配下 entry を `artifact` から削除しなければならない
- **AND** 他 source の辞書エントリを変更してはならない

### Requirement: 辞書 artifact は source 単位と横断の検索を提供しなければならない
システムは、`dictionary_artifact` を通じて source 単位の一覧取得と、全 source 横断の検索を提供しなければならない。Dictionary Builder UI と translation flow 向けの再利用は、同じ artifact 正本を検索して成立しなければならない。

#### Scenario: source 単位のページング一覧を返せる
- **WHEN** Dictionary Builder が特定 source の entries をページ単位で取得する
- **THEN** システムは `artifact` から当該 source の entries をページング付きで返さなければならない
- **AND** UI は別 DB へのフォールバックなしで一覧表示できなければならない

#### Scenario: 全 source 横断検索を返せる
- **WHEN** translation flow または Dictionary Builder が全 source 横断検索を行う
- **THEN** システムは `artifact` の正本から一致 entry を返さなければならない
- **AND** source 情報を含めて返し、ヒット元 source を識別できなければならない

### Requirement: 辞書 artifact は slice 非依存の DTO 契約を持たなければならない
`pkg/artifact/dictionary_artifact` は、自前の DTO と repository 契約を公開しなければならない。artifact package は `pkg/slice/dictionary` の DTO や内部型に依存してはならない。

#### Scenario: dictionary slice が artifact 契約を利用する
- **WHEN** dictionary slice が shared dictionary の保存または検索を実行する
- **THEN** slice は `dictionary_artifact` の contract と DTO を使って処理しなければならない
- **AND** artifact package から slice package への import を追加してはならない
