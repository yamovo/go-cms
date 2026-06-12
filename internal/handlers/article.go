package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vortexcms/go-cms/internal/middleware"
	"github.com/vortexcms/go-cms/internal/models"
	"github.com/vortexcms/go-cms/internal/services"
	"gorm.io/gorm"
)

// ArticleHandler handles article-related HTTP requests.
type ArticleHandler struct {
	svc *services.ArticleService
}

// NewArticleHandler creates a new article handler.
func NewArticleHandler(svc *services.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

// List returns a paginated list of articles.
// GET /api/v1/articles?page=1&page_size=20&status=published&category_id=1&tag=go&sort=newest
func (h *ArticleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	filter := services.ListArticlesFilter{
		Page:       page,
		PageSize:   pageSize,
		Status:     c.Query("status"),
		PostType:   c.Query("post_type"),
		CategoryID: c.Query("category_id"),
		TagSlug:    c.Query("tag"),
		Search:     c.Query("search"),
		Sort:       c.DefaultQuery("sort", "newest"),
		AuthorID:   c.Query("author_id"),
	}

	result, err := h.svc.List(filter)
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, result)
}

// Get returns a single article by ID.
// GET /api/v1/articles/:id
func (h *ArticleHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "Invalid article ID")
		return
	}

	article, err := h.svc.Get(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Article not found")
			return
		}
		InternalError(c)
		return
	}

	Success(c, article)
}

// GetBySlug returns a single article by slug (public endpoint).
// GET /api/v1/articles/slug/:slug
func (h *ArticleHandler) GetBySlug(c *gin.Context) {
	article, err := h.svc.GetBySlug(c.Param("slug"))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			NotFound(c, "Article not found")
			return
		}
		InternalError(c)
		return
	}

	Success(c, article)
}

// Create creates a new article.
// POST /api/v1/articles
func (h *ArticleHandler) Create(c *gin.Context) {
	var req services.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	user := getCurrentUser(c)
	if user == nil {
		return
	}

	article, err := h.svc.Create(req, user.ID)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	Created(c, article)
}

// Update updates an existing article.
// PUT /api/v1/articles/:id
func (h *ArticleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "Invalid article ID")
		return
	}

	var req services.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	user := getCurrentUser(c)
	if user == nil {
		return
	}

	article, err := h.svc.Update(uint(id), req, user.ID, user.IsEditor())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	Success(c, article)
}

// Delete soft-deletes an article.
// DELETE /api/v1/articles/:id
func (h *ArticleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "Invalid article ID")
		return
	}

	user := getCurrentUser(c)
	if user == nil {
		return
	}

	if err := h.svc.Delete(uint(id), user.ID, user.IsEditor()); err != nil {
		handleServiceError(c, err)
		return
	}

	Success(c, gin.H{"message": "Article deleted successfully"})
}

// BulkAction handles bulk operations on articles.
// POST /api/v1/articles/bulk
func (h *ArticleHandler) BulkAction(c *gin.Context) {
	var req services.BulkActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	user := getCurrentUser(c)
	if user == nil {
		return
	}
	if !user.IsEditor() {
		Forbidden(c, "Insufficient permissions")
		return
	}

	affected, err := h.svc.BulkAction(req)
	if err != nil {
		BadRequest(c, err.Error())
		return
	}

	Success(c, gin.H{
		"message":  "Bulk action completed",
		"action":   req.Action,
		"affected": affected,
	})
}

// Revisions returns the revision history for an article.
// GET /api/v1/articles/:id/revisions
func (h *ArticleHandler) Revisions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "Invalid article ID")
		return
	}

	revisions, err := h.svc.Revisions(uint(id))
	if err != nil {
		InternalError(c)
		return
	}

	Success(c, revisions)
}

// RestoreRevision restores an article to a specific revision.
// POST /api/v1/articles/:id/revisions/:revision_id/restore
func (h *ArticleHandler) RestoreRevision(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "Invalid article ID")
		return
	}
	revisionID, err := strconv.ParseUint(c.Param("revision_id"), 10, 64)
	if err != nil {
		BadRequest(c, "Invalid revision ID")
		return
	}

	user := getCurrentUser(c)
	if user == nil {
		return
	}

	if err := h.svc.RestoreRevision(uint(id), uint(revisionID), user.ID); err != nil {
		handleServiceError(c, err)
		return
	}

	Success(c, gin.H{"message": "Revision restored successfully"})
}

// LikeArticle increments the like count.
// POST /api/v1/articles/:id/like
func (h *ArticleHandler) LikeArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		BadRequest(c, "Invalid article ID")
		return
	}

	if err := h.svc.LikeArticle(uint(id)); err != nil {
		InternalError(c)
		return
	}

	Success(c, gin.H{"message": "Article liked"})
}

// Feed returns articles as RSS/XML.
// GET /api/v1/feed
func (h *ArticleHandler) Feed(c *gin.Context) {
	xml, err := h.svc.GenerateFeed()
	if err != nil {
		InternalError(c)
		return
	}

	c.Data(200, "application/rss+xml; charset=utf-8", []byte(xml))
}


// getCurrentUser returns the authenticated user or sends a 401 response.
func getCurrentUser(c *gin.Context) *models.User {
	user := middleware.GetCurrentUser(c)
	if user == nil {
		Unauthorized(c, "Not authenticated")
	}
	return user
}
