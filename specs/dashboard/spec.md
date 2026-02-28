# Dashboard

## Purpose
TBD: アプリケーションの全体像を把握するためのダッシュボード画面に関する振る舞いを定義します。

## Requirements

### Requirement: Active Jobs Dashboard Overview
ダッシュボード画面では、起動時またはデフォルトの画面として、全体の進行中ジョブ・ステータスをテーブル形式で一覧表示しなければならない。

#### Scenario: Initial Dashboard Rendering
- **WHEN** メインコンテンツ領域でダッシュボードを表示するルート (`/`) にアクセスする
- **THEN** 「進行中のジョブ」を含むステータスカード領域が表示される
- **AND** ジョブの一覧として、対象（Mod名）、フェーズ（進捗状態バッジ）、詳細な進捗バー（%表示）が表示される
