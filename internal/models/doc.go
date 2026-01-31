// Package models berisi entity/domain objects untuk aplikasi.
//
// Dalam Clean Architecture, models (atau entities) adalah:
// - Pure business objects tanpa dependencies ke layer lain
// - Berisi business rules dan validation
// - Tidak tahu tentang database, UI, atau framework
//
// Entity vs DTO:
// - Entity: representasi internal business object
// - DTO: Data Transfer Object untuk komunikasi antar layer
//
// Package ini berisi struct untuk:
// - Wallet: dompet/akun keuangan
// - Category: kategori transaksi
// - Transaction: pemasukan/pengeluaran
// - Transfer: transfer antar wallet
// - Budget: anggaran per kategori
// - Recurring: transaksi berulang
// - Goal: target tabungan
package models

// File ini akan berisi common types dan interfaces
// yang digunakan oleh entity lainnya.
//
// Contoh:
// - BaseEntity dengan ID dan timestamps
// - Validation interface
// - Error types

// TODO: Add common types dan validation helpers
