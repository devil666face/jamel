package store

import (
	"time"

	"log"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type database struct {
	db      *gorm.DB
	sqlitec string
	tables  []any
}

func Must(
	_sqlitec string,
	_tables []any,
) *database {
	d := database{
		sqlitec: _sqlitec,
		tables:  _tables,
	}
	if err := d.connect(); err != nil {
		log.Fatalf("database connection error: %s", err)
	}
	if err := d.migrate(); err != nil {
		log.Printf("migrations not create: %s", err)
	}
	return &d
}

func (d *database) DB() *gorm.DB {
	return d.db
}

func (d *database) migrate() error {
	return d.db.AutoMigrate(d.tables...)
}

func (d *database) connect() error {
	return d.sqlite()
}

func (d *database) sqlite() error {
	var err error
	if d.db, err = gorm.Open(sqlite.Open(d.sqlitec+"?cache=shared&mode=rwc&_busy_timeout=50000"), &gorm.Config{
		NowFunc: func() time.Time { return time.Now().Local() },
		// Logger:         logger.Default.LogMode(logger.Info),
		Logger:         logger.Default.LogMode(logger.Silent),
		TranslateError: true,
	}); err != nil {
		return err
	}
	return nil
}
