# Contalyst — User Stories & Acceptance Criteria

> INVEST 準拠のユーザーストーリーと、Given/When/Then 形式の受け入れ基準。
> 対応要件は [01-requirements.md](./01-requirements.md)、ペルソナは [02-personas.md](./02-personas.md)。

---

### US-1: コンテナ一覧を状態の色分けで把握する  *(P0)*
**As a** 開発者/運用者, **I want** 全コンテナを状態の色付きで一覧したい, **so that** どれが動いていて、どれが落ちているか一目で分かる。
- **Given** Docker に複数コンテナがある **When** Contalyst を起動する **Then** 名前/イメージ/状態/ステータス/ポートが一覧表示される。
- **Given** 一覧表示中 **Then** running/exited/paused 等が異なる色で示される。
- **Given** コンテナの状態が変化する **When** 一定時間が経つ **Then** 一覧が自動更新される。
- 要件: FR-C1, FR-C2, NFR-U2

### US-2: ログをリアルタイムに追尾する  *(P0)*
**As a** ユーザー, **I want** 選択コンテナのログを follow したい, **so that** 何が起きているかを即座に追える。
- **Given** コンテナを選択 **When** ログを開く **Then** 既存ログが表示され、新規行が遅延なく追記される。
- **Given** ログ表示中 **Then** 時間窓で勝手に切り詰められず過去ログも見える。
- **Given** ログ追尾中 **When** 上にスクロール **Then** 自動スクロールが一時停止し、最下部で再開する。
- 要件: FR-C3, FR-C4, NFR-P2

### US-3: コンテナを起動/停止/再起動する  *(P0)*
**As a** ユーザー, **I want** 選択コンテナを start/stop/restart したい, **so that** CLI に戻らず制御できる。
- **Given** 停止中コンテナを選択 **When** start キーを押す **Then** 起動し一覧に反映される。
- **Given** 稼働中コンテナを選択 **When** stop/restart **Then** 該当操作が実行され状態が更新される。
- **Given** 操作が失敗 **Then** クラッシュせずエラーが通知される。
- 要件: FR-C6, FR-C7, NFR-R4

### US-4: リソース使用状況を確認する  *(P1)*
**As a** SRE, **I want** コンテナの CPU/メモリ/ネット/IO をリアルタイムで見たい, **so that** リソースを食っている犯人を特定できる。
- **Given** コンテナを選択 **When** 統計を開く **Then** CPU%/メモリ/ネット/IO が定期更新される。
- **Given** CPU% **Then** cpu_stats と precpu_stats の差分から正しく算出される。
- 要件: FR-C11

### US-5: コンテナ内シェルへ exec する  *(P0)*
**As a** 開発者, **I want** 選択コンテナのシェルに入りたい, **so that** 中を直接調べられる。
- **Given** 稼働中コンテナを選択 **When** exec **Then** インタラクティブシェルが起動する。
- **Given** 既定シェルが無い **Then** `/bin/sh` にフォールバックする。
- **Given** シェルを抜ける **Then** Contalyst の一覧に戻る。
- 要件: FR-C9

### US-6: 破壊的操作を安全に行う  *(P0)*
**As a** ユーザー, **I want** 削除/kill の前に確認したい, **so that** 誤って消さない。
- **Given** コンテナを選択 **When** 削除キー **Then** 確認ダイアログが出る。
- **Given** 確認ダイアログ **Then** 既定フォーカスは「キャンセル（非破壊）」側。
- **Given** 明示的に承認 **Then** 削除が実行され一覧から消える。
- 要件: FR-C8, FR-NAV8, NFR-R3

### US-7: 一覧をフィルタする  *(P0)*
**As a** 多数コンテナを扱う運用者, **I want** ファジー検索で絞り込みたい, **so that** 目的の対象に素早く到達できる。
- **Given** 一覧表示中 **When** `/` を押し文字を入力 **Then** 一致行のみに絞られる。
- **Given** フィルタ中 **When** `Esc` **Then** フィルタが解除される。
- 要件: FR-NAV3

### US-8: コマンドパレットでリソースを切り替える  *(P1)*
**As a** 運用者, **I want** `:` で Docker の語彙を使ってリソース種別を切替たい, **so that** メニューを辿らず目的のビューへ飛べる。
- **Given** 任意のビュー **When** `:images` 等を入力 **Then** 該当リソース一覧へ切り替わる。
- **Given** `:` 入力中 **Then** 候補が autosuggest される。
- 要件: FR-NAV2

### US-9: SSH 越し / コンテナ内でも動く  *(P0)*
**As a** SRE, **I want** リモートやコンテナ内でも全機能が使えること, **so that** 本番調査で詰まらない。
- **Given** SSH 接続先 / コンテナ内で実行 **Then** 一覧・ログ・統計が正しく表示される。
- **Given** 任意の Engine バージョン **Then** API ネゴシエーションで動作する。
- 要件: FR-E2, NFR-R2

### US-10: ヒントバーで操作を学習不要にする  *(P0)*
**As a** 新規ユーザー, **I want** 画面に有効なキー操作が常に出ていること, **so that** ドキュメントを読まずに使える。
- **Given** いずれかのビュー **Then** 下部に現在有効な単キー操作のヒントバーが表示される。
- **Given** ビューが変わる **Then** ヒントの内容も追従する。
- **Given** `?` を押す **Then** コンテキスト別のキーバインド一覧が出る。
- 要件: FR-NAV5, FR-NAV7, NFR-U1

### US-11: 見た目を整える / テーマを選ぶ  *(P1)*
**As a** ユーザー, **I want** カラフルで themeable な見た目, **so that** 使っていて気持ちよい。
- **Given** 起動 **Then** truecolor/256 色でカラフルに表示される。
- **Given** 設定 **When** テーマ切替 **Then** 配色が変わる。
- **Given** 角丸フレーム等 **Then** opt-in で切替できる。
- 要件: FR-NAV9, NFR-U3, NFR-U5

### US-12: コンテナを inspect する  *(P1)*
**As a** ユーザー, **I want** 選択コンテナの詳細を整形表示したい, **so that** 設定やマウント等を確認できる。
- **Given** コンテナを選択 **When** inspect **Then** 整形された詳細が表示される。
- 要件: FR-C10

### US-13: イメージ/ボリューム/ネットワークを管理する  *(P1)*
**As a** 運用者, **I want** イメージ/ボリューム/ネットワークの一覧と削除/prune, **so that** 不要リソースを掃除できる。
- **Given** 各リソースビュー **Then** 一覧が表示される。
- **Given** 対象を選択 **When** 削除/prune **Then** 確認後に実行される。
- 要件: FR-I1, FR-I2, FR-V1, FR-N1

### US-14: 設定なしで起動する  *(P0)*
**As a** ホームラボ運用者, **I want** インストールしてすぐ起動できること, **so that** 設定に時間をかけたくない。
- **Given** Docker が動くマシン **When** バイナリを実行 **Then** デフォルトソケットに接続し一覧が出る。
- **Given** Docker に接続不可 **Then** 原因の分かるエラーを表示し異常終了しない。
- 要件: FR-E1, FR-E3, NFR-M2
