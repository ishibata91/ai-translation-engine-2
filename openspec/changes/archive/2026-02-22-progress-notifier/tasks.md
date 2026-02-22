# Tasks: Progress Notifier

## 1. パッケージ作成 (`pkg/infrastructure/progress/`)

- [x] 1.1 `notifier.go` を作成し、`ProgressEvent` 型・`ProgressNotifier` インターフェース・ステータス定数（`StatusInProgress` / `StatusCompleted` / `StatusFailed`）を定義する
- [x] 1.2 `noop.go` を作成し、`NoopNotifier` 構造体と `OnProgress` メソッド（何もしない）を実装する
- [x] 1.3 `provider.go` を作成し、`NewNoopNotifier` 関数と `ProviderSet` を定義する

## 2. Wire プロバイダ登録

- [x] 2.1 ルート `wire.go` に `progress.ProviderSet` を追加し、`ProgressNotifier` インターフェースが DI で解決できるようにする

## 3. job-queue-infrastructure の design.md 更新

- [x] 3.1 `job-queue-infrastructure/design.md` の Decision 3・5 において `ProgressNotifier` の参照先を `pkg/infrastructure/progress` の本 Change に更新する

## 4. 検証

- [x] 4.1 `NoopNotifier.OnProgress` を任意の `ProgressEvent` で呼び出してパニックしないことを確認するスライスレベルテストを作成する
- [x] 4.2 Wire による DI 解決が通ることを `wire_gen.go` 再生成で確認する
