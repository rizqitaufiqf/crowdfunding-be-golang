package campaign

import "gorm.io/gorm"

type Repository interface {
	FindAll() ([]Campaign, error)
	FindAllByID(UserId string) ([]Campaign, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *repository {
	return &repository{db}
}

func (r *repository) FindAll() ([]Campaign, error) {
	var campaigns []Campaign
	// Preload "CampaignImages" association with the condition "campaign_images.is_primary = 1"
	// This ensures that related CampaignImages are eagerly loaded along with the Campaigns during the Find operation.
	err := r.db.Preload("CampaignImages", "campaign_images.is_primary = 1").Find(&campaigns).Error
	if err != nil {
		return campaigns, err
	}

	return campaigns, nil
}

func (r *repository) FindAllByID(userID string) ([]Campaign, error) {
	var campaigns []Campaign
	err := r.db.Preload("CampaignImages", "campaign_images.is_primary = 1").Where("user_id = ?", userID).Find(&campaigns).Error
	if err != nil {
		return campaigns, err
	}

	return campaigns, nil
}
