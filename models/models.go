package models

type Room struct {
	Id          int64  `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"type:varchar(100)"`
	Type        string `json:"type" gorm:"type:varchar(20)"`
	Description string `json:"description" gorm:"type:text"`
	UserCount   int    `json:"userCount" gorm:"default:0"`
	CreatedAt   int64  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   int64  `json:"updatedAt" gorm:"autoUpdateTime"`
}

type Message struct {
	Id        int64  `json:"id" gorm:"primaryKey"`
	Type      string `json:"type" gorm:"type:varchar(20)"`
	Content   string `json:"content"`
	RoomId    int64  `json:"roomId"`
	Room      Room   `json:"room" gorm:"foreignKey:RoomId;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	UserId    int64  `json:"userId"`
	CreatedAt int64  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt int64  `json:"updatedAt" gorm:"autoUpdateTime"`
}
