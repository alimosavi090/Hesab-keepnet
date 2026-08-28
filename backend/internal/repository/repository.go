package repository

import (
	"context"

	"gorm.io/gorm"
)

type Repository[T any] struct {
	DB *gorm.DB
}

func New[T any](db *gorm.DB) Repository[T] {
	return Repository[T]{DB: db}
}

func (r Repository[T]) Create(ctx context.Context, tx *gorm.DB, entity *T) error {
	db := r.pick(tx).WithContext(ctx)
	return db.Create(entity).Error
}

func (r Repository[T]) FindByID(ctx context.Context, id int64) (*T, error) {
	var entity T
	if err := r.DB.WithContext(ctx).First(&entity, id).Error; err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r Repository[T]) List(ctx context.Context) ([]T, error) {
	var entities []T
	if err := r.DB.WithContext(ctx).Order("id").Find(&entities).Error; err != nil {
		return nil, err
	}
	return entities, nil
}

func (r Repository[T]) SoftDelete(ctx context.Context, tx *gorm.DB, entity *T) error {
	db := r.pick(tx).WithContext(ctx)
	return db.Delete(entity).Error
}

func (r Repository[T]) Count(ctx context.Context, query func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	db := query(r.DB.WithContext(ctx).Model(new(T)))
	if err := db.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r Repository[T]) pick(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.DB
}
