-- Rollback: Remove seeded categories
-- HATI-HATI: Ini akan menghapus semua kategori default!
-- Kategori custom user juga akan terhapus jika tidak ada constraint

-- Hapus semua kategori
-- ON DELETE CASCADE akan menghapus sub-kategori juga
TRUNCATE TABLE categories CASCADE;
