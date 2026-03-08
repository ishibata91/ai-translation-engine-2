## Context

MasterPersona の実経路を触る前に、どこまでが実行制御基盤で、どこからが外部資源依頼かを固定する必要がある。特に `llm` は queue worker から利用されるため、`runtime -> gateway` の限定依存を明確にしないと package 依存ルールが破綻しやすい。

## Goals / Non-Goals

**Goals:**
- `workflow` の contract と配置を確定する
- `runtime` と `gateway` の責務差をコード配置と spec で固定する
- `runtime -> gateway` の限定依存ルールを定義する
- `go-cleanarch` を品質ゲートへ導入する

**Non-Goals:**
- MasterPersona の実経路移行
- Wails binding の差し替え
- task API の互換実装

## Decisions

### 1. `workflow` は contract 先行で導入する

まず `workflow` の公開契約と state 管理 contract を導入し、controller が依存する先を固定する。実ユースケース移行は後続 change で行う。

### 2. `runtime` は実行制御基盤に限定する

queue、progress、workflow state、event、telemetry は `runtime` として扱う。runtime は phase や保存判定を決めず、workflow から呼ばれる実行時サービスとして振る舞う。

### 3. `gateway` は外部依頼口に限定する

LLM、DB、config、secret、file、外部 API への接続は `gateway` に集約する。slice は自分の業務成立に必要な gateway 契約へ依存してよい。

### 4. `runtime -> gateway` は queue worker の executor 用途に限定する

queue worker が LLM gateway を利用するのは許可するが、runtime が slice 固有ロジックや保存判断を持つのは禁止する。

### 5. `go-cleanarch` で依存方向を固定する

`controller -> workflow`
`workflow -> runtime`
`workflow -> usecase slice`
`usecase slice -> gateway`
`runtime -> gateway` の限定依存

以外を lint で検出する。

## Risks / Trade-offs

- [runtime と gateway の境界が抽象的すぎる] → queue worker / llm の具体例を設計と spec に明記する
- [先に contract だけ導入すると未使用コードが増える] → change 範囲を MasterPersona 移行前提の最小スキャフォールドに留める
- [go-cleanarch のルールが過剰] → 初回は関係 package に対象を絞る
