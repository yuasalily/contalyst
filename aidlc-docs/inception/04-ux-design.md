# Contalyst — UI/UX & Visual Design（UI/UX・ビジュアルデザイン方針）

> 「Zellij のようにモダンでカラフル、見ればわかる」を実現するための画面・操作・配色の方針。
> 既存ツール調査（lazydocker / ctop / dry / oxker / k9s / lazygit / Zellij）の比較から得た意思決定（DR-3）に基づく。
> これは Inception レベルの方針であり、確定実装は Construction で詰める。

---

## 1. デザイン原則

1. **Discoverability first** — ドキュメントを読まずに使える。常時ヒントバー + `?` ヘルプ + コマンドパレット。
2. **Simplicity over power** — 機能数より主要フローの操作性（Zellij の哲学）。
3. **Safe by default** — 破壊的操作は確認、既定フォーカスは非破壊側。
4. **Keyboard-first, vim-friendly** — 主操作はキーボード。vim ライクなキーを一貫採用。
5. **Colorful but legible** — 色は意味（状態/危険/フォーカス）に紐づける。装飾過多にしない。
6. **No surprises** — 端末予約キーを奪わない、ログを勝手に切らない、状態は常に最新。

---

## 2. 画面レイアウト

3 段構成を基本とする。

```
┌─ Contalyst ─ Containers ─ /api ────────────────[ ◐ docker 29.5 ]┐  ← ヘッダ: アプリ名 / リソース種別 / パンくず / 接続状態
│ NAME            IMAGE              STATE     STATUS        PORTS  │
│ ● api           myorg/api:latest   running   Up 2 min      :8080  │  ← メイン: live テーブル（k9s 流）
│ ● db            postgres:16        running   Up 1 h         5432  │     行頭ドットの色で状態表現
│ ○ worker        myorg/worker       exited    Exited (1)     -     │     選択行はハイライト
│ ‖ cache         redis:7            paused    Paused         6379  │
│                                                                   │
│                                                                   │
├───────────────────────────────────────────────────────────────┤
│ ↑↓ move  ⏎ logs  s start/stop  r restart  e exec  d delete  : cmd │  ← ヒントバー（Zellij 流・常時・カラフル）
│ / filter   ? help   q quit                                        │     コンパクト時は1行
└───────────────────────────────────────────────────────────────┘
```

ドリルダウン（例: コンテナ → 詳細）では oxker 流の「全部見える」分割詳細を採用:

```
┌─ Contalyst ─ Containers ▸ api ─────────────────────────────────┐  ← パンくず: 種別 ▸ 対象
│ ┌─ Logs (follow) ─────────────────┐┌─ Stats ──────────────────┐ │
│ │ 12:00:01 server started         ││ CPU   ▕███▏ 18%          │ │  ← 左: ログ追尾 / 右: 統計
│ │ 12:00:03 GET /health 200        ││ MEM   ▕██▏  256MB / 1GB  │ │
│ │ 12:00:05 ...                    ││ NET   ↓1.2MB ↑0.3MB      │ │
│ │                                 ││ IO    r:10MB w:2MB       │ │
│ └─────────────────────────────────┘└──────────────────────────┘ │
├───────────────────────────────────────────────────────────────┤
│ ↑↓ scroll  f follow  / search  t timestamps  i inspect  ⎋ back   │
└───────────────────────────────────────────────────────────────┘
```

- ログパネルの高さ調整（oxker の `-`/`=` 相当）は P1。
- 端末幅が狭い場合は分割を縦積みにフォールバック。

---

## 3. ナビゲーションモデル（DR-3）

- **背骨 = k9s 流リソース一覧**: 1 画面 1 リソース種別の live テーブル。起動時は Containers。
- **`:` コマンドパレット**: Docker の語彙でリソース切替・操作。`:ps`/`:containers`, `:images`, `:volumes`, `:networks`, `:help`, `:theme`, `:quit`。autosuggest と一覧（`:` 単体）を提供。
- **`/` フィルタ**: ファジー絞り込み。`Esc` で解除。将来は反転フィルタ（k9s の `/!`）も検討。
- **ドリルダウン**: `Enter` で対象の詳細へ、`Esc` で戻る。下部 or ヘッダにパンくず。`[`/`]` で履歴移動は P1。
- **モードは最小**: `:`（コマンド）と `/`（フィルタ）だけを軽量モードとし、それ以外は直接単キー操作。重いモーダルにしない。

---

## 4. キーバインド方針

> 端末予約キー（`Ctrl-D`/`Ctrl-R`/`Ctrl-S`）は使わない（R4）。将来は再割当可能にする。

**グローバル**
| キー | 動作 |
|---|---|
| `j`/`k`, `↓`/`↑` | 行移動 |
| `g`/`G` | 先頭/末尾 |
| `Enter` | ドリルダウン |
| `Esc` | 戻る / モード解除 |
| `/` | フィルタ |
| `:` | コマンドパレット |
| `?` | ヘルプ |
| `Tab` | コンパクトヒントバー切替 (P1) |
| `q` / `Ctrl+C` | 終了 |

**コンテナビュー**（単キー、ヒントバーに常時表示）
| キー | 動作 |
|---|---|
| `s` | start/stop トグル |
| `r` | restart |
| `p` | pause/unpause (P1) |
| `e` | exec シェル |
| `l` / `Enter` | ログ/詳細 |
| `i` | inspect |
| `d` | delete（確認あり） |
| `K` | kill（確認あり, P1） |

**ログ/詳細ビュー**: `f` follow トグル, `/` 検索, `t` タイムスタンプ, `Esc` 戻る。

---

## 5. カラー & テーマ

Zellij のテーマ思想（`base` + `emphasis_0..3` + 状態色）を踏襲し、Lip Gloss で実装。MVP は埋め込みテーマ「Catalyst（既定・暗背景・ネオン寄り）」を用意し、追加テーマは段階的に。

**意味に紐づく色（state semantics）**
| 用途 | 色の意図 |
|---|---|
| running | green 系（健全） |
| exited (0) | gray 系（停止・正常） |
| exited (非0) / dead | red 系（異常） |
| paused | yellow/amber 系 |
| restarting / created | blue/cyan 系 |
| 選択行 | アクセント背景 + 高コントラスト前景 |
| 危険操作（削除ダイアログ） | red アクセント。ただし既定フォーカスは安全側 |
| ヒントバーのキー文字 | アクセント色、説明文は控えめ色 |

**描画方針**
- truecolor 優先、256/16 色へ degrade（NFR-U3）。
- 角丸フレーム・枠の装飾は opt-in（NFR-U5）。フォーカス枠は色で強調。
- アイコン/グリフは ASCII フォールバックを用意（●○‖ 等が出ない端末向け）。

---

## 6. ダイアログ & フィードバック

- **確認ダイアログ**: 削除/kill/prune で表示。タイトルに対象名、ボタンは `[Cancel]`（既定フォーカス） / `[Delete]`。Enter 連打で誤実行しない配置（k9s #961 の教訓）。
- **トースト/ステータス行**: 操作結果（成功/失敗）を一時表示。失敗時はエラー要約。
- **ローディング**: ストリーム接続・操作中は spinner。
- **空状態**: コンテナ 0 件・Docker 未接続時は、原因と次アクション（例: `DOCKER_HOST` 確認）を案内。

---

## 7. 既存ツールから「避ける」こと（アンチパターン）

- ログを既定で 1 時間に切り詰めて空に見せる（lazydocker の混乱）→ 過去ログ＋追尾を既定に。
- 破壊ダイアログの既定ボタンが破壊側（k9s 事故）→ 常に安全側を既定フォーカス。
- 端末予約キーへのバインド（Zellij/k9s の衝突）→ 使わない。
- コンテナ内/SSH でログ・統計が出ない（lazydocker の代表バグ）→ demux 分岐＋テスト。
- ヒントバーが常時 2 行で画面圧迫（Zellij への不満）→ コンパクト 1 行モード。
- API バージョン不一致で起動時クラッシュ（dry/ctop）→ ネゴシエーション必須。

---

## 8. アクセシビリティ / 端末互換

- 色だけに依存せず、状態はテキスト列でも表現（色覚多様性配慮）。
- 最小幅/高さを下回る端末ではレイアウトを段階的に簡略化。
- マウスは P1（クリックでの行選択・タブ切替）。MVP はキーボードで完結。

---

## 9. v2 UI/UX 拡張（M5–M8）

> [00-inception.md](./00-inception.md) §5.2 でスコープへ昇格した 4 領域の画面・操作方針。MVP の 3 段レイアウト・k9s 流一覧・確認ダイアログ・操作ログ無しの原則をそのまま継承する。要件は [01-requirements.md](./01-requirements.md) 「v2 機能要件」。

### 9.1 docker-compose（M5 / FR-CMP*）

新リソース種別 `compose` を `:` コマンドパレットに追加（`:compose`）。一覧はプロジェクト単位、ドリルダウンでサービス（コンテナ）一覧 → さらに従来の分割詳細へ。

```
┌─ Contalyst ─ Compose ─ /api ───────────────────[ ◐ docker 29.5 ]┐
│ PROJECT         SERVICES  RUNNING  STATE                         │
│ ▸ myapp          5         5/5      up                           │  ← プロジェクト集約（稼働数/状態）
│ ▸ data-stack     3         1/3      degraded                     │     状態色: up=green / degraded=amber / down=gray
│ ▸ legacy         2         0/2      down                         │
├───────────────────────────────────────────────────────────────┤
│ ⏎ services  u up  d down  r restart  b rebuild  B rebuild --no-cache  l logs │
│ : cmd   / filter   @ ops   ? help   q quit                       │
└───────────────────────────────────────────────────────────────┘
```

- **キー**（compose プロジェクトビュー）: `u` up / `d` down（**確認あり**・破壊系） / `r` restart / `b` rebuild / `B` rebuild `--no-cache` / `Enter` サービス一覧 / `l` プロジェクト横断ログ。
- 依存順（`depends_on`）は `docker compose` に委譲（DR-5）。操作中は spinner、結果はトースト + 操作ログ（§9.4）。
- `docker compose` 未導入時は一覧の代わりに「compose 無効: `docker compose` が見つかりません」を案内（R9 / 空状態の流儀）。

### 9.2 バルク / 複数選択（M6 / FR-B*）

通常の一覧（コンテナ）に複数選択を重畳。マーク中は専用ヒントを表示し、操作はマーク集合へ適用。

```
│ ✓ ● api           myorg/api:latest   running   Up 2 min   :8080  │  ← ✓ = マーク済み（アクセント色）
│   ● db            postgres:16        running   Up 1 h      5432  │
│ ✓ ○ worker        myorg/worker       exited    Exited (1)   -    │
├───────────────────────────────────────────────────────────────┤
│ 2 selected ─ s start  S stop  r restart  d delete  Space (un)mark  a all  Esc clear │
└───────────────────────────────────────────────────────────────┘
```

- **キー**: `Space` マーク/解除トグル / `a` 全選択⇄全解除 / `Esc` マーク解除。マークが 1 件以上ある間、`s`/`S`/`r`/`d` はマーク集合へ一括適用（無マーク時は従来どおりカーソル行のみ）。
- バルク `delete` は対象件数を明示した 1 ダイアログ（例: `Delete 2 containers? [Cancel] [Delete]`、既定フォーカス Cancel）。
- 実行は並行、完了後に `2 ok / 0 failed` のように集約。失敗があれば操作ログで詳細（§9.4 / R10）。

### 9.3 複数ホスト / コンテキスト切替（M7 / FR-H*）

`:context`（別名 `:hosts`）でコンテキスト選択オーバーレイ。ヘッダ右にアクティブホストを常時表示。

```
┌─ Contalyst ─ Containers ─ /api ──────────[ ◐ prod (ssh://prod) ]┐  ← ヘッダにアクティブ context/host
│            ┌─ Switch context ──────────────────────┐            │
│            │ ● default     unix:///var/run/docker… │            │  ← ● = 現在
│            │   prod        ssh://prod                │            │
│            │   staging     tcp://10.0.0.5:2376       │            │
│            └────────────────────────────────────────┘            │
```

- 選択 → `Enter` で切替。切替時は全ストリームを teardown、選択/フィルタ/ドリルダウンを reset（R11 / OQ-5）。接続失敗は MVP の接続失敗画面と同じ流儀で原因表示（R2/E3）。

### 9.4 メンテナンスビュー（M8 / FR-L1, FR-PR*, FR-OL*）

**イメージレイヤービュー**: イメージ一覧から `Enter`（または `L`）でレイヤー（history）へドリルダウン。

```
┌─ Contalyst ─ Images ▸ myorg/api:latest ▸ Layers ───────────────┐
│ SIZE     CREATED BY                                             │
│ 78MB     RUN apt-get install -y build-essential …               │
│ 12MB     COPY . /app                                            │
│  0B      CMD ["/app/server"]                                    │
└───────────────────────────────────────────────────────────────┘
```

**prune ダッシュボード**: `:prune` で再利用可能量を種別ごとに集計。`Space` で種別選択 → `Enter` で確認後 prune（R12）。

```
┌─ Contalyst ─ Prune ────────────────────────────────────────────┐
│   TYPE            RECLAIMABLE   ITEMS                            │
│ ✓ images           1.8 GB        12   (dangling + unused)        │
│ ✓ build cache      940 MB         –                             │
│   volumes          320 MB         4                             │
│   networks           –            2                             │
│   stopped containers 60 MB         3                            │
├───────────────────────────────────────────────────────────────┤
│ Space select  ⏎ prune selected (2.7 GB)  Esc cancel             │  ← 既定フォーカスは安全側
└───────────────────────────────────────────────────────────────┘
```

**操作ログ（lazygit `@` 相当）**: `@` でセッション内の操作履歴オーバーレイを開く（OQ-6: 永続化しない）。

```
┌─ Operation log ────────────────────────────────────────────────┐
│ 12:03:11  stop      api                ok                       │
│ 12:03:30  compose down  data-stack     ok                       │
│ 12:04:02  prune     images+cache       ok  (2.7 GB reclaimed)   │
│ 12:05:18  remove    worker             FAILED: in use           │  ← 失敗は red、エラー要約付き
└───────────────────────────────────────────────────────────────┘
```

### 9.5 キーバインド / パレット追加サマリ

| 追加キー / コマンド | コンテキスト | 動作 |
|---|---|---|
| `:compose` | グローバル | compose プロジェクト一覧へ切替 |
| `u`/`d`/`r`/`b`/`B` | compose プロジェクト | up / down(確認) / restart / rebuild / rebuild --no-cache |
| `Space` / `a` | コンテナ一覧 | マーク トグル / 全選択⇄解除 |
| `S` | コンテナ一覧（マーク時） | バルク stop（`s` は start、大文字で stop に分離） |
| `:context` / `:hosts` | グローバル | コンテキスト切替オーバーレイ |
| `L` / `Enter` | イメージ一覧 | レイヤービューへドリルダウン |
| `:prune` | グローバル | prune ダッシュボード |
| `@` | グローバル | 操作ログオーバーレイ |

> いずれも端末予約キー（`Ctrl-D/R/S`）を避ける方針を維持（R4 / NFR-U4）。新キーは追加後 `?` ヘルプとヒントバーに反映する（discoverability / NFR-U1）。
