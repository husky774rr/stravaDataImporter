# Strava Data Importer

[![CI/CD Pipeline](https://github.com/your-username/stravaDataImporter/actions/workflows/ci.yml/badge.svg)](https://github.com/your-username/stravaDataImporter/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/your-username/stravaDataImporter)](https://goreportcard.com/report/github.com/your-username/stravaDataImporter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Stravaのアクティビティデータを自動的に取得・分析し、InfluxDBに保存するWebアプリケーションです。TSS（Training Stress Score）、NP（Normalized Power）、週次・月次・年次レポートの自動生成、SNS投稿機能を提供します。

## 機能

### 🔄 自動データ同期
- Strava APIからアクティビティデータを1時間毎に自動取得
- OAuth認証による安全なデータアクセス
- アクセストークンの自動リフレッシュ（24時間毎）

### 📊 高度な分析機能
- **TSS (Training Stress Score)**: FTPベースのトレーニング負荷計算
- **NP (Normalized Power)**: 正規化パワーの算出
- **IF (Intensity Factor)**: インテンシティファクター計算
- FTPデータはCSVファイルから読み取り、日付ベースで適用

### 📈 自動集計レポート
- **週次集計**: 月曜日〜日曜日のTSS、運動時間、走行距離、獲得標高
- **月次集計**: 月初〜月末の合計データ
- **年次集計**: 年初〜年末の合計データ
- データ取得時に自動的に集計を更新

### 🌐 Webポータル
- レスポンシブデザインの美しいUI
- 最新アクティビティの詳細表示
- 週次・月次・年次サマリーの可視化
- ローディングアニメーション付きのリアルタイム更新

### 🐦 SNS自動投稿
- アクティビティ完了後のTwitter自動投稿
- 直近1週間のトレンドグラフ画像生成
- 日本語でのフォーマット済み投稿内容

## アーキテクチャ

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Strava API    │───▶│  Go Web App     │───▶│   InfluxDB      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                               │
                               ▼
                       ┌─────────────────┐
                       │   Twitter API   │
                       └─────────────────┘
```

## クイックスタート

### 前提条件

- Go 1.21以上
- Docker & Docker Compose
- Strava Developer Account
- Twitter Developer Account（SNS投稿機能を使用する場合）

### 1. リポジトリのクローン

```bash
git clone https://github.com/your-username/stravaDataImporter.git
cd stravaDataImporter
```

### 2. 環境設定

```bash
# 環境ファイルの作成
make setup-env

# .envファイルを編集してAPI認証情報を設定
vi .env
```

### 3. 開発環境のセットアップ

```bash
# 開発ツールのインストールと依存関係の解決
make init-project

# Dockerサービスの起動（InfluxDB、Grafana）
make docker-up
```

### 4. アプリケーションの起動

```bash
# 開発モードで起動
make run-dev

# または本番ビルドして起動
make build && make run
```

アプリケーションは http://localhost:8080 でアクセス可能です。

## 設定

### 環境変数

| 変数名 | 説明 | デフォルト |
|--------|------|------------|
| `PORT` | サーバーポート | `8080` |
| `LOG_LEVEL` | ログレベル (debug, info, warn, error) | `info` |
| `STRAVA_CLIENT_ID` | Strava API クライアントID | - |
| `STRAVA_CLIENT_SECRET` | Strava API クライアントシークレット | - |
| `STRAVA_REDIRECT_URI` | OAuth リダイレクトURI | `http://localhost:8080/auth/callback` |
| `INFLUXDB_URL` | InfluxDB URL | `http://localhost:8086` |
| `INFLUXDB_TOKEN` | InfluxDB 認証トークン | - |
| `INFLUXDB_ORG` | InfluxDB 組織名 | `strava` |
| `INFLUXDB_BUCKET` | InfluxDB バケット名 | `activities` |
| `TOKEN_REFRESH_INTERVAL` | トークンリフレッシュ間隔 | `24h` |
| `DATA_IMPORT_INTERVAL` | データインポート間隔 | `1h` |

### FTPデータの設定

`conf/ftp.csv`ファイルでFTP（Functional Threshold Power）データを管理します：

```csv
date,ftp
2024-01-01,170
2024-08-29,191
2024-10-27,217
2025-02-05,248
```

## 開発

### 開発環境

このプロジェクトはVS Code + Dev Containerでの開発を推奨しています。

```bash
# Dev Containerで開く
code .
# "Reopen in Container"を選択
```

### テストの実行

```bash
# 単体テスト
make test

# カバレッジ付きテスト
make test-coverage

# 統合テスト（Docker必要）
make test-integration

# ベンチマーク
make benchmark
```

### コード品質

```bash
# リンター実行
make lint

# コードフォーマット
make format

# セキュリティチェック
make security
```

### ビルド

```bash
# ローカルビルド
make build

# 全プラットフォーム向けビルド
make build-all

# Dockerイメージビルド
make docker-build
```

## デプロイメント

### Docker Compose

```bash
# 本番環境用のDocker Composeファイルを使用
docker-compose -f docker/docker-compose.yml up -d
```

### GitHub Actions

プロジェクトには以下のワークフローが含まれています：

- **CI/CD Pipeline** (`.github/workflows/ci.yml`): テスト、ビルド、Docker イメージ作成
- **Release** (`.github/workflows/release.yml`): タグベースの自動リリース

### Ubuntu 24.04 LTS デプロイメント

```bash
# Dockerのインストール
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# リポジトリのクローン
git clone https://github.com/your-username/stravaDataImporter.git
cd stravaDataImporter

# 環境設定
cp .env.example .env
# .envファイルを編集

# サービス起動
docker-compose -f docker/docker-compose.yml up -d
```

## API エンドポイント

| エンドポイント | メソッド | 説明 |
|----------------|----------|------|
| `/` | GET | インデックスページ |
| `/login` | GET | ログインページ |
| `/portal` | GET | ポータルページ（認証必要） |
| `/auth/login` | GET | Strava OAuth開始 |
| `/auth/callback` | GET | OAuth コールバック |
| `/auth/logout` | GET | ログアウト |
| `/api/activities` | GET | アクティビティ一覧取得 |
| `/api/summaries/weekly` | GET | 週次サマリー取得 |
| `/api/summaries/monthly` | GET | 月次サマリー取得 |
| `/api/summaries/yearly` | GET | 年次サマリー取得 |

## データベーススキーマ

### InfluxDB Measurements

#### activities
アクティビティの詳細データ

| Field | Type | Description |
|-------|------|-------------|
| `distance` | float | 距離 (km) |
| `moving_time` | int | 運動時間 (秒) |
| `elevation_gain` | float | 獲得標高 (m) |
| `calories` | float | 消費カロリー |
| `average_power` | float | 平均パワー (W) |
| `normalized_power` | float | 正規化パワー (W) |
| `tss` | float | Training Stress Score |
| `intensity_factor` | float | インテンシティファクター |
| `ftp` | int | FTP (W) |

#### summaries
週次・月次・年次サマリーデータ

| Field | Type | Description |
|-------|------|-------------|
| `total_distance` | float | 合計距離 (km) |
| `total_moving_time` | int | 合計運動時間 (秒) |
| `total_elevation_gain` | float | 合計獲得標高 (m) |
| `total_tss` | float | 合計TSS |
| `activity_count` | int | アクティビティ数 |

## 監視とメトリクス

- **Grafana**: http://localhost:3000 でダッシュボードにアクセス
- **InfluxDB UI**: http://localhost:8086 でデータベースを管理
- アプリケーションログは構造化ログ（slog）で出力

## トラブルシューティング

### よくある問題

**Q: Strava認証に失敗する**
A: 
- Strava Developer DashboardでリダイレクトURIが正しく設定されているか確認
- `STRAVA_CLIENT_ID`と`STRAVA_CLIENT_SECRET`が正しく設定されているか確認

**Q: InfluxDBに接続できない**
A:
- InfluxDBが起動しているか確認: `docker-compose logs influxdb`
- 認証トークンが正しく設定されているか確認

**Q: テストが失敗する**
A:
- テスト用のInfluxDBが起動しているか確認: `make docker-up`
- 環境変数が正しく設定されているか確認

### ログの確認

```bash
# アプリケーションログ
docker-compose logs app

# InfluxDBログ
docker-compose logs influxdb

# 全サービスのログ
make docker-logs
```

## コントリビューション

1. このリポジトリをフォーク
2. 機能ブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add amazing feature'`)
4. ブランチにプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

### 開発ガイドライン

- Goの標準的なコーディング規約に従う
- すべての新機能にはテストを含める
- コミットメッセージは明確で説明的にする
- プルリクエスト前に`make lint`と`make test`を実行

## ライセンス

このプロジェクトはMITライセンスの下で公開されています。詳細は [LICENSE](LICENSE) ファイルを参照してください。

## 謝辞

- [Strava API](https://developers.strava.com/) - アクティビティデータの提供
- [InfluxDB](https://www.influxdata.com/) - 時系列データベース
- [Gin Web Framework](https://gin-gonic.com/) - Go言語Webフレームワーク

## サポート

問題や質問がある場合は、[GitHub Issues](https://github.com/your-username/stravaDataImporter/issues)でお気軽にお知らせください。
