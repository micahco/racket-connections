package models

import "github.com/jackc/pgx/v5/pgxpool"

type Models struct {
	Post         *PostModel
	Skill        *SkillLevelModel
	Sport        *SportModel
	User         *UserModel
	Contact      *ContactModel
	Timeslot     *TimeslotModel
	Verification *VerificationModel
}

func New(pool *pgxpool.Pool) Models {
	return Models{
		Post:         &PostModel{pool},
		Skill:        &SkillLevelModel{pool},
		Sport:        &SportModel{pool},
		User:         &UserModel{pool},
		Contact:      &ContactModel{pool},
		Timeslot:     &TimeslotModel{pool},
		Verification: &VerificationModel{pool},
	}
}
