# Wallet Twin Migration Files

Folder ini berisi SQL migration files untuk database schema.

## Format Filename

```
{version}_{description}.up.sql   - Untuk apply migration
{version}_{description}.down.sql - Untuk rollback migration
```

Contoh:
- `000001_create_wallets.up.sql`
- `000001_create_wallets.down.sql`

## Menjalankan Migration

```bash
# Apply semua migrations
wallet init

# Atau menggunakan golang-migrate CLI
migrate -path ./migrations -database "postgres://..." up

# Rollback 1 step
migrate -path ./migrations -database "postgres://..." down 1
```

## Best Practices

1. **Incremental** - Setiap migration harus kecil dan fokus
2. **Reversible** - Selalu buat file `.down.sql`
3. **Idempotent** - Gunakan `IF NOT EXISTS` jika memungkinkan
4. **No Data Loss** - Hati-hati dengan ALTER/DROP
