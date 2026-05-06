package post

import (
	"welog/internal/model"

	"gorm.io/gorm"
)

type PostRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) Create(post *model.Post) error {
	return r.db.Create(post).Error
}

func (r *PostRepository) FindAll(postType uint, offset, limit int) ([]model.Post, error) {
	var posts []model.Post
	query := r.db.Model(&model.Post{})

	// 1이면 개인, 2이면 공용 게시판 Type
	if postType != 0 {
		query = query.Where("type = ?", postType)
	}

	err := query.Order("created_at desc").Offset(offset).Limit(limit).Find(&posts).Error
	return posts, err
}

func (r *PostRepository) FindByID(postID uint) (*model.Post, error) {
	var post model.Post
	err := r.db.Preload("Comments").First(&post, postID).Error
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *PostRepository) Delete(postID uint) error {
	return r.db.Delete(&model.Post{}, postID).Error
}
