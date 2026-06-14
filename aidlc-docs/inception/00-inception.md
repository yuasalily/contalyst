# Contalyst — Inception Document

> AIDLC (AI-Driven Development Life Cycle) の **Inception フェーズ** に相当するドキュメントです。
> 「何を、なぜ作るのか (WHAT / WHY)」を定義します。「どう作るか (HOW)」の詳細は Construction フェーズで扱います。

- **Project**: Contalyst — A modern, colorful Docker TUI
- **Phase**: Inception
- **Status**: Drafted (承認ゲートはスキップ。判断は本ドキュメント内に記録)
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

### 5.1 In Scope（MVP に含む）

- **コンテナ管理**: 一覧（live 更新）、start / stop / restart / pause / unpause / remove / kill。
- **ログ**: リアルタイム追尾（follow）、スクロール、検索、タイムスタンプ表示切替、truncate しない取得。
- **exec**: コンテナ内シェルへの起動（`/bin/sh` フォールバック）。
- **inspect**: 整形された詳細情報の表示。
- **stats**: CPU / メモリ / ネットワーク / IO のリアルタイム表示。
- **イメージ / ボリューム / ネットワーク**: 一覧と基本操作（削除・prune）。
- **ナビゲーション**: `:` コマンドパレット（リソース種別切替）、`/` ファジーフィルタ、ドリルダウン/戻る + パンくず。
- **操作性 UI**: 常時表示キーヒントバー、`?` コンテキストヘルプ、破壊的操作の確認ダイアログ。
- **見た目**: 1 つ以上の良質なカラーテーマ、状態の色分け、角丸フレーム（opt-in）。

### 5.2 Out of Scope（MVP では作らない / 後続）

- docker-compose の一級サポート（up/down/rebuild --no-cache、依存順）→ **競合の最大の隙。Post-MVP の最優先候補**。
- Docker Swarm（nodes / services / stacks）。
- 複数 Docker ホスト / リモートコンテキストの切替 UI。
- イメージレイヤービュー、prune ダッシュボード、操作ログ（lazygit の `@` 相当）。
- バルク/複数選択操作、マウスのフルサポート、ログのファイル保存。
- プラグイン機構、設定ファイルによる詳細カスタマイズ（キーバインド再割当は MVP で最小限）。

### 5.3 Release Roadmap（増分計画）

- **M1 — Core Read**: コンテナ一覧 + ログ追尾 + inspect + ヒントバー + 1 テーマ。
- **M2 — Core Control**: start/stop/restart/remove + 確認ダイアログ + stats + exec。
- **M3 — Multi-resource**: images / volumes / networks + `:` コマンドパレット + `/` フィルタ。
- **M4 — Polish**: テーマ複数化、ヘルプ充実、コンパクトヒントバー、パフォーマンス調整。
- **Post-MVP**: docker-compose 一級サポート、複数ホスト、バルク操作。

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

---

## 9. Priorities & Trade-offs（優先順位とトレードオフ）

| 軸 | 固定 (Fixed) | 交渉可能 (Negotiable) |
|---|---|---|
| UX 品質・操作の分かりやすさ | ◎ 最優先（判断基準そのもの） | |
| 見た目（カラフル/モダン） | ◎ 要件 | 個別配色は調整可 |
| 機能の網羅性 | | ○ MVP を絞り段階的に拡張 |
| 開発スピード | | ○ |
| docker-compose 等の高度機能 | | △ Post-MVP |

**トレードオフ方針**: 機能数より「少数の主要フローを最高の操作性で」。迷ったら *simplicity を power に優先する*（Zellij の哲学）。

---

## 10. Sizing & Timeline（概算規模）

ROM（ラフな見積り、AIDLC の "bolt" 単位で実行想定）:

- M1 Core Read: 小〜中
- M2 Core Control: 中
- M3 Multi-resource: 中
- M4 Polish: 小

各マイルストーンは [05-units-of-work.md](./05-units-of-work.md) の Unit に対応し、並行実行可能な単位へ分解する。

---

## 11. Open Questions & Decision Log（未解決事項と決定ログ）

### 決定ログ (Decisions)

| ID | 日付 | 決定 | 根拠 |
|---|---|---|---|
| DR-1 | 2026-06-14 | Charm 系(MIT)を Bubble Tea 指定に内包し採用。他は Apache-2.0 限定 | 見た目要件に Lip Gloss が不可欠。MIT は permissive で結合可 |
| DR-2 | 2026-06-14 | Docker SDK は公式 `docker/docker/client`、`dockerx` で隔離 | API ネゴシエーション、将来の moby 移行に備える |
| DR-3 | 2026-06-14 | ナビは k9s 流一覧 + Zellij 流ヒントバー + oxker 流分割詳細 | 各ツール比較で最良 UX |
| DR-4 | 2026-06-14 | プロダクト名は **Contalyst** | リポジトリ名に一致、container+catalyst |

### 未解決事項 (Open Questions)

| ID | 質問 | 暫定方針 |
|---|---|---|
| OQ-1 | Lip Gloss は v1 / v2 どちらを採用するか | Construction 時の安定度で判断。v2 は AdaptiveColor 廃止に注意 |
| OQ-2 | 確認ダイアログは huh を使うか自前 overlay か | プロトタイプで比較し UX が良い方 |
| OQ-3 | テーマ定義の形式（埋め込み Go / 設定ファイル） | MVP は埋め込み、後続で外部化検討 |

---

## 12. Sign-off（承認ゲート）

AIDLC 本来の承認ゲートはユーザー方針によりスキップ。本ドキュメントの確定をもって Construction フェーズ（HOW の設計・実装）へ進む。
