# Contalyst — Inception Document

> AIDLC (AI-Driven Development Life Cycle) の **Inception フェーズ** に相当するドキュメントです。
> 「何を、なぜ作るのか (WHAT / WHY)」を定義します。「どう作るか (HOW)」の詳細は Construction フェーズで扱います。

- **Project**: Contalyst — A modern, colorful Docker TUI
- **Phase**: Inception（**v2 スコープ拡張で再オープン**）
- **Status**: MVP (M1–M4 / U0–U8) は Construction 完了・実装済み。本ドキュメントは **v2 拡張** として、旧 Out of Scope の 4 領域（compose / バルク操作 / 複数ホスト / メンテナンスビュー）をスコープへ昇格（承認ゲートはスキップ。判断は本ドキュメント内に記録）。
- **Last updated**: 2026-06-14
- **Owner**: yuasalily1011@gmail.com

関連アーティファクト:
- [01-requirements.md](./01-requirements.md) — 機能要件 / 非機能要件
- [02-personas.md](./02-personas.md) — ペルソナ
- [03-stories.md](./03-stories.md) — ユーザーストーリー & 受け入れ基準
- [04-ux-design.md](./04-ux-design.md) — UI/UX・ビジュアルデザイン方針
- [05-units-of-work.md](./05-units-of-work.md) — 作業単位 (Units of Work) と依存関係

---

## 1. Vision & Intent（ビジョンと意図）

開発者が日常的に行う Docker コンテナの確認・操作（状態確認、ログ追尾、起動/停止、シェルへの exec、リソース使用状況の把握）は、`docker ps`・`docker logs -f`・`docker exec` といったコマンドを何度も打ち、ターミナルとコンテナ ID を行き来する作業の繰り返しになりがちである。Docker Desktop の GUI は重く、SSH 越しのリモートホストでは使えない。

**Contalyst** は、ターミナル上で完結する Docker コンテナ管理 TUI である。**Zellij のようにモダンでカラフル、かつ「見ればわかる (discoverable)」UI** を提供し、キーボード中心の高速な操作で、コンテナ・イメージ・ボリューム・ネットワークの一覧と操作、リアルタイムなログ/メトリクスの閲覧を一つの画面体験に統合する。SSH 越しでも軽快に動作し、ローカルでも本番調査でも「最初に開くツール」になることを目指す。

**One-line vision**:
> 開発者が Docker を触るときに最初に開く、速くて美しく、迷わず使えるターミナル UI。

---

## 2. Elevator Pitch（プロダクトコンセプト）

> **Docker を日常的に使う開発者・運用者** にとって、コンテナの状態確認・ログ調査・トラブルシューティングを CLI コマンドの反復で行うのは煩雑である。
> **Contalyst** は **Go + Bubble Tea 製の Docker コンテナ管理 TUI** であり、**Zellij のように色彩豊かで操作が一目でわかる UI から、一覧・ログ追尾・exec・リソース監視をキーボードだけで完結できる**。
> **lazydocker や ctop** とは異なり、**k9s 流のリソース一覧＋コマンドパレットによる拡張性の高いナビゲーションと、Zellij 流の常時表示されるカラフルなキーヒントバーによる学習不要の操作性** を両立する。

**3 つの主要セールスポイント**
1. **Discoverable**: 画面下部の常時表示ヒントバーと `?` ヘルプで、ドキュメントを読まずに操作できる。
2. **Fast & keyboard-first**: vim ライクなキー操作、`/` フィルタ、`:` コマンドパレットで瞬時に目的の操作へ到達。
3. **Beautiful & themeable**: Zellij インスパイアの配色・角丸フレーム・状態に応じた色分けで、SSH 越しでも気持ちよく使える。

---

## 3. Objectives & Success Metrics（目的と成功指標）

| # | 目的 (Objective) | 成功指標 (Success Metric) |
|---|---|---|
| O1 | コマンド反復を置き換える | 主要操作（一覧→ログ→exec）が、起動後マウスなし・3 キー以内で到達できる |
| O2 | 学習コストの最小化 | ドキュメント未読のユーザーが、ヒントバーのみで起動/停止/ログ/exec を実行できる |
| O3 | モダンでカラフルな見た目 | 状態（running/exited/paused 等）が色で即判別でき、テーマ切替が可能 |
| O4 | 軽快な応答性 | 一覧の描画・更新が 60ms 以内、ログ追尾が遅延なく流れる（体感ベース） |
| O5 | リモート/コンテナ内でも動く | SSH 越し・コンテナ内実行でもログ/統計が正しく表示される（既存ツールの典型バグを回避） |
| O6 | 破壊的操作の安全性 | 削除等は必ず確認ダイアログを経由し、デフォルトフォーカスは非破壊側 |

---

## 4. Stakeholders & Personas（ステークホルダーとペルソナ）

- **Sponsor / Decision-maker**: プロジェクトオーナー（本リポジトリ所有者）。最終的な UX 判断基準は「類似 TUI と比較して UX が最良であること」。
- **Primary users**: アプリ開発者、DevOps/SRE、ホームラボ運用者。
- 詳細は [02-personas.md](./02-personas.md) を参照。

---

## 5. Scope（スコープ）

### 5.1 In Scope（MVP — ✅ 実装済み）

- **コンテナ管理**: 一覧（live 更新）、start / stop / restart / pause / unpause / remove / kill。
- **ログ**: リアルタイム追尾（follow）、スクロール、検索、タイムスタンプ表示切替、truncate しない取得。
- **exec**: コンテナ内シェルへの起動（`/bin/sh` フォールバック）。
- **inspect**: 整形された詳細情報の表示。
- **stats**: CPU / メモリ / ネットワーク / IO のリアルタイム表示。
- **イメージ / ボリューム / ネットワーク**: 一覧と基本操作（削除・prune）。
- **ナビゲーション**: `:` コマンドパレット（リソース種別切替）、`/` ファジーフィルタ、ドリルダウン/戻る + パンくず。
- **操作性 UI**: 常時表示キーヒントバー、`?` コンテキストヘルプ、破壊的操作の確認ダイアログ。
- **見た目**: 1 つ以上の良質なカラーテーマ、状態の色分け、角丸フレーム（opt-in）。

### 5.2 In Scope（v2 — 本更新で Out of Scope から昇格）

> MVP（U0–U8）完了後、旧 §5.2 Out of Scope のうち以下 4 領域を新たにスコープへ取り込む。AIDLC に従い inception を再オープンして定義（2026-06-14）。要件は [01-requirements.md](./01-requirements.md) の「v2 機能要件」、ストーリーは US-15〜US-21、作業単位は U9〜U12 に対応。

- **docker-compose 一級サポート** (M5・最優先): `com.docker.compose.project` ラベルでプロジェクト/サービスをグルーピングし、`:compose` で一覧。プロジェクト単位の up / down / restart / rebuild（`--no-cache` 含む）、`depends_on` を尊重した依存順、サービス横断のログ集約。→ **競合の最大の隙を埋める最優先領域**（旧 §5.2 で「Post-MVP の最優先候補」と記載）。
- **バルク / 複数選択操作** (M6): 一覧で複数行をマーク（`Space`）し、start / stop / restart / remove を一括実行。確認は対象件数を明示した 1 ダイアログに集約、実行は並行・個別結果をフィードバック（部分失敗を許容）。
- **複数ホスト / コンテキスト切替** (M7): Docker context（`~/.docker/contexts` / `DOCKER_CONTEXT` / `DOCKER_HOST`）を列挙し、ランタイムで接続先を切替。ヘッダにアクティブホストを表示。切替時はストリームを teardown して状態を reset。
- **メンテナンスビュー** (M8): イメージレイヤービュー（history: レイヤー / サイズ / 生成コマンド）、prune ダッシュボード（images / containers / volumes / networks / build cache の再利用可能量を集計し、種別を選択して prune）、操作ログ（lazygit の `@` 相当: 実行した操作を時刻・結果付きで記録・閲覧）。

### 5.3 Out of Scope（v2 でも作らない / さらに後続）

- Docker Swarm（nodes / services / stacks）。
- マウスのフルサポート（行クリック / ホイール; 端末のテキスト選択を奪うため opt-in が前提）。
- ログのファイル保存。
- プラグイン機構、設定ファイルによる詳細カスタマイズ（キーバインド再割当を含む）。

### 5.4 Release Roadmap（増分計画）

**MVP（完了済み）**

- **M1 — Core Read**: コンテナ一覧 + ログ追尾 + inspect + ヒントバー + 1 テーマ。
- **M2 — Core Control**: start/stop/restart/remove + 確認ダイアログ + stats + exec。
- **M3 — Multi-resource**: images / volumes / networks + `:` コマンドパレット + `/` フィルタ。
- **M4 — Polish**: テーマ複数化、ヘルプ充実、コンパクトヒントバー、パフォーマンス調整。

**v2（本更新で計画化）**

- **M5 — Compose**: compose プロジェクト/サービスのグルーピング表示 + up/down/restart/rebuild(`--no-cache`) + 依存順 + サービス横断ログ。(U9)
- **M6 — Bulk Actions**: 複数選択（`Space`）+ 一括 start/stop/restart/remove + 件数明示の集約確認。(U10)
- **M7 — Multi-host**: Docker context 列挙 + ランタイム切替 + ヘッダのアクティブホスト表示。(U11)
- **M8 — Maintenance**: イメージレイヤービュー + prune ダッシュボード + 操作ログ。(U12)

**Post-v2**: Docker Swarm、フルマウス対応、ログのファイル保存、プラグイン/設定ファイル。

---

## 6. Solution Overview（ソリューション概要）

### 6.1 アーキテクチャ方針

Bubble Tea の **The Elm Architecture (Model–Update–View)** に従う。

- **Docker layer**: 公式 Go SDK `github.com/docker/docker/client` を薄くラップした `dockerx` パッケージ。API バージョンネゴシエーション必須（`WithAPIVersionNegotiation`）。ストリーム（logs/stats）は `tea.Cmd` 経由で `tea.Msg` としてモデルへ流す。
- **State (Model)**: アクティブなリソース種別、選択行、フィルタ文字列、ドリルダウンスタック（パンくず）、ストリーム購読状態を保持。
- **View**: Lip Gloss でスタイリング。画面は「ヘッダ（パンくず/コンテキスト）＋メイン（一覧 or 分割詳細）＋常時ヒントバー」の 3 段構成。
- **Components**: Bubbles の `table` / `viewport`（ログ）/ `textinput`（フィルタ・コマンドパレット）/ `help` / `spinner` を活用。
- **Concurrency**: ログ/統計のストリーミングは goroutine + channel → `tea.Cmd`。コンテキストキャンセルで購読解除。

詳細な画面設計・キーバインド・配色は [04-ux-design.md](./04-ux-design.md)。

### 6.2 技術スタックとライセンス方針

ユーザー指定: **Go + Bubble Tea を使用。その他ライブラリは Apache-2.0 を許容。**

調査結果に基づく判断:

| 区分 | ライブラリ | ライセンス | 採否・備考 |
|---|---|---|---|
| 言語 | Go (1.26+) | BSD-3 | 指定 |
| TUI フレームワーク | `charmbracelet/bubbletea` | **MIT** | **指定。** 以下 Charm 系も同一エコシステムとして採用 |
| スタイリング | `charmbracelet/lipgloss` | **MIT** | カラフルな見た目に必須 |
| UI 部品 | `charmbracelet/bubbles` | **MIT** | table/viewport/textinput/help 等 |
| フォーム/ダイアログ | `charmbracelet/huh`（任意） | **MIT** | 確認ダイアログに有用 |
| Docker SDK | `github.com/docker/docker/client` | **Apache-2.0** | 採用 |
| ログ多重分離 | `github.com/docker/docker/pkg/stdcopy` | **Apache-2.0** | TTY なしストリームの demux に必須 |
| compose（v2 候補） | `docker compose` CLI シェルアウト | — | **DR-5: 既定戦略**。バイナリ外依存（R9） |
| compose 純 Go（v2 代替案） | `github.com/compose-spec/compose-go` | **Apache-2.0** | OQ-4 で評価。単一バイナリ方針に合致するが実装量大 |

> **License Decision (DR-1)**: ユーザーは Bubble Tea を明示的に指定したが、Bubble Tea とその周辺（Lip Gloss / Bubbles 等）は **MIT** であり、指定された「Apache-2.0」とは異なる。MIT は Apache-2.0 と同等以上に寛容（permissive）で、同一バイナリへの結合・再配布に法的支障はない。「カラフルでモダンな見た目」という要件達成には Lip Gloss が事実上不可欠。
> → **結論**: Charm エコシステム（MIT）は Bubble Tea 指定に内包されるものとして採用。**Charm 以外の依存はすべて Apache-2.0 に限定**（Docker SDK が該当）し、GPL/LGPL/MPL 等のコピーレフトは一切採用しない。本判断は承認不要の方針に従い確定とする。

> **Tech Note (DR-2)**: Docker Engine v29 以降、`github.com/docker/docker` は `github.com/moby/moby/client` へ移行が進む（ライセンスは Apache-2.0 のまま）。インポートパスは将来移行できるよう `dockerx` で隔離する。

### 6.3 採用するナビゲーションモデル（DR-3）

調査した既存ツール（lazydocker / ctop / dry / oxker / k9s / lazygit / Zellij）の比較から:

- **背骨は k9s 流のリソース一覧モデル**: 1 画面 1 リソース種別の live テーブル。Docker の多様なリソース種別に対し、lazydocker の固定 6 パネルよりスケールする。
- **`:` コマンドパレット** で Docker の語彙（`:ps`/`:containers`, `:images`, `:volumes`, `:networks`）を使ってリソース切替。
- **`Enter` でドリルダウン**（コンテナ → ログ/inspect/stats の分割詳細、oxker の「全部見える」思想）、**`Esc` で戻る**、下部にパンくず。
- **Zellij 流の常時カラフルなキーヒントバー**（lazygit で最も評価された discoverability）。コンパクト 1 行モードも提供。
- **モードは最小限**: Docker の操作集合は Zellij より小さいため、重いモーダルより直接単キー操作を優先。`:`（コマンド）と `/`（フィルタ）のみ軽量モードとして持つ。

---

## 7. Assumptions, Dependencies & Constraints（前提・依存・制約）

**前提 (Assumptions)**
- ユーザーのマシンに Docker Engine が動作し、`DOCKER_HOST` もしくはデフォルトソケット経由で接続できる。
- 256 色以上（できれば truecolor）対応のターミナルで実行される。
- 主対象は Linux/macOS のターミナル。Windows は WSL 経由を想定。

**依存 (Dependencies)**
- Docker Engine API（バージョンネゴシエーションで複数バージョンに追従）。
- Charm エコシステム（MIT）と Docker SDK（Apache-2.0）。

**制約 (Constraints)**
- ライセンス: Charm=MIT、その他=Apache-2.0 のみ（DR-1）。コピーレフト禁止。
- フレームワーク: Bubble Tea 必須。
- 単一バイナリ配布（外部ランタイム不要）。

---

## 8. Risks & Mitigations（リスクと対策）

| # | リスク | 影響 | 対策 |
|---|---|---|---|
| R1 | ログ/統計ストリームの demux 誤り（TTY 有無で形式が異なる） | 既存ツールの典型バグ。文字化け | コンテナの TTY フラグで分岐。非 TTY は `stdcopy.StdCopy`、TTY は raw |
| R2 | API バージョン不一致で起動時クラッシュ（dry/ctop の前例） | 起動不能 | `WithAPIVersionNegotiation()` を必須化 |
| R3 | コンテナ内実行でログ/統計が出ない（lazydocker の代表バグ） | リモート調査で機能不全 | コンテナ内実行をテスト対象に含める |
| R4 | 端末予約キー（Ctrl-D/R/S）との衝突 | 操作不能・誤動作 | 予約キーにバインドしない。将来は再割当可能に |
| R5 | 破壊的操作の誤実行（k9s の delete ダイアログ事故） | データ/コンテナ消失 | 確認ダイアログ必須、デフォルトフォーカスは非破壊側 |
| R6 | ヒントバーが常に 2 行で画面を圧迫（Zellij への不満） | 表示領域の浪費 | コンパクト 1 行モードを提供 |
| R7 | ストリーミング goroutine のリーク | メモリ増加・不安定 | context によるキャンセルと購読解除の徹底 |
| R8 | MIT 採用が Apache-2.0 指定と齟齬 | ライセンス方針の不整合 | DR-1 で明文化済み。NOTICE/ライセンス表記を同梱 |

**v2 で追加するリスク**

| # | リスク | 影響 | 対策 |
|---|---|---|---|
| R9 | compose 機能が `docker compose` プラグインへのシェルアウトに依存（単一バイナリ方針 NFR-M2 と齟齬） | compose 未導入環境で機能不全 | exec と同方針（KI-2/CR-6）。未導入なら compose ビューを無効化し原因を通知。将来は compose-go（Apache-2.0）で純 Go 化を検討（OQ-4） |
| R10 | バルク操作の部分失敗（一部成功・一部失敗） | 状態の不整合・誤認 | 全体ロールバックはせず、対象ごとの成功/失敗を集約表示。確認ダイアログで件数を明示 |
| R11 | コンテキスト切替時のストリームリーク / 旧ホストの状態混在 | リーク・誤情報表示 | 切替時に全購読を teardown し、選択行・フィルタ・ドリルダウンスタックを reset。接続失敗は R2/E3 と同じく非クラッシュ通知 |
| R12 | prune の誤実行で広範なデータ消失（build cache / volumes 含む） | データ損失 | 種別ごとに再利用可能量を明示し、選択 + 確認を必須化。既定フォーカスは安全側（R5 と同方針） |

---

## 9. Priorities & Trade-offs（優先順位とトレードオフ）

| 軸 | 固定 (Fixed) | 交渉可能 (Negotiable) |
|---|---|---|
| UX 品質・操作の分かりやすさ | ◎ 最優先（判断基準そのもの） | |
| 見た目（カラフル/モダン） | ◎ 要件 | 個別配色は調整可 |
| 機能の網羅性 | | ○ MVP を絞り段階的に拡張 |
| 開発スピード | | ○ |
| docker-compose 一級サポート | ◎ v2 最優先（M5） | 実装戦略は交渉可（OQ-4） |
| バルク操作 / 複数ホスト / メンテナンスビュー | | ○ v2（M6–M8）で段階的に |

**トレードオフ方針**: 機能数より「少数の主要フローを最高の操作性で」。迷ったら *simplicity を power に優先する*（Zellij の哲学）。

---

## 10. Sizing & Timeline（概算規模）

ROM（ラフな見積り、AIDLC の "bolt" 単位で実行想定）:

MVP（完了済み）:

- M1 Core Read: 小〜中
- M2 Core Control: 中
- M3 Multi-resource: 中
- M4 Polish: 小

v2（本更新で計画化）:

- M5 Compose: 中〜大（compose 戦略の選定・依存順・グルーピングが中核）
- M6 Bulk Actions: 小〜中（既存テーブル/確認ダイアログを再利用）
- M7 Multi-host: 中（ランタイム再接続・状態 reset が注意点）
- M8 Maintenance: 中（3 ビューの合算。各々は小〜中）

各マイルストーンは [05-units-of-work.md](./05-units-of-work.md) の Unit（MVP は U0–U8、v2 は U9–U12）に対応し、並行実行可能な単位へ分解する。

---

## 11. Open Questions & Decision Log（未解決事項と決定ログ）

### 決定ログ (Decisions)

| ID | 日付 | 決定 | 根拠 |
|---|---|---|---|
| DR-1 | 2026-06-14 | Charm 系(MIT)を Bubble Tea 指定に内包し採用。他は Apache-2.0 限定 | 見た目要件に Lip Gloss が不可欠。MIT は permissive で結合可 |
| DR-2 | 2026-06-14 | Docker SDK は公式 `docker/docker/client`、`dockerx` で隔離 | API ネゴシエーション、将来の moby 移行に備える |
| DR-3 | 2026-06-14 | ナビは k9s 流一覧 + Zellij 流ヒントバー + oxker 流分割詳細 | 各ツール比較で最良 UX |
| DR-4 | 2026-06-14 | プロダクト名は **Contalyst** | リポジトリ名に一致、container+catalyst |
| DR-5 | 2026-06-14 | compose は `docker compose` CLI へのシェルアウトを既定戦略とする（exec と同方針） | SDK に compose 機能はなく、`docker compose` の挙動互換が最も確実。exec で既にシェルアウトしておりリスク許容済み（KI-2）。純 Go 化は OQ-4 で継続検討 |
| DR-6 | 2026-06-14 | 複数ホストは Docker context（`~/.docker/contexts` + `DOCKER_CONTEXT`/`DOCKER_HOST`）を情報源とし、`dockerx` でクライアントを再生成して切替 | Docker 標準のコンテキスト機構に追従。`dockerx` 隔離（NFR-M1）の範囲で再接続を局所化 |
| DR-7 | 2026-06-14 | 旧 Out of Scope の 4 領域（compose / バルク / 複数ホスト / メンテナンス）を v2 としてスコープへ昇格。Swarm・フルマウス・ログ保存・プラグインは引き続き Out of Scope | MVP 完了を受け、競合差別化（compose）と運用価値（複数ホスト/バルク/メンテナンス）の高い順に拡張。Swarm 等は需要・複雑度の観点で据え置き |

### 未解決事項 (Open Questions)

| ID | 質問 | 暫定方針 |
|---|---|---|
| OQ-1 | Lip Gloss は v1 / v2 どちらを採用するか | Construction 時の安定度で判断。v2 は AdaptiveColor 廃止に注意（MVP: v1 採用済み / CR-1） |
| OQ-2 | 確認ダイアログは huh を使うか自前 overlay か | プロトタイプで比較し UX が良い方（MVP: 自前 overlay 採用済み） |
| OQ-3 | テーマ定義の形式（埋め込み Go / 設定ファイル） | MVP は埋め込み、後続で外部化検討 |
| OQ-4 | compose 実装は `docker compose` シェルアウト（DR-5）か compose-go（Apache-2.0）ライブラリか | M5 着手時に再評価。単一バイナリ方針（NFR-M2）と互換性のトレードオフ。暫定はシェルアウト |
| OQ-5 | コンテキスト切替時、選択行/フィルタ/履歴をホストごとに保持するか全 reset するか | M7 で UX を比較。暫定は全 reset（R11 のリーク回避を優先） |
| OQ-6 | 操作ログ（M8）は永続化するか（セッション内メモリのみ / ファイル保存） | ログのファイル保存は Out of Scope のため、暫定はセッション内リングバッファ |

---

## 12. Sign-off（承認ゲート）

AIDLC 本来の承認ゲートはユーザー方針によりスキップ。

- **MVP（M1–M4 / U0–U8）**: Construction 完了・実装済み（[`../construction/`](../construction/) 参照）。
- **v2（M5–M8 / U9–U12）**: 本ドキュメントの確定をもって Construction フェーズ（HOW の設計・実装）へ進む。
