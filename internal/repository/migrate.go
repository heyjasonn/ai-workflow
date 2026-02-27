package repository

import (
    "context"
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "sort"
    "strings"

    "gorm.io/gorm"
)

type Migration struct {
    Version string
    Path    string
}

func ApplyMigrations(ctx context.Context, db *gorm.DB, dir string) error {
    if err := ensureMigrationsTable(ctx, db); err != nil {
        return err
    }

    migrations, err := loadMigrations(dir)
    if err != nil {
        return err
    }

    for _, m := range migrations {
        applied, err := isApplied(ctx, db, m.Version)
        if err != nil {
            return err
        }
        if applied {
            continue
        }

        sqlBytes, err := os.ReadFile(m.Path)
        if err != nil {
            return err
        }

        if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
            if err := tx.Exec(string(sqlBytes)).Error; err != nil {
                return err
            }
            return tx.Exec("INSERT INTO schema_migrations (version) VALUES (?)", m.Version).Error
        }); err != nil {
            return fmt.Errorf("apply migration %s: %w", m.Version, err)
        }
    }

    return nil
}

func ensureMigrationsTable(ctx context.Context, db *gorm.DB) error {
    return db.WithContext(ctx).Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
        version TEXT PRIMARY KEY,
        applied_at TIMESTAMP NOT NULL DEFAULT NOW()
    );`).Error
}

func loadMigrations(dir string) ([]Migration, error) {
    entries, err := os.ReadDir(dir)
    if err != nil {
        if os.IsNotExist(err) {
            return nil, nil
        }
        return nil, err
    }

    var migrations []Migration
    for _, entry := range entries {
        if entry.IsDir() {
            continue
        }
        name := entry.Name()
        if !strings.HasSuffix(name, ".sql") {
            continue
        }
        version := strings.SplitN(name, "_", 2)[0]
        migrations = append(migrations, Migration{
            Version: version,
            Path:    filepath.Join(dir, name),
        })
    }

    sort.Slice(migrations, func(i, j int) bool {
        return migrations[i].Version < migrations[j].Version
    })

    return migrations, nil
}

func isApplied(ctx context.Context, db *gorm.DB, version string) (bool, error) {
    var count int64
    if err := db.WithContext(ctx).Raw("SELECT COUNT(1) FROM schema_migrations WHERE version = ?", version).Scan(&count).Error; err != nil {
        if errorsIsMissingTable(err) {
            return false, nil
        }
        return false, err
    }
    return count > 0, nil
}

func errorsIsMissingTable(err error) bool {
    if err == nil {
        return false
    }
    msg := err.Error()
    return strings.Contains(msg, "schema_migrations") && strings.Contains(msg, "does not exist")
}

var _ fs.DirEntry
