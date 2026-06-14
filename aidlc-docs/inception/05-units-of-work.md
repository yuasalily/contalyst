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

## v2 Units（M5–M8 / 旧 Out of Scope 昇格分）

> MVP（U0–U8）完了後に着手する作業単位。MVP の基盤（テーブル U1、確認ダイアログ U4、コマンドパレット U7、`dockerx` 隔離 U0）を再利用する。

### U9 — Compose Layer & Project View（docker-compose 一級サポート）
- **内容**: `com.docker.compose.project` ラベルでのプロジェクト/サービス検出・グルーピング、compose 操作（`docker compose` シェルアウト = DR-5、`dockerx` 隣接へ隔離 = NFR-M5）、`:compose` リソース種別の追加、プロジェクト集約一覧 → サービスドリルダウン、サービス横断ログ集約、未導入環境のフォールバック（R9）。
- **対応ストーリー**: US-15, US-16
- **対応要件**: FR-CMP1〜FR-CMP7, NFR-CMP1, NFR-M5
- **依存**: U1（テーブル）, U4（`down` 確認）, U7（コマンドパレット）, U3（ログ集約は詳細/ログ基盤を再利用）
- **規模**: 中〜大
- **マイルストーン**: M5

### U10 — Multi-select & Bulk Actions（バルク/複数選択）
- **内容**: 一覧の選択状態管理（`Space` マーク、全選択 `a`、可視マーカー）、マーク集合への lifecycle 一括適用、件数明示の集約確認ダイアログ、並行実行と個別結果の集約フィードバック（部分失敗許容 = R10）。
- **対応ストーリー**: US-17
- **対応要件**: FR-B1〜FR-B5, NFR-B1
- **依存**: U1（テーブル）, U4（確認ダイアログ）
- **規模**: 小〜中
- **マイルストーン**: M6

### U11 — Hosts / Context Switching（複数ホスト/コンテキスト）
- **内容**: Docker context 列挙（`~/.docker/contexts` + `DOCKER_CONTEXT`/`DOCKER_HOST` = DR-6）、ランタイムでのクライアント再生成（`dockerx`）、切替オーバーレイ（`:context`/`:hosts`）、ヘッダのアクティブホスト表示、切替時の購読 teardown と状態 reset（R11）、接続失敗の非クラッシュ処理。
- **対応ストーリー**: US-18
- **対応要件**: FR-H1〜FR-H5, NFR-H1
- **依存**: U0（接続/`dockerx`）, U2（ヘッダ）, U7（コマンドパレット）
- **規模**: 中（ランタイム再接続・状態 reset が技術的注意点）
- **マイルストーン**: M7

### U12 — Maintenance: Layers / Prune Dashboard / Op Log（メンテナンスビュー）
- **内容**: イメージレイヤービュー（image history）、prune ダッシュボード（種別別の再利用可能量集計 + 選択 prune、R12）、操作ログ（セッション内リングバッファ、`@` オーバーレイ、OQ-6）。操作ログは U4/U9/U10 の全操作経路から記録する横断機能。
- **対応ストーリー**: US-19, US-20, US-21
- **対応要件**: FR-L1, FR-PR1, FR-PR2, FR-OL1, FR-OL2
- **依存**: U8（イメージ/リソース基盤）, U4（prune 確認）
- **規模**: 中（3 ビューの合算。各々は小〜中）
- **マイルストーン**: M8

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

── v2 ──────────────────────────────────────────
 U9  (Compose)                [U1,U3,U4,U7]
 U10 (Bulk/Multi-select)      [U1,U4]
 U11 (Hosts/Context)          [U0,U2,U7]
 U12 (Maintenance views)      [U8,U4]
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
| U9 | U1,U3,U4,U7 | U10, U11, U12 |
| U10 | U1,U4 | U9, U11, U12 |
| U11 | U0,U2,U7 | U9, U10, U12 |
| U12 | U8,U4 | U9, U10, U11 |

> v2 の U9–U12 は MVP 基盤に依存するのみで、相互依存はない（4 単位は並行実行可能）。ただし操作ログ（U12）は U9/U10 の操作経路からも記録されるため、記録フックの I/F は U12 着手時に確定させ U9/U10 がそれに合わせる。

---

## マイルストーンへの対応

| Milestone | 含む Units |
|---|---|
| M1 — Core Read | U0, U1, U2, U3 |
| M2 — Core Control | U4, U5, U6 |
| M3 — Multi-resource | U7, U8 |
| M4 — Polish | 全 Unit にまたがる調整（テーマ追加、コンパクトバー、性能） |
| **M5 — Compose** (v2) | U9 |
| **M6 — Bulk Actions** (v2) | U10 |
| **M7 — Multi-host** (v2) | U11 |
| **M8 — Maintenance** (v2) | U12 |

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
| US-15 | U9 |
| US-16 | U9 |
| US-17 | U10 |
| US-18 | U11 |
| US-19 | U12 |
| US-20 | U12 |
| US-21 | U12 |
