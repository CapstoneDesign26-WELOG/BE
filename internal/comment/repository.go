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

func (r *CommentRepository) Transaction(fn func(repo *CommentRepository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&CommentRepository{
			db: tx,
		})
	})
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
	err := r.db.Preload("Post", func(db *gorm.DB) *gorm.DB {
		return db.Select("posts.*, (SELECT COUNT(*) FROM comments WHERE comments.post_id = posts.id AND comments.deleted_at IS NULL) AS comment_count")
	}).Preload("Post.Comments", func(db *gorm.DB) *gorm.DB {
		return db.Order("created_at ASC")
	}).Where("user_id = ?", userID).Find(&comments).Error
	return comments, err
}

func (r *CommentRepository) GetLike(userID, commentID uint) (*model.CommentLike, error) {
	var like model.CommentLike
	err := r.db.Where("user_id = ? AND comment_id = ?", userID, commentID).First(&like).Error
	if err != nil {
		return nil, err
	}
	return &like, nil
}

func (r *CommentRepository) GetLikedCommentsIDs(userID uint, commentsIDs []uint) (map[uint]bool, error) {
	var likedIDs []uint
	err := r.db.Model(&model.CommentLike{}).
		Where("user_id = ? AND comment_id IN ?", userID, commentsIDs).
		Pluck("comment_id", &likedIDs).Error

	likedMap := make(map[uint]bool)
	for _, id := range likedIDs {
		likedMap[id] = true
	}
	return likedMap, err
}

func (r *CommentRepository) CreateLike(like *model.CommentLike) error {
	return r.db.Create(like).Error
}

func (r *CommentRepository) DeleteLike(userID, commentID uint) error {
	return r.db.Where("user_id = ? AND comment_id = ?", userID, commentID).Delete(&model.CommentLike{}).Error
}

func (r *CommentRepository) UpdateLikeCount(commentID uint, delta int) error {
	return r.db.Model(&model.Comment{}).Where("id = ?", commentID).
		Update("like_count", gorm.Expr("like_count + ?", delta)).Error
}
