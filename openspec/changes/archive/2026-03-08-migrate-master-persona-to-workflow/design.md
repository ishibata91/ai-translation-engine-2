## Context

MasterPersona は現状もっとも複雑な長時間ユースケースであり、開始、待機、再開、リトライ、保存、cleanup の全状態を持つ。ここを移行対象に絞れば、新責務区分が実際に成立するかを 1 change で検証できる。

## Goals / Non-Goals

**Goals:**
- MasterPersona の開始 / 再開 / キャンセル / 完了 cleanup を workflow 主導へ移す
- persona DTO マッピングを workflow に集約する
- 既存 Wails API 互換を保つ
- controller 相当コード配置と `go-cleanarch` 区分を実装実態へ揃える

**Non-Goals:**
- 他ユースケースへの横展開
- DB スキーマ刷新
- 全 task 機能の再設計

## Decisions

### 1. 公開 API は維持し、内部接続だけを差し替える

フロント変更を最小にするため、Wails binding のシグネチャは維持する。controller / task API は workflow を呼ぶ adapter として再構成する。

暫定的に `pkg/task` に残る controller adapter は、この change の中で責務を明示する。必要であれば `pkg/controller` へ再配置し、少なくとも lint 上は controller 区分として扱える状態まで持っていく。

### 2. workflow が MasterPersona の全状態遷移を持つ

開始、resume、cancel、progress、phase、cleanup は workflow が決定する。runtime は queue と進捗配信を担うだけに留める。

### 3. persona は request 準備と保存に集中する

`PreparePrompts` と `SaveResults` は維持し、queue 操作や task state 変更は行わない。

### 4. 既存テストは経路移行後の互換確認に使う

現在の bridge / queue 経路テストを workflow 版に寄せて、開始・再開・キャンセル・cleanup の振る舞いが保たれていることを確認する。

### 5. lint 区分は実コード配置に合わせて更新する

`go-cleanarch` の controller / workflow / runtime / gateway 区分は、導入した足場 package だけでなく MasterPersona の実経路を正しく表せる必要がある。`pkg/task` に controller 相当コードが残るなら lint 設定側で扱いを見直し、`pkg/controller` へ整理するなら設定も追従させる。

## Risks / Trade-offs

- [互換 API を残すため adapter が一時的に増える] → 旧 task は薄い adapter として短期的に残す
- [workflow へ責務が寄りすぎる] → MasterPersona 固有ロジックは workflow ファイル単位に閉じ、slice 内ロジックへ侵入しない
- [resume と cleanup の不整合] → 既存 integration test を維持しつつ移行する
- [`go-cleanarch` が実配置を表し切れない] → controller 相当 package の再配置か lint 区分拡張を同 change で行う
