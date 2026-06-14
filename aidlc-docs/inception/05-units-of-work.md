# Contalyst — Units of Work（作業単位）

> AIDLC の "Unit of Work" = 設計・実装・テストを一括でこなせる、関連ストーリーの論理的なまとまり。
> マイクロサービスでもファイルでもスプリントでもなく、**並行実行可能な作業パッケージ**。各 Unit は "bolt"（短い集中サイクル）で実行する想定。
> 対応ストーリーは [03-stories.md](./03-stories.md)、要件は [01-requirements.md](./01-requirements.md)。

---

## Units 定義

### U0 — Foundation & Docker Layer（基盤・Docker 接続）
- **内容**: Go module 初期化、依存導入、`dockerx` パッケージ（公式 SDK ラッパ、API ネゴシエーション、接続エラー処理）、Bubble Tea アプリ骨格（Model/Update/View）、終了処理。
- **対応ストーリー**: US-9, US-14
- **対応要件**: FR-E1, FR-E2, FR-E3, NFR-M1, NFR-M2, NFR-R1
- **依存**: なし（最初に着手）
- **規模**: 中

### U1 — Container List & Live Updates（コンテナ一覧）
- **内容**: live テーブル（Bubbles table）、状態の色分け、自動更新、行移動。3 段レイアウトの骨格。
- **対応ストーリー**: US-1
- **対応要件**: FR-C1, FR-C2, NFR-U2, NFR-P1, NFR-P3
- **依存**: U0
- **規模**: 中

### U2 — Hint Bar, Help & Theming（ヒントバー・ヘルプ・テーマ）
- **内容**: 常時カラフルなヒントバー（コンパクト切替）、`?` コンテキストヘルプ、既定テーマ「Catalyst」、色 degrade、ASCII フォールバック。
- **対応ストーリー**: US-10, US-11
- **対応要件**: FR-NAV5, FR-NAV6, FR-NAV7, FR-NAV9, NFR-U1, NFR-U3, NFR-U4, NFR-U5
- **依存**: U1（一覧上に重ねるため）
- **規模**: 中

### U3 — Log Streaming & Detail View（ログ追尾・詳細）
- **内容**: ログの follow（goroutine→tea.Msg）、TTY 有無での demux 分岐（stdcopy）、viewport スクロール、検索、タイムスタンプ、oxker 流分割詳細レイアウト。
- **対応ストーリー**: US-2, US-12
- **対応要件**: FR-C3, FR-C4, FR-C5, FR-C10, NFR-P2, NFR-R2
- **依存**: U1, U2
- **規模**: 中〜大

### U4 — Container Controls & Safety Dialogs（操作・確認ダイアログ）
- **内容**: start/stop/restart/pause/kill、削除、確認ダイアログ（既定フォーカス安全側）、操作結果トースト、エラー処理。
- **対応ストーリー**: US-3, US-6
- **対応要件**: FR-C6, FR-C7, FR-C8, FR-NAV8, NFR-R3, NFR-R4
- **依存**: U1, U2
- **規模**: 中

### U5 — Exec Into Shell（exec）
- **内容**: コンテナ内インタラクティブシェル起動、`/bin/sh` フォールバック、Bubble Tea の suspend/再開ハンドリング、復帰時の一覧再描画。
- **対応ストーリー**: US-5
- **対応要件**: FR-C9
- **依存**: U1, U0
- **規模**: 中（端末ハンドオフが技術的注意点）

### U6 — Stats Streaming（統計）
- **内容**: CPU/メモリ/ネット/IO のストリーミング、CPU% の delta 算出、詳細ビューの右ペイン描画。
- **対応ストーリー**: US-4
- **対応要件**: FR-C11
- **依存**: U3（詳細ビュー基盤）
- **規模**: 中

### U7 — Filter & Command Palette（フィルタ・コマンドパレット）
- **内容**: `/` ファジーフィルタ、`:` コマンドパレット（autosuggest, 一覧）、ドリルダウン/戻る + パンくず、履歴移動。
- **対応ストーリー**: US-7, US-8
- **対応要件**: FR-NAV1, FR-NAV2, FR-NAV3, FR-NAV4, FR-NAV10
- **依存**: U1, U2
- **規模**: 中

### U8 — Images / Volumes / Networks（その他リソース）
- **内容**: 各リソースの一覧と削除/prune（確認経由）。`:` から切替。U1 のテーブル基盤を再利用。
- **対応ストーリー**: US-13
- **対応要件**: FR-I1, FR-I2, FR-V1, FR-N1
- **依存**: U1, U4（確認ダイアログ）, U7（コマンドパレット）
- **規模**: 中

---

## 依存関係マトリクス

```
U0 (Foundation)
 └─> U1 (Container List)
      ├─> U2 (Hint/Help/Theme)
      │    ├─> U3 (Logs/Detail) ──> U6 (Stats)
      │    ├─> U4 (Controls/Dialogs)
      │    └─> U7 (Filter/Palette)
      └─> U5 (Exec)            [U0,U1]
 U8 (Images/Vol/Net)          [U1,U4,U7]
```

| Unit | 直接依存 | 並行可能な相手 |
|---|---|---|
| U0 | — | — |
| U1 | U0 | — |
| U2 | U1 | U5 |
| U3 | U1,U2 | U4, U5, U7 |
| U4 | U1,U2 | U3, U5, U7 |
| U5 | U0,U1 | U2,U3,U4,U7 |
| U6 | U3 | U4,U7 |
| U7 | U1,U2 | U3,U4,U5 |
| U8 | U1,U4,U7 | U6 |

---

## マイルストーンへの対応

| Milestone | 含む Units |
|---|---|
| M1 — Core Read | U0, U1, U2, U3 |
| M2 — Core Control | U4, U5, U6 |
| M3 — Multi-resource | U7, U8 |
| M4 — Polish | 全 Unit にまたがる調整（テーマ追加、コンパクトバー、性能） |

---

## Story → Unit マップ

| Story | Unit |
|---|---|
| US-1 | U1 |
| US-2 | U3 |
| US-3 | U4 |
| US-4 | U6 |
| US-5 | U5 |
| US-6 | U4 |
| US-7 | U7 |
| US-8 | U7 |
| US-9 | U0 |
| US-10 | U2 |
| US-11 | U2 |
| US-12 | U3 |
| US-13 | U8 |
| US-14 | U0 |
