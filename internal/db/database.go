package db

import (
	"fmt"
	"time"

	"github.com/xiaohongshu-image/internal/config"
	"github.com/xiaohongshu-image/internal/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func New(cfg *config.DatabaseConfig) (*Database, error) {
	// 使用 Asia/Shanghai 时区 (GMT+8)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FShanghai",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	if err := db.AutoMigrate(
		&models.Setting{},
		&models.Note{},
		&models.Comment{},
		&models.Task{},
		&models.Delivery{},
		&models.AuditLog{},
	); err != nil {
		return nil, fmt.Errorf("failed to auto migrate: %w", err)
	}

	return &Database{DB: db}, nil
}

func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (d *Database) GetSetting() (*models.Setting, error) {
	var setting models.Setting
	err := d.DB.First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (d *Database) UpdateSetting(setting *models.Setting) error {
	return d.DB.Save(setting).Error
}

func (d *Database) GetOrCreateNote(noteTarget string) (*models.Note, error) {
	var note models.Note
	err := d.DB.Where("note_target = ?", noteTarget).FirstOrCreate(&note, models.Note{
		NoteTarget: noteTarget,
	}).Error
	if err != nil {
		return nil, err
	}
	return &note, nil
}

func (d *Database) UpdateNote(note *models.Note) error {
	return d.DB.Save(note).Error
}

func (d *Database) CreateComment(comment *models.Comment) error {
	return d.DB.Create(comment).Error
}

func (d *Database) CommentExists(commentUID string) (bool, error) {
	var count int64
	err := d.DB.Model(&models.Comment{}).Where("comment_uid = ?", commentUID).Count(&count).Error
	return count > 0, err
}

func (d *Database) GetCommentByUID(commentUID string) (*models.Comment, error) {
	var comment models.Comment
	err := d.DB.Where("comment_uid = ?", commentUID).First(&comment).Error
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (d *Database) CreateTask(task *models.Task) error {
	return d.DB.Create(task).Error
}

func (d *Database) GetTaskByID(id uint) (*models.Task, error) {
	var task models.Task
	err := d.DB.Preload("Comment").Preload("Deliveries").Where("id = ?", id).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (d *Database) GetTaskByCommentID(commentID uint) (*models.Task, error) {
	var task models.Task
	err := d.DB.Where("comment_id = ?", commentID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (d *Database) UpdateTask(task *models.Task) error {
	return d.DB.Save(task).Error
}

func (d *Database) ListTasks(limit int, offset int) ([]models.Task, error) {
	var tasks []models.Task
	err := d.DB.Preload("Comment").Order("created_at DESC").Limit(limit).Offset(offset).Find(&tasks).Error
	return tasks, err
}

func (d *Database) CreateDelivery(delivery *models.Delivery) error {
	return d.DB.Create(delivery).Error
}

func (d *Database) CreateAuditLog(log *models.AuditLog) error {
	return d.DB.Create(log).Error
}
