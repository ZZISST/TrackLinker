package db

import (
	"context"
	"fmt"
	"time"

	ports "gitlab.com/etherealSent/tracklinker-backend/collection/internal/ports"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type Adapter struct {
	db *gorm.DB
}

const DatabaseTimeout = 5 * time.Second

func NewAdapter(dsn string) (*Adapter, error) {
	db, openErr := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})

	if openErr != nil {
		return nil, fmt.Errorf("db connection error: %v", openErr)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(200)

	return &Adapter{db: db}, nil
}

func (d *Adapter) AutoMigrate(models ...any) error {
	return d.db.AutoMigrate(models...)
}

func (d *Adapter) WithTransaction(function func() error) error {
	callback := func(db *gorm.DB) error {
		return function()
	}

	tx := d.db.Begin()
	if err := callback(tx); err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (d *Adapter) Preload(relation string, conditions ...interface{}) *Adapter {
	if len(conditions) > 0 {
		// If conditions are provided, pass them to the Preload method
		d.db = d.db.Preload(relation, conditions...)
	} else {
		// If no conditions are provided, preload the relation without conditions
		d.db = d.db.Preload(relation)
	}
	return d
}

func (d *Adapter) Create(ctx context.Context, doc any) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	return d.db.Create(doc).Error
}

func (d *Adapter) CreateInBatches(ctx context.Context, docs any, batchSize int) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	return d.db.CreateInBatches(docs, batchSize).Error
}

func (d *Adapter) Update(ctx context.Context, doc any) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	return d.db.Save(doc).Error
}

func (d *Adapter) Delete(ctx context.Context, value any, opts ...ports.FindOption) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	query := d.applyOptions(opts...)
	return query.Delete(value).Error
}

func (d *Adapter) FindById(ctx context.Context, id string, result any) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	if err := d.db.Where("id = ? ", id).First(result).Error; err != nil {
		return err
	}

	return nil
}

func (d *Adapter) FindOne(ctx context.Context, result any, opts ...ports.FindOption) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	option := d.applyOptions(opts...)

	if err := option.First(result).Error; err != nil {
		return err
	}

	return nil
}

func (d *Adapter) Joins(query string, args ...any) ports.DBPort {
	d.db = d.db.Joins(query, args...)
	return d
}

func (d *Adapter) Find(ctx context.Context, result any, opts ...ports.FindOption) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	query := d.applyOptions(opts...)
	if err := query.Find(result).Error; err != nil {
		return err
	}

	return nil
}

func (d *Adapter) Count(ctx context.Context, model any, total *int64, opts ...ports.FindOption) error {
	ctx, cancel := context.WithTimeout(ctx, DatabaseTimeout)
	defer cancel()

	query := d.applyOptions(opts...)
	if err := query.Model(model).Count(total).Error; err != nil {
		return err
	}

	return nil
}

func (d *Adapter) GetDB() *gorm.DB {
	return d.db
}

func (d *Adapter) applyOptions(opts ...ports.FindOption) *gorm.DB {
	query := d.db

	opt := getOption(opts...)

	if len(opt.Preloads) != 0 {
		for _, preload := range opt.Preloads {
			query = query.Preload(preload)
		}
	}

	if opt.Query != nil {
		for _, q := range opt.Query {
			query = query.Where(q.Query, q.Args...)
		}
	}

	if opt.Order != "" {
		query = query.Order(opt.Order)
	}

	if opt.Offset != 0 {
		query = query.Offset(opt.Offset)
	}

	if opt.Limit != 0 {
		query = query.Limit(opt.Limit)
	}

	if opt.Join != "" {
		query = query.Joins(opt.Join)
	}

	return query
}

func getOption(opts ...ports.FindOption) ports.Option {
	opt := ports.Option{
		Query:  []ports.Query{},
		Offset: 0,
		Limit:  1000,
		Order:  "id",
	}

	for _, o := range opts {
		o.Apply(&opt)
	}

	return opt
}
