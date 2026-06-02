package post

import (
	"welog/internal/model"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{
		db: db,
	}
}

func (r *PostRepository) Create(post *model.Post) error {
	return r.db.Create(post).Error
}

func (r *PostRepository) FindAllPublic(offset, limit int) ([]model.Post, error) {
	var posts []model.Post
	query := r.db.Model(&model.Post{}).
		Where("type = ?", model.PostTypePublic).
		Select("posts.*, (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id AND comments.deleted_at IS NULL) AS comment_count")

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

func (r *PostRepository) FindAllPrivate(userID uint, offset, limit int) ([]model.Post, error) {
	var posts []model.Post
	query := r.db.Model(&model.Post{}).
		Where("type = ? AND user_id = ?", model.PostTypePrivate, userID).
		Select("posts.*, (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id AND comments.deleted_at IS NULL) AS comment_count")

	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

func (r *PostRepository) FindByID(postID uint) (*model.Post, error) {
	var post model.Post
	err := r.db.Model(&model.Post{}).
		Select("posts.*, (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id AND comments.deleted_at IS NULL) AS comment_count").
		Preload("Comments", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).Preload("Comments.User").First(&post, postID).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) FindAllByUserID(userID uint) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.Model(&model.Post{}).
		Select("posts.*, (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id AND comments.deleted_at IS NULL) AS comment_count").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&posts).Error
	return posts, err
}

func (r *PostRepository) Delete(postID uint) error {
	return r.db.Delete(&model.Post{}, postID).Error
}
