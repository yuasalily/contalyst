# Contalyst — Inception Artifacts

**Contalyst** は Go + Bubble Tea 製の、モダンでカラフルな Docker コンテナ管理 TUI です（Zellij インスパイア）。
このディレクトリは AIDLC（AI-Driven Development Life Cycle）の **Inception フェーズ**に相当するドキュメント群で、「何を・なぜ作るか」を定義します。

## ドキュメント一覧（読む順）

| # | ドキュメント | 内容 |
|---|---|---|
| 00 | [00-inception.md](./00-inception.md) | ビジョン / スコープ / 目的 / ソリューション概要 / リスク / 決定ログ（**ここから読む**） |
| 01 | [01-requirements.md](./01-requirements.md) | 機能要件 (FR) / 非機能要件 (NFR) |
| 02 | [02-personas.md](./02-personas.md) | ペルソナと Story 対応 |
| 03 | [03-stories.md](./03-stories.md) | ユーザーストーリー & 受け入れ基準 |
| 04 | [04-ux-design.md](./04-ux-design.md) | UI/UX・ビジュアルデザイン方針（レイアウト・キー・配色） |
| 05 | [05-units-of-work.md](./05-units-of-work.md) | 作業単位と依存関係・マイルストーン対応 |

## 主要な確定事項（要約）

- **名称**: Contalyst（container + catalyst/analyst、リポジトリ名に一致）
- **技術**: Go 1.26+ / Bubble Tea / Lip Gloss / Bubbles（Charm=MIT）+ Docker 公式 SDK（Apache-2.0）
- **ライセンス方針**: Charm 系のみ MIT を許容、その他は Apache-2.0 限定、コピーレフト不採用（DR-1）
- **ナビゲーション**: k9s 流リソース一覧 + Zellij 流カラフルヒントバー + oxker 流分割詳細（DR-3）
- **判断基準**: 既存類似 TUI と比較して **UX が最良**であること

## 次フェーズ

Construction フェーズ（各 Unit of Work U0→U8 の HOW＝設計・実装）は **完了済み**です。
設計・実装ドキュメントは [`../construction/`](../construction/) を参照してください
（アーキテクチャ、実装ステータス、開発者ガイド、既知の課題）。MVP は全 Unit 実装済みです。
