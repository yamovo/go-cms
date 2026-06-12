package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/vortexcms/go-cms/internal/models"
	"gorm.io/gorm"
)

// ─── Content Type Service ───────────────────────────────────────────────────

// ContentTypeService manages content types and their dynamic entries.
type ContentTypeService struct {
	db *gorm.DB
}

// NewContentTypeService creates a new ContentTypeService.
func NewContentTypeService(db *gorm.DB) *ContentTypeService {
	return &ContentTypeService{db: db}
}

// ─── Content Type CRUD ──────────────────────────────────────────────────────

// CreateContentTypeRequest is the payload for creating a content type.
type CreateContentTypeRequest struct {
	UID          string                   `json:"uid" binding:"required,max=64"`
	Name         string                   `json:"name" binding:"required,max=128"`
	Description  string                   `json:"description"`
	IsSingle     bool                     `json:"is_single"`
	DraftPublish bool                     `json:"draft_publish"`
	Fields       []CreateFieldRequest     `json:"fields" binding:"required,min=1"`
}

// CreateFieldRequest defines a field during content type creation.
type CreateFieldRequest struct {
	Name         string   `json:"name" binding:"required,max=64"`
	Label        string   `json:"label" binding:"required,max=128"`
	FieldType    string   `json:"field_type" binding:"required"`
	Required     bool     `json:"required"`
	Unique       bool     `json:"unique"`
	DefaultValue string   `json:"default_value"`
	Options      []string `json:"options"`       // for enum
	RelationType string   `json:"relation_type"`
	RelationUID  string   `json:"relation_uid"`
	MinLength    *int     `json:"min_length"`
	MaxLength    *int     `json:"max_length"`
	MinValue     *float64 `json:"min_value"`
	MaxValue     *float64 `json:"max_value"`
}

// CreateContentType creates a new content type with fields.
func (s *ContentTypeService) CreateContentType(req CreateContentTypeRequest) (*models.ContentType, error) {
	// Validate UID format (lowercase, underscores only).
	if !isValidUID(req.UID) {
		return nil, errors.New("uid must be lowercase letters, numbers, and underscores only")
	}

	// Check uniqueness.
	var count int64
	s.db.Model(&models.ContentType{}).Where("uid = ?", req.UID).Count(&count)
	if count > 0 {
		return nil, errors.New("content type uid already exists")
	}

	// Validate field types.
	for _, f := range req.Fields {
		if !models.ValidFieldTypes[f.FieldType] {
			return nil, fmt.Errorf("invalid field type: %s", f.FieldType)
		}
		if f.FieldType == models.FieldTypeEnum && len(f.Options) == 0 {
			return nil, fmt.Errorf("enum field %s must have options", f.Name)
		}
	}

	ct := models.ContentType{
		UID:          req.UID,
		Name:         req.Name,
		Description:  req.Description,
		IsSingle:     req.IsSingle,
		DraftPublish: req.DraftPublish,
	}

	for i, f := range req.Fields {
		ct.Fields = append(ct.Fields, models.ContentField{
			Name:         f.Name,
			Label:        f.Label,
			FieldType:    f.FieldType,
			Required:     f.Required,
			Unique:       f.Unique,
			DefaultValue: f.DefaultValue,
			Options:      f.Options,
			RelationType: f.RelationType,
			RelationUID:  f.RelationUID,
			MinLength:    f.MinLength,
			MaxLength:    f.MaxLength,
			MinValue:     f.MinValue,
			MaxValue:     f.MaxValue,
			SortOrder:    i,
		})
	}

	if err := s.db.Create(&ct).Error; err != nil {
		return nil, errors.New("failed to create content type")
	}

	return &ct, nil
}

// ListContentTypes returns all content types with entry counts.
func (s *ContentTypeService) ListContentTypes() ([]models.ContentType, error) {
	var types []models.ContentType
	if err := s.db.Preload("Fields", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Order("created_at ASC").Find(&types).Error; err != nil {
		return nil, err
	}

	// Fill entry counts.
	for i := range types {
		var count int64
		s.db.Model(&models.ContentEntry{}).Where("content_type_id = ?", types[i].ID).Count(&count)
		types[i].EntryCount = count
	}

	return types, nil
}

// GetContentType returns a single content type by UID.
func (s *ContentTypeService) GetContentType(uid string) (*models.ContentType, error) {
	var ct models.ContentType
	if err := s.db.Preload("Fields", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Where("uid = ?", uid).First(&ct).Error; err != nil {
		return nil, errors.New("content type not found")
	}
	return &ct, nil
}

// DeleteContentType deletes a content type and all its entries.
func (s *ContentTypeService) DeleteContentType(uid string) error {
	var ct models.ContentType
	if err := s.db.Where("uid = ?", uid).First(&ct).Error; err != nil {
		return errors.New("content type not found")
	}

	// Delete all entries first.
	s.db.Where("content_type_id = ?", ct.ID).Delete(&models.ContentEntry{})
	// Delete fields.
	s.db.Where("content_type_id = ?", ct.ID).Delete(&models.ContentField{})
	// Delete the type itself.
	return s.db.Delete(&ct).Error
}

// ─── Content Entry CRUD ─────────────────────────────────────────────────────

// CreateEntryRequest is the payload for creating an entry.
type CreateEntryRequest struct {
	Data   map[string]interface{} `json:"data" binding:"required"`
	Status string                 `json:"status"` // draft (default) or published
}

// UpdateEntryRequest is the payload for updating an entry.
type UpdateEntryRequest struct {
	Data   map[string]interface{} `json:"data"`
	Status *string                `json:"status"`
}

// ListEntriesParams holds query parameters for listing entries.
type ListEntriesParams struct {
	Page     int
	PageSize int
	Status   string
	Search   string
	Sort     string
	Filters  map[string]string // field_name=value
}

// ListEntries returns entries of a content type.
func (s *ContentTypeService) ListEntries(uid string, params ListEntriesParams) (interface{}, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return nil, err
	}

	query := s.db.Model(&models.ContentEntry{}).Where("content_type_id = ?", ct.ID)

	// Status filter.
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}

	// JSON field filters.
	for field, value := range params.Filters {
		query = query.Where("json_extract(data, ?) = ?", "$."+field, value)
	}

	// Search in text fields.
	if params.Search != "" {
		query = query.Where("data LIKE ?", "%"+params.Search+"%")
	}

	// Sorting.
	switch params.Sort {
	case "oldest":
		query = query.Order("created_at ASC")
	case "updated":
		query = query.Order("updated_at DESC")
	default:
		query = query.Order("created_at DESC")
	}

	// Pagination.
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 20
	}

	var total int64
	query.Count(&total)

	var entries []models.ContentEntry
	offset := (params.Page - 1) * params.PageSize
	if err := query.Offset(offset).Limit(params.PageSize).Find(&entries).Error; err != nil {
		return nil, err
	}

	return models.NewListResponse(entries, models.Paginate{
		Page:     params.Page,
		PageSize: params.PageSize,
		Total:    total,
	}), nil
}

// GetEntry returns a single entry by document_id.
func (s *ContentTypeService) GetEntry(uid string, documentID string) (*models.ContentEntry, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return nil, err
	}

	var entry models.ContentEntry
	if err := s.db.Where("content_type_id = ? AND document_id = ?", ct.ID, documentID).First(&entry).Error; err != nil {
		return nil, errors.New("entry not found")
	}

	return &entry, nil
}

// CreateEntry creates a new entry for a content type.
func (s *ContentTypeService) CreateEntry(uid string, req CreateEntryRequest, userID uint) (*models.ContentEntry, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return nil, err
	}

	// Validate required fields.
	if err := s.validateEntryData(ct, req.Data); err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		if ct.DraftPublish {
			status = models.EntryStatusDraft
		} else {
			status = models.EntryStatusPublished
		}
	}

	entry := models.ContentEntry{
		ContentTypeID: ct.ID,
		DocumentID:    uuid.New().String(),
		Status:        status,
		Data:          req.Data,
		CreatedByID:   userID,
		UpdatedByID:   userID,
	}

	if status == models.EntryStatusPublished {
		now := time.Now()
		entry.PublishedAt = &now
	}

	if err := s.db.Create(&entry).Error; err != nil {
		return nil, errors.New("failed to create entry")
	}

	return &entry, nil
}

// UpdateEntry updates an existing entry.
func (s *ContentTypeService) UpdateEntry(uid string, documentID string, req UpdateEntryRequest, userID uint) (*models.ContentEntry, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return nil, err
	}

	var entry models.ContentEntry
	if err := s.db.Where("content_type_id = ? AND document_id = ?", ct.ID, documentID).First(&entry).Error; err != nil {
		return nil, errors.New("entry not found")
	}

	// Merge data.
	if req.Data != nil {
		if err := s.validateEntryData(ct, req.Data); err != nil {
			return nil, err
		}
		// Merge with existing data.
		for k, v := range req.Data {
			entry.Data[k] = v
		}
	}

	// Update status.
	if req.Status != nil {
		entry.Status = *req.Status
		if *req.Status == models.EntryStatusPublished && entry.PublishedAt == nil {
			now := time.Now()
			entry.PublishedAt = &now
		}
	}

	entry.UpdatedByID = userID

	if err := s.db.Save(&entry).Error; err != nil {
		return nil, errors.New("failed to update entry")
	}

	return &entry, nil
}

// DeleteEntry deletes an entry by document_id.
func (s *ContentTypeService) DeleteEntry(uid string, documentID string) error {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return err
	}

	result := s.db.Where("content_type_id = ? AND document_id = ?", ct.ID, documentID).Delete(&models.ContentEntry{})
	if result.RowsAffected == 0 {
		return errors.New("entry not found")
	}
	return result.Error
}

// PublishEntry publishes a draft entry.
func (s *ContentTypeService) PublishEntry(uid string, documentID string, userID uint) (*models.ContentEntry, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return nil, err
	}

	var entry models.ContentEntry
	if err := s.db.Where("content_type_id = ? AND document_id = ?", ct.ID, documentID).First(&entry).Error; err != nil {
		return nil, errors.New("entry not found")
	}

	now := time.Now()
	entry.Status = models.EntryStatusPublished
	entry.PublishedAt = &now
	entry.UpdatedByID = userID

	if err := s.db.Save(&entry).Error; err != nil {
		return nil, errors.New("failed to publish entry")
	}

	return &entry, nil
}

// UnpublishEntry reverts a published entry to draft.
func (s *ContentTypeService) UnpublishEntry(uid string, documentID string, userID uint) (*models.ContentEntry, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return nil, err
	}

	var entry models.ContentEntry
	if err := s.db.Where("content_type_id = ? AND document_id = ?", ct.ID, documentID).First(&entry).Error; err != nil {
		return nil, errors.New("entry not found")
	}

	entry.Status = models.EntryStatusDraft
	entry.UpdatedByID = userID

	if err := s.db.Save(&entry).Error; err != nil {
		return nil, errors.New("failed to unpublish entry")
	}

	return &entry, nil
}

// ─── Validation ─────────────────────────────────────────────────────────────

func (s *ContentTypeService) validateEntryData(ct *models.ContentType, data map[string]interface{}) error {
	for _, field := range ct.Fields {
		value, exists := data[field.Name]

		if field.Required && (!exists || value == nil || value == "") {
			return fmt.Errorf("field %s is required", field.Name)
		}

		if !exists || value == nil {
			continue
		}

		// Type validation.
		switch field.FieldType {
		case models.FieldTypeInteger:
			switch v := value.(type) {
			case float64:
				if field.MinValue != nil && v < *field.MinValue {
					return fmt.Errorf("field %s: value must be >= %v", field.Name, *field.MinValue)
				}
				if field.MaxValue != nil && v > *field.MaxValue {
					return fmt.Errorf("field %s: value must be <= %v", field.Name, *field.MaxValue)
				}
			default:
				// Try to convert.
				if _, err := strconv.ParseFloat(fmt.Sprintf("%v", v), 64); err != nil {
					return fmt.Errorf("field %s: must be a number", field.Name)
				}
			}

		case models.FieldTypeFloat:
			if _, ok := value.(float64); !ok {
				if _, err := strconv.ParseFloat(fmt.Sprintf("%v", value), 64); err != nil {
					return fmt.Errorf("field %s: must be a number", field.Name)
				}
			}

		case models.FieldTypeBoolean:
			if _, ok := value.(bool); !ok {
				return fmt.Errorf("field %s: must be a boolean", field.Name)
			}

		case models.FieldTypeEnum:
			strVal := fmt.Sprintf("%v", value)
			found := false
			for _, opt := range field.Options {
				if opt == strVal {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("field %s: invalid enum value %s, allowed: %v", field.Name, strVal, field.Options)
			}

		case models.FieldTypeText, models.FieldTypeRichText:
			strVal := fmt.Sprintf("%v", value)
			if field.MinLength != nil && len(strVal) < *field.MinLength {
				return fmt.Errorf("field %s: minimum length is %d", field.Name, *field.MinLength)
			}
			if field.MaxLength != nil && len(strVal) > *field.MaxLength {
				return fmt.Errorf("field %s: maximum length is %d", field.Name, *field.MaxLength)
			}

		case models.FieldTypeJSON:
			// Validate it's valid JSON by re-marshaling.
			b, err := json.Marshal(value)
			if err != nil {
				return fmt.Errorf("field %s: invalid JSON", field.Name)
			}
			var check interface{}
			if err := json.Unmarshal(b, &check); err != nil {
				return fmt.Errorf("field %s: invalid JSON", field.Name)
			}
		}
	}
	return nil
}

func isValidUID(uid string) bool {
	if len(uid) == 0 || len(uid) > 64 {
		return false
	}
	for _, c := range uid {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}
	return true
}

// GetEntriesByUID returns entries for a content type by UID (for relation loading).
func (s *ContentTypeService) GetEntriesByUID(uid string, ids []uint) ([]models.ContentEntry, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return nil, err
	}
	var entries []models.ContentEntry
	s.db.Where("content_type_id = ? AND id IN ?", ct.ID, ids).Find(&entries)
	return entries, nil
}

// SearchEntries searches across all text fields of a content type.
func (s *ContentTypeService) SearchEntries(uid string, query string, limit int) ([]models.ContentEntry, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}

	var entries []models.ContentEntry
	s.db.Where("content_type_id = ? AND data LIKE ?", ct.ID, "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).
		Find(&entries)

	return entries, nil
}

// ExportEntries exports all entries of a content type as JSON.
func (s *ContentTypeService) ExportEntries(uid string) (string, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return "", err
	}

	var entries []models.ContentEntry
	s.db.Where("content_type_id = ?", ct.ID).Order("created_at ASC").Find(&entries)

	b, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// ImportEntries imports entries from JSON.
func (s *ContentTypeService) ImportEntries(uid string, data string, userID uint) (int, error) {
	ct, err := s.GetContentType(uid)
	if err != nil {
		return 0, err
	}

	var entries []models.ContentEntry
	if err := json.Unmarshal([]byte(data), &entries); err != nil {
		return 0, errors.New("invalid JSON data")
	}

	count := 0
	for _, entry := range entries {
		entry.ID = 0 // reset ID
		entry.ContentTypeID = ct.ID
		entry.DocumentID = uuid.New().String()
		entry.CreatedByID = userID
		entry.UpdatedByID = userID
		if err := s.db.Create(&entry).Error; err == nil {
			count++
		}
	}

	return count, nil
}
