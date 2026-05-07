package comment

import (
	"welog/internal/model"

	"gorm.io/gorm"
)

type CommentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) *CommentRepository {
	return &CommentRepository{
		db: db,
	}
}

func (r *CommentRepository) Create(comment *model.Comment) error {
	return r.db.Create(comment).Error
}

func (r *CommentRepository) FindAllByPostID(postID uint) ([]model.Comment, error) {
	var comments []model.Comment
	query := r.db.Model(&model.Comment{}).Where("post_id = ?", postID)

	err := query.Order("created_at desc").Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) FindAllByUserID(userID uint) ([]model.Comment, error) {
	var comments []model.Comment
	query := r.db.Model(&model.Comment{}).Where("user_id = ?", userID)

	err := query.Order("created_at desc").Find(&comments).Error
	return comments, err
}
