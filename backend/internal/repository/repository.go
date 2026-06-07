package repository

import (
	"gorm.io/gorm"
)

type Repositories struct {
	User      UserRepository
	Event     EventRepository
	News      NewsRepository
	Resource  ResourceRepository
	Showcase  ShowcaseRepository
	Favorite  FavoriteRepository
	Admin     AdminRepository
	Comment   CommentRepository
	Download  DownloadRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User:     NewUserRepository(db),
		Event:    NewEventRepository(db),
		News:     NewNewsRepository(db),
		Resource: NewResourceRepository(db),
		Showcase: NewShowcaseRepository(db),
		Favorite: NewFavoriteRepository(db),
		Admin:    NewAdminRepository(db),
		Comment:  NewCommentRepository(db),
		Download: NewDownloadRepository(db),
	}
}