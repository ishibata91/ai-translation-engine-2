## Context

MasterPersona は現状もっとも複雑な長時間ユースケースであり、開始、待機、再開、リトライ、保存、cleanup の全状態を持つ。ここを移行対象に絞れば、新責務区分が実際に成立するかを 1 change で検証できる。

## Goals / Non-Goals

**Goals:**
- MasterPersona の開始 / 再開 / キャンセル / 完了 cleanup を workflow 主導へ移す
- persona DTO マッピングを workflow に集約する
- 既存 Wails API 互換を保つ

**Non-Goals:**
- 他ユースケースへの横展開
- DB スキーマ刷新
- 全 task 機能の再設計

## Decisions

### 1. 公開 API は維持し、内部接続だけを差し替える

フロント変更を最小にするため、Wails binding のシグネチャは維持する。controller / task API は workflow を呼ぶ adapter として再構成する。

### 2. workflow が MasterPersona の全状態遷移を持つ

開始、resume、cancel、progress、phase、cleanup は workflow が決定する。runtime は queue と進捗配信を担うだけに留める。

### 3. persona は request 準備と保存に集中する

`PreparePrompts` と `SaveResults` は維持し、queue 操作や task state 変更は行わない。

### 4. 既存テストは経路移行後の互換確認に使う

現在の bridge / queue 経路テストを workflow 版に寄せて、開始・再開・キャンセル・cleanup の振る舞いが保たれていることを確認する。

## Risks / Trade-offs

- [互換 API を残すため adapter が一時的に増える] → 旧 task は薄い adapter として短期的に残す
- [workflow へ責務が寄りすぎる] → MasterPersona 固有ロジックは workflow ファイル単位に閉じ、slice 内ロジックへ侵入しない
- [resume と cleanup の不整合] → 既存 integration test を維持しつつ移行する
