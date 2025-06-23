package models

import (
	"time"

	"gorm.io/gorm"

	H "mercadillo-global/helpers"
)

type Question struct {
	ID           string    `json:"id" gorm:"type:char(36);primaryKey"`
	ProductID    string    `json:"product_id" gorm:"type:char(36);not null;index"`
	Question     string    `json:"question" gorm:"type:text;not null"`
	Answer       string    `json:"answer" gorm:"type:text"`
	AnsweredByIA bool      `json:"answered_by_ia" gorm:"default:false"`
	Helpful      int       `json:"helpful" gorm:"default:0"`
	Status       string    `json:"status" gorm:"type:enum('wait_for_ia','wait_for_human_review','hidden','answered');default:'wait_for_ia'"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Relations
	Product       Product        `json:"product" gorm:"foreignKey:ProductID"`
	QuestionVotes []QuestionVote `json:"question_votes" gorm:"foreignKey:QuestionID"`
}

type Review struct {
	ID        string    `json:"id" gorm:"type:char(36);primaryKey"`
	ProductID string    `json:"product_id" gorm:"type:char(36);not null;index"`
	Name      string    `json:"name" gorm:"type:varchar(255);not null"`
	Rating    int       `json:"rating" gorm:"type:tinyint;not null"`
	Comment   string    `json:"comment" gorm:"type:text"`
	Helpful   int       `json:"helpful" gorm:"default:0"`
	Status    string    `json:"status" gorm:"type:enum('approved','wait_for_ia','wait_for_human_review','hidden');default:'wait_for_ia'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Product     Product      `json:"product" gorm:"foreignKey:ProductID"`
	ReviewVotes []ReviewVote `json:"review_votes" gorm:"foreignKey:ReviewID"`
}

type QuestionVote struct {
	ID         string    `json:"id" gorm:"type:char(36);primaryKey"`
	UserID     string    `json:"user_id" gorm:"type:char(36);not null;index"`
	QuestionID string    `json:"question_id" gorm:"type:char(36);not null;index"`
	Vote       int       `json:"vote" gorm:"type:tinyint;not null;comment:'1 for helpful, -1 for not helpful'"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Relations
	User     User     `json:"user" gorm:"foreignKey:UserID"`
	Question Question `json:"question" gorm:"foreignKey:QuestionID"`
}

type ReviewVote struct {
	ID        string    `json:"id" gorm:"type:char(36);primaryKey"`
	UserID    string    `json:"user_id" gorm:"type:char(36);not null;index"`
	ReviewID  string    `json:"review_id" gorm:"type:char(36);not null;index"`
	Vote      int       `json:"vote" gorm:"type:tinyint;not null;comment:'1 for helpful, -1 for not helpful'"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	User   User   `json:"user" gorm:"foreignKey:UserID"`
	Review Review `json:"review" gorm:"foreignKey:ReviewID"`
}

// GORM Hooks
func (q *Question) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(q.ID) {
		q.ID = H.NewUUID()
	}
	return nil
}

func (r *Review) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(r.ID) {
		r.ID = H.NewUUID()
	}
	return nil
}

func (qv *QuestionVote) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(qv.ID) {
		qv.ID = H.NewUUID()
	}
	return nil
}

func (rv *ReviewVote) BeforeCreate(tx *gorm.DB) error {
	if H.IsEmpty(rv.ID) {
		rv.ID = H.NewUUID()
	}
	return nil
}
