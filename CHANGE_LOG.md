## v0.4.10-alpha1-cs1.0.3
- Rclone削除に伴い、関連コードを削除

## v0.4.10-alpha1-cs1.0.2
- opapi-codegenのソースURLをCassetteOSに変更

## v0.4.10-alpha1-cs1.0.1
- LocalStorageの設定ファイルパスを修正
- プロジェクトのツールチェインバージョンを Go 1.21 に変更。
  当初は Go 1.22 を使用予定だったが、標準ライブラリのライセンス取得で `go-licenses` に不具合が発生したため断念。
  Go 1.21 は `toolchain` ディレクティブにも対応しており、`go-licenses` の互換性も保てる安定した妥協案として採用。

## v0.4.10-alpha1-cs1.0.0
- Based on CasaOS-LocalStorage v0.4.10
- Replaced module paths to use our own GitHub fork instead of the original IceWhaleTech repository  
  (e.g., `github.com/IceWhaleTech/CasaOS-LocalStorage` → `github.com/BeesNestInc/CassetteOS-LocalStorage`)
- Deleted unused cloud drive code (Dropbox / Google Drive)