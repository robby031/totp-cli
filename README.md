# totp-cli

CLI TOTP

## Pasang versi terbaru:

#### macOS / Linux:

```bash
curl -fsSL https://github.com/robby031/totp-cli/releases/latest/download/install.sh | sh
```

#### Windows (PowerShell):

```bash
irm https://github.com/robby031/totp-cli/releases/latest/download/install.ps1 | iex
```

## Pasang versi tertentu:

```bash
curl -fsSL https://github.com/robby031/totp-cli/releases/latest/download/install.sh | TOTP_VERSION=v0.1.42 sh
```

```bash
$env:TOTP_VERSION="v0.1.42"; irm https://github.com/robby031/totp-cli/releases/latest/download/install.ps1 | iex
```

## Build dari source

```go
go mod tidy
go build -o totp ./cmd/totp
```

## Penggunaan

| Perintah                  | Deskripsi                                            |
| :------------------------ | :--------------------------------------------------- |
| `totp add <nama_akun>`    | Tambah akun baru (paste secret Base32 setelah enter) |
| `totp code <nama_akun>`   | Tampilkan kode untuk satu akun tertentu              |
| `totp`                    | Tampilkan kode semua akun (sekali tampil)            |
| `totp watch`              | Pantau kode semua akun (update tiap detik)           |
| `totp watch <nama_akun>`  | Pantau kode satu akun spesifik secara real-time      |
| `totp list`               | Lihat daftar semua akun tersimpan                    |
| `totp remove <nama_akun>` | Hapus akun dari daftar                               |
| `totp help`               | Tampilkan daftar bantuan perintah                    |
| `totp --version`          | Cek versi aplikasi                                   |
