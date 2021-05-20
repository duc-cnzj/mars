package models

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

//$table->id();
//$table->string('name', 50);
//$table->unsignedBigInteger('user_id')->index();
//$table->string('user_name');
//$table->json('image_pull_secrets')->nullable()->comment("数组");
//$table->timestamps();

type Namespace struct {
	ID int `json:"id" gorm:"primaryKey;"`

	Name string `json:"name" gorm:"size:100;not null;unique;comment:'项目空间名'"`

	// ImagePullSecrets
	ImagePullSecrets string `json:"image_pull_secrets" gorm:"size:255;not null;default:'';comment:'项目空间拉取镜像的secrets，数组'"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`

	Projects []Project
}

func (ns *Namespace) ImagePullSecretsArray() []string {
	return strings.Split(ns.ImagePullSecrets, ",")
}
