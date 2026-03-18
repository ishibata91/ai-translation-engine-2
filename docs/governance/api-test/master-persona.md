# Master Persona API Test Scope

`api-test` capability における Master Persona 系 controller API テストの対象と観点を定義する。

## 対象 controller

- `pkg/controller/persona_task_controller.go`

## 対象公開メソッド

- `GetAllTasks`
- `StartMasterPersonTask`
- `ResumeTask`
- `ResumeMasterPersonaTask`
- `CancelTask`
- `GetTaskRequestState`
- `GetTaskRequests`

## テスト観点

### 正常系

- `StartMasterPersonTask` が入力 DTO を workflow 入力へ正しく写像して task ID を返す
- `ResumeTask` と `ResumeMasterPersonaTask` が指定 task ID で再開導線を呼び出す
- `GetTaskRequestState` と `GetTaskRequests` が workflow の戻り値をそのまま返す
- `GetAllTasks` が manager store の結果を返す

### 主要異常系

- workflow 未設定時に `StartMasterPersonTask`、`ResumeTask`、`GetTaskRequestState`、`GetTaskRequests` が設定不足 error を返す
- workflow 側の失敗が controller から返る
- `CancelTask` は workflow 未設定時に panic せず no-op で終わる

### 境界責務

- `SetContext` で注入した context が manager / workflow 呼び出しへ伝播する
- `StartMasterPersonTask` が `task.StartMasterPersonTaskInput` を `workflow.StartMasterPersonaInput` に変換する責務だけを持つ
- `CancelTask` は controller 境界で追加判断を持たず、下流の cancel 呼び出しへ委譲する

## 優先ケース

| ケースID | 対象 | 目的 |
| :-- | :-- | :-- |
| MPAPI-01 | `StartMasterPersonTask` | Master Persona 開始の入力写像を固定する |
| MPAPI-02 | `ResumeTask` | 再開導線を固定する |
| MPAPI-03 | `ResumeMasterPersonaTask` | 互換 API が `ResumeTask` と同義であることを固定する |
| MPAPI-04 | `GetTaskRequestState` | 主要な状態取得導線を固定する |
| MPAPI-05 | `GetTaskRequests` | request 一覧取得導線を固定する |
| MPAPI-06 | `CancelTask` | no-op と cancel 委譲の両方を確認する |
