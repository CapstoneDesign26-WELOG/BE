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

func (r *CommentRepository) FindByID(commentID uint) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.Preload("User").First(&comment, commentID).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepository) Delete(commentID uint) error {
	return r.db.Delete(&model.Comment{}, commentID).Error
}

func (r *CommentRepository) UpsertPreference(userID, aiType uint, score int) error {
	var pref model.UserPreference
	err := r.db.Where("user_id = ? AND ai_type = ?", userID, aiType).First(&pref).Error

	if err == gorm.ErrRecordNotFound {
		pref = model.UserPreference{
			UserID: userID,
			AIType: aiType,
			Score:  score,
		}
		return r.db.Create(&pref).Error
	} else if err != nil {
		return err
	}

	return r.db.Model(&pref).Update("score", gorm.Expr("score + ?", score)).Error
}

func (r *CommentRepository) FindAllByUserIDWithPost(userID uint) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.Preload("Post").Preload("Post.Comments").Where("user_id = ?", userID).Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) IncrementLikeCount(commentID uint) error {
	return r.db.Model(&model.Comment{}).
		Where("id = ?", commentID).
		Update("like_count", gorm.Expr("like_count + ?", 1)).Error
}
