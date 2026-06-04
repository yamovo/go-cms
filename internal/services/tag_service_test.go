package services

import (
	"testing"
)

func TestTagService_Create(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTagService(db)

	req := CreateTagRequest{Name: "Go"}
	tag, err := svc.Create(req)
	if err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	if tag.Name != "Go" {
		t.Errorf("Name = %q, want %q", tag.Name, "Go")
	}
	if tag.Slug != "go" {
		t.Errorf("Slug = %q, want %q", tag.Slug, "go")
	}
}

func TestTagService_Create_Duplicate(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTagService(db)

	svc.Create(CreateTagRequest{Name: "Go"})
	_, err := svc.Create(CreateTagRequest{Name: "Go"})
	if err == nil {
		t.Error("Create() should fail for duplicate tag")
	}
}

func TestTagService_List_WithSearch(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTagService(db)

	svc.Create(CreateTagRequest{Name: "Go"})
	svc.Create(CreateTagRequest{Name: "Golang"})
	svc.Create(CreateTagRequest{Name: "Python"})

	tags, total, err := svc.List(TagListParams{Search: "Go", Sort: "name"})
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}

	if total != 2 {
		t.Errorf("Total = %d, want 2", total)
	}
	if len(tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(tags))
	}
}

func TestTagService_Update(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTagService(db)

	tag, _ := svc.Create(CreateTagRequest{Name: "Old"})

	err := svc.Update(tag.ID, UpdateTagRequest{Name: "New"})
	if err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	got, _ := svc.Get(tag.ID)
	if got.Name != "New" {
		t.Errorf("Name = %q, want %q", got.Name, "New")
	}
}

func TestTagService_Delete(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTagService(db)

	tag, _ := svc.Create(CreateTagRequest{Name: "ToDelete"})

	err := svc.Delete(tag.ID)
	if err != nil {
		t.Fatalf("Delete() error: %v", err)
	}

	_, err = svc.Get(tag.ID)
	if err == nil {
		t.Error("Get() should fail after delete")
	}
}

func TestTagService_Merge(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTagService(db)
	articleSvc := NewArticleService(db, "http://localhost:8080")
	user := createTestUser(t, db, "author1", "author")

	tag1, _ := svc.Create(CreateTagRequest{Name: "Go"})
	tag2, _ := svc.Create(CreateTagRequest{Name: "Golang"})
	tag3, _ := svc.Create(CreateTagRequest{Name: "Target"})

	// Create articles with source tags.
	articleSvc.Create(CreateArticleRequest{Title: "A1", TagIDs: []uint{tag1.ID}, Status: "draft"}, user.ID)
	articleSvc.Create(CreateArticleRequest{Title: "A2", TagIDs: []uint{tag2.ID}, Status: "draft"}, user.ID)

	err := svc.Merge([]uint{tag1.ID, tag2.ID}, tag3.ID, true)
	if err != nil {
		t.Fatalf("Merge() error: %v", err)
	}

	// Source tags should be deleted.
	_, err = svc.Get(tag1.ID)
	if err == nil {
		t.Error("Source tag 1 should be deleted after merge")
	}
	_, err = svc.Get(tag2.ID)
	if err == nil {
		t.Error("Source tag 2 should be deleted after merge")
	}

	// Target should still exist.
	_, err = svc.Get(tag3.ID)
	if err != nil {
		t.Error("Target tag should still exist")
	}
}
