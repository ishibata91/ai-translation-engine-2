## Context

ユースケース単位の spec に、UI 表示要件、queue / progress / retry などの横断要件、構造化ログや品質運用が混在している。これでは change ごとの仕様差分を小さく保てず、1 つの修正が複数 spec の説明重複に波及する。

## Goals / Non-Goals

**Goals:**
- `architecture.md` を構造・責務・依存方向だけを書く文書へ整理する
- 共通要件をユースケース spec から分離するための spec 配置ルールを定義する
- `AGENTS.md` から各 spec の役割を辿れるようにする

**Non-Goals:**
- バックエンドコード実装の移動
- MasterPersona フローの実装変更
- `runtime / gateway` のコード導入

## Decisions

### 1. `architecture.md` は純粋な構造文書に限定する

`architecture.md` には以下だけを残す。
- 責務区分
- 依存方向
- DTO / Contract / DI 原則
- composition root の責務

品質ゲート、テスト設計、ログ運用、フロント構造は専用 spec に委譲する。

### 2. 共通要件はユースケース spec から分離する

UI、workflow、runtime、gateway の共通要件は、特定ユースケース spec に埋め込まず、共通 spec として切り出す。ユースケース spec は、そのユースケース固有の振る舞いに集中させる。

### 3. `AGENTS.md` は文書責務の入口として使う

`AGENTS.md` から、設計判断時にどの spec を参照すべきかを明示する。これにより AI が `architecture.md` へ品質ルールやログ設計を再度書き戻すのを防ぐ。

## Risks / Trade-offs

- [共通 spec の分割先が多すぎる] → 最初は責務の強い文書だけ分離し、細分化しすぎない
- [既存 spec との重複が残る] → 棚卸しタスクで重複箇所を一覧化してから分割する
- [挙動変更と誤解される] → この change は文書整理のみで、仕様の意味変更は行わない
