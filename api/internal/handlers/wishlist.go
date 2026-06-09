package handlers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/boxingoctopus/snackmates/api/internal/auth"
	"github.com/boxingoctopus/snackmates/api/internal/friends"
	"github.com/boxingoctopus/snackmates/api/internal/middleware"
	"github.com/boxingoctopus/snackmates/api/internal/models"
	"github.com/boxingoctopus/snackmates/api/internal/search"
	"github.com/boxingoctopus/snackmates/api/internal/snacksearch"
	"github.com/boxingoctopus/snackmates/api/internal/storage"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WishlistHandler struct {
	pool    *pgxpool.Pool
	search  *search.Client
	storage *storage.Client
}

func NewWishlistHandler(pool *pgxpool.Pool, s *search.Client, storage *storage.Client) *WishlistHandler {
	return &WishlistHandler{pool: pool, search: s, storage: storage}
}

func (h *WishlistHandler) RegisterRoutes(app fiber.Router) {
	w := app.Group("/wishlists")
	w.Get("/friends", middleware.RequireAuth(h.pool), h.ListFriends)
	w.Get("/", middleware.RequireAuth(h.pool), h.ListMine)
	w.Post("/", middleware.RequireAuth(h.pool), h.Create)
	w.Get("/:slug", h.Get)
	w.Put("/:slug", middleware.RequireAuth(h.pool), h.Update)
	w.Delete("/:slug", middleware.RequireAuth(h.pool), h.Delete)
	w.Post("/:slug/banner", middleware.RequireAuth(h.pool), h.UploadBanner)
	w.Put("/:slug/banner", middleware.RequireAuth(h.pool), h.SetBannerURL)
	w.Post("/:slug/items", middleware.RequireAuth(h.pool), h.AddItem)
	w.Put("/:slug/items/:itemId", middleware.RequireAuth(h.pool), h.UpdateItem)
	w.Delete("/:slug/items/:itemId", middleware.RequireAuth(h.pool), h.DeleteItem)
	w.Post("/:slug/items/:itemId/snag", middleware.RequireAuth(h.pool), h.SnagItem)
}

func populateWishlistBanner(ctx context.Context, s *storage.Client, w *models.Wishlist, bannerKey, bannerURL *string) {
	w.BannerURL = resolvePublicBannerURL(ctx, s, bannerURL, bannerKey)
}

func (h *WishlistHandler) populateBanner(ctx context.Context, w *models.Wishlist, bannerKey, bannerURL *string) {
	populateWishlistBanner(ctx, h.storage, w, bannerKey, bannerURL)
}

func (h *WishlistHandler) ListMine(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	rows, err := h.pool.Query(c.Context(), `
		SELECT w.id, w.user_id, w.slug, w.title, w.description, w.is_public, w.banner_key, w.banner_url,
		       w.created_at, w.updated_at,
		       (SELECT COUNT(*) FROM wishlist_items wi WHERE wi.wishlist_id = w.id)
		FROM wishlists w WHERE w.user_id = $1 ORDER BY w.updated_at DESC
	`, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	lists := make([]models.Wishlist, 0)
	for rows.Next() {
		var w models.Wishlist
		var bannerKey, bannerURL *string
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.Slug, &w.Title, &w.Description, &w.IsPublic, &bannerKey, &bannerURL,
			&w.CreatedAt, &w.UpdatedAt, &w.ItemCount,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		h.populateBanner(c.Context(), &w, bannerKey, bannerURL)
		lists = append(lists, w)
	}
	return c.JSON(lists)
}

func (h *WishlistHandler) ListFriends(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	rows, err := h.pool.Query(c.Context(), `
		SELECT w.id, w.user_id, w.slug, w.title, w.description, w.is_public, w.banner_key, w.banner_url,
		       w.created_at, w.updated_at,
		       (SELECT COUNT(*)::int FROM wishlist_items wi WHERE wi.wishlist_id = w.id),
		       u.id, u.username, u.display_name, COALESCE(u.bio, ''), COALESCE(u.country, ''),
		       u.avatar_key, u.avatar_url
		FROM wishlists w
		JOIN friendships f ON f.status = 'accepted'
		  AND (($1 = f.requester_id AND f.addressee_id = w.user_id)
		    OR ($1 = f.addressee_id AND f.requester_id = w.user_id))
		JOIN users u ON u.id = w.user_id
		WHERE w.is_public = TRUE
		  AND w.user_id <> $1
		  AND u.email_verified = TRUE
		  AND u.deactivated_at IS NULL
		ORDER BY w.updated_at DESC
		LIMIT 50
	`, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	lists := make([]models.FriendWishlist, 0)
	for rows.Next() {
		var fw models.FriendWishlist
		var bannerKey, bannerURL *string
		var ownerAvatarKey, ownerAvatarURL *string
		if err := rows.Scan(
			&fw.ID, &fw.UserID, &fw.Slug, &fw.Title, &fw.Description, &fw.IsPublic, &bannerKey, &bannerURL,
			&fw.CreatedAt, &fw.UpdatedAt, &fw.ItemCount,
			&fw.Owner.ID, &fw.Owner.Username, &fw.Owner.DisplayName, &fw.Owner.Bio, &fw.Owner.Country,
			&ownerAvatarKey, &ownerAvatarURL,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		h.populateBanner(c.Context(), &fw.Wishlist, bannerKey, bannerURL)
		if h.storage != nil {
			if avatarURL, err := h.storage.ResolveObjectURL(c.Context(), ownerAvatarURL, ownerAvatarKey); err == nil && avatarURL != "" {
				fw.Owner.AvatarURL = &avatarURL
			}
		} else if ownerAvatarURL != nil && *ownerAvatarURL != "" {
			fw.Owner.AvatarURL = ownerAvatarURL
		}
		lists = append(lists, fw)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(lists)
}

type createWishlistRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

func (h *WishlistHandler) Create(c *fiber.Ctx) error {
	var req createWishlistRequest
	if err := c.BodyParser(&req); err != nil || req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title required"})
	}
	wishlistSlug, err := allocateWishlistSlug(c.Context(), h.pool, req.Title, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	var w models.Wishlist
	err = h.pool.QueryRow(c.Context(), `
		INSERT INTO wishlists (user_id, slug, title, description, is_public)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, slug, title, description, is_public, created_at, updated_at
	`, middleware.GetUserID(c), wishlistSlug, req.Title, req.Description, req.IsPublic).Scan(
		&w.ID, &w.UserID, &w.Slug, &w.Title, &w.Description, &w.IsPublic, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(w)
}

func (h *WishlistHandler) Get(c *fiber.Ctx) error {
	id, err := lookupWishlistID(c.Context(), h.pool, c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	var w models.Wishlist
	var bannerKey, bannerURL *string
	err = h.pool.QueryRow(c.Context(), `
		SELECT id, user_id, slug, title, description, is_public, banner_key, banner_url, created_at, updated_at,
		       (SELECT COUNT(*) FROM wishlist_items wi WHERE wi.wishlist_id = wishlists.id)
		FROM wishlists WHERE id = $1 AND (is_public = TRUE OR user_id = $2)
	`, id, middleware.GetUserID(c)).Scan(
		&w.ID, &w.UserID, &w.Slug, &w.Title, &w.Description, &w.IsPublic, &bannerKey, &bannerURL,
		&w.CreatedAt, &w.UpdatedAt, &w.ItemCount,
	)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	h.populateBanner(c.Context(), &w, bannerKey, bannerURL)

	viewerID := middleware.GetUserID(c)
	viewerCanSnag := false
	if viewerID != uuid.Nil && viewerID != w.UserID {
		viewerCanSnag = hasActiveMatchWith(c.Context(), h.pool, viewerID, w.UserID)
	}

	rows, err := h.pool.Query(c.Context(), `
		SELECT wi.id, wi.wishlist_id, wi.name, wi.type, wi.brand, wi.notes, wi.image_url,
		       wi.created_at, wi.updated_at,
		       u.id, u.display_name, s.delivery_method, s.tracking_number
		FROM wishlist_items wi
		LEFT JOIN wishlist_item_snags s ON s.wishlist_item_id = wi.id
		LEFT JOIN users u ON u.id = s.snagged_by_user_id
		WHERE wi.wishlist_id = $1
		ORDER BY wi.created_at ASC
	`, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	items := []models.WishlistItem{}
	for rows.Next() {
		var item models.WishlistItem
		var snaggerID *uuid.UUID
		var snaggerName *string
		var deliveryMethod *string
		var trackingNumber *string
		if err := rows.Scan(
			&item.ID, &item.WishlistID, &item.Name, &item.Type, &item.Brand, &item.Notes, &item.ImageURL,
			&item.CreatedAt, &item.UpdatedAt,
			&snaggerID, &snaggerName, &deliveryMethod, &trackingNumber,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if snaggerID != nil && snaggerName != nil {
			method := "in_person"
			if deliveryMethod != nil && *deliveryMethod != "" {
				method = *deliveryMethod
			}
			item.SnaggedBy = &models.SnaggedBy{
				ID:             *snaggerID,
				DisplayName:    *snaggerName,
				DeliveryMethod: method,
				TrackingNumber: trackingNumber,
			}
		}
		items = append(items, item)
	}
	return c.JSON(models.WishlistDetail{
		Wishlist:      w,
		Items:         items,
		ViewerCanSnag: viewerCanSnag,
	})
}

func (h *WishlistHandler) Update(c *fiber.Ctx) error {
	id, err := lookupWishlistID(c.Context(), h.pool, c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	var req createWishlistRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.Title == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "title required"})
	}
	nextSlug, err := allocateWishlistSlug(c.Context(), h.pool, req.Title, &id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	tag, err := h.pool.Exec(c.Context(), `
		UPDATE wishlists SET slug = $3, title = $4, description = $5, is_public = $6
		WHERE id = $1 AND user_id = $2
	`, id, middleware.GetUserID(c), nextSlug, req.Title, req.Description, req.IsPublic)
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	h.syncWishlistItemsIndex(c.Context(), id)
	return c.JSON(fiber.Map{"ok": true, "slug": nextSlug})
}

func (h *WishlistHandler) Delete(c *fiber.Ctx) error {
	id, err := lookupWishlistID(c.Context(), h.pool, c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	tag, err := h.pool.Exec(c.Context(), `DELETE FROM wishlists WHERE id = $1 AND user_id = $2`, id, middleware.GetUserID(c))
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	_ = h.search.DeleteWishlistItems(c.Context(), id.String())
	return c.JSON(fiber.Map{"ok": true})
}

func (h *WishlistHandler) UploadBanner(c *fiber.Ctx) error {
	wishlistID, err := lookupWishlistID(c.Context(), h.pool, c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	userID := middleware.GetUserID(c)
	var ownerID uuid.UUID
	err = h.pool.QueryRow(c.Context(), `SELECT user_id FROM wishlists WHERE id = $1`, wishlistID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	if h.storage == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "storage unavailable"})
	}

	file, err := c.FormFile("banner")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "banner file required"})
	}
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	key := fmt.Sprintf("wishlist-banners/%s/%s", wishlistID, file.Filename)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !strings.HasPrefix(contentType, "image/") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "image file required"})
	}
	if err := h.storage.UploadBanner(c.Context(), key, bytes.NewReader(data), contentType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	_, err = h.pool.Exec(c.Context(), `
		UPDATE wishlists SET banner_key = $2, banner_url = NULL WHERE id = $1 AND user_id = $3
	`, wishlistID, key, userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	bannerURL, err := h.storage.ResolveObjectURL(c.Context(), nil, &key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"banner_url": bannerURL})
}

func (h *WishlistHandler) SetBannerURL(c *fiber.Ctx) error {
	wishlistID, err := lookupWishlistID(c.Context(), h.pool, c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	userID := middleware.GetUserID(c)
	var body struct {
		BannerURL string `json:"banner_url"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if strings.TrimSpace(body.BannerURL) == "" {
		tag, err := h.pool.Exec(c.Context(), `
			UPDATE wishlists SET banner_key = NULL, banner_url = NULL WHERE id = $1 AND user_id = $2
		`, wishlistID, userID)
		if err != nil || tag.RowsAffected() == 0 {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
		}
		return c.JSON(fiber.Map{"banner_url": nil})
	}

	if !isAllowedBannerURL(body.BannerURL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "banner URL must be from Unsplash"})
	}

	url := strings.TrimSpace(body.BannerURL)
	tag, err := h.pool.Exec(c.Context(), `
		UPDATE wishlists SET banner_key = NULL, banner_url = $3 WHERE id = $1 AND user_id = $2
	`, wishlistID, userID, url)
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	return c.JSON(fiber.Map{"banner_url": url})
}

type itemRequest struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Brand    string `json:"brand"`
	Notes    string `json:"notes"`
	ImageURL string `json:"image_url"`
}

func normalizeItemRequest(req *itemRequest) error {
	req.Name = strings.TrimSpace(req.Name)
	req.Type = strings.TrimSpace(req.Type)
	req.Brand = strings.TrimSpace(req.Brand)
	req.Notes = strings.TrimSpace(req.Notes)
	req.ImageURL = strings.TrimSpace(req.ImageURL)
	if req.Name == "" {
		return fmt.Errorf("name required")
	}
	if req.Type == "" {
		req.Type = "Candy"
	}
	if !models.IsValidSnackType(req.Type) {
		return fmt.Errorf("invalid snack type")
	}
	return nil
}

func (h *WishlistHandler) AddItem(c *fiber.Ctx) error {
	wishlistID, err := lookupWishlistID(c.Context(), h.pool, c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	var req itemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := normalizeItemRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	userID := middleware.GetUserID(c)
	var ownerID uuid.UUID
	err = h.pool.QueryRow(c.Context(), `SELECT user_id FROM wishlists WHERE id = $1`, wishlistID).Scan(&ownerID)
	if err != nil || ownerID != userID {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}

	var item models.WishlistItem
	err = h.pool.QueryRow(c.Context(), `
		INSERT INTO wishlist_items (wishlist_id, name, type, brand, notes, image_url)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, wishlist_id, name, type, brand, notes, image_url, created_at, updated_at
	`, wishlistID, req.Name, req.Type, req.Brand, req.Notes, req.ImageURL).Scan(
		&item.ID, &item.WishlistID, &item.Name, &item.Type, &item.Brand, &item.Notes, &item.ImageURL, &item.CreatedAt, &item.UpdatedAt,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	h.syncItemIndex(c.Context(), item.ID)
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *WishlistHandler) UpdateItem(c *fiber.Ctx) error {
	wishlistID, err := lookupWishlistID(c.Context(), h.pool, c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	itemID, _ := uuid.Parse(c.Params("itemId"))
	var req itemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}
	if err := normalizeItemRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}
	tag, err := h.pool.Exec(c.Context(), `
		UPDATE wishlist_items wi SET name = $4, type = $5, brand = $6, notes = $7, image_url = $8
		FROM wishlists w
		WHERE wi.id = $3 AND wi.wishlist_id = w.id AND w.id = $1 AND w.user_id = $2
	`, wishlistID, middleware.GetUserID(c), itemID, req.Name, req.Type, req.Brand, req.Notes, req.ImageURL)
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "item not found"})
	}
	h.syncItemIndex(c.Context(), itemID)
	return c.JSON(fiber.Map{"ok": true})
}

func (h *WishlistHandler) DeleteItem(c *fiber.Ctx) error {
	wishlistID, err := lookupWishlistID(c.Context(), h.pool, c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	itemID, _ := uuid.Parse(c.Params("itemId"))
	tag, err := h.pool.Exec(c.Context(), `
		DELETE FROM wishlist_items wi USING wishlists w
		WHERE wi.id = $3 AND wi.wishlist_id = w.id AND w.id = $1 AND w.user_id = $2
	`, wishlistID, middleware.GetUserID(c), itemID)
	if err != nil || tag.RowsAffected() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "item not found"})
	}
	_ = h.search.DeleteItem(c.Context(), itemID.String())
	return c.JSON(fiber.Map{"ok": true})
}

func (h *WishlistHandler) SnagItem(c *fiber.Ctx) error {
	wishlistID, err := lookupWishlistID(c.Context(), h.pool, c.Params("slug"))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "wishlist not found"})
	}
	itemID, err := uuid.Parse(c.Params("itemId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid item id"})
	}

	viewerID := middleware.GetUserID(c)
	var ownerID uuid.UUID
	err = h.pool.QueryRow(c.Context(), `
		SELECT w.user_id
		FROM wishlists w
		JOIN wishlist_items wi ON wi.wishlist_id = w.id
		WHERE w.id = $1 AND wi.id = $2 AND (w.is_public = TRUE OR w.user_id = $3)
	`, wishlistID, itemID, viewerID).Scan(&ownerID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "item not found"})
	}
	if ownerID == viewerID {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "cannot snag items on your own wishlist"})
	}
	if !hasActiveMatchWith(c.Context(), h.pool, viewerID, ownerID) {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "not matched with this wishlist owner"})
	}

	var body struct {
		DeliveryMethod string `json:"delivery_method"`
		TrackingNumber string `json:"tracking_number"`
	}
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid body"})
		}
	}

	deliveryMethod := strings.TrimSpace(body.DeliveryMethod)
	if deliveryMethod == "" {
		deliveryMethod = "in_person"
	}
	if deliveryMethod != "in_person" && deliveryMethod != "mail" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "delivery_method must be in_person or mail"})
	}

	trackingNumber := strings.TrimSpace(body.TrackingNumber)
	var trackingArg *string
	if deliveryMethod == "mail" && trackingNumber != "" {
		trackingArg = &trackingNumber
	}

	var snaggerName string
	var snagDeliveryMethod string
	var snagTracking *string
	err = h.pool.QueryRow(c.Context(), `
		INSERT INTO wishlist_item_snags (wishlist_item_id, snagged_by_user_id, delivery_method, tracking_number)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (wishlist_item_id) DO NOTHING
		RETURNING
			(SELECT display_name FROM users WHERE id = $2),
			delivery_method,
			tracking_number
	`, itemID, viewerID, deliveryMethod, trackingArg).Scan(&snaggerName, &snagDeliveryMethod, &snagTracking)
	if err != nil {
		var existingSnagger uuid.UUID
		err = h.pool.QueryRow(c.Context(), `
			SELECT s.snagged_by_user_id, s.delivery_method, s.tracking_number
			FROM wishlist_item_snags s
			WHERE s.wishlist_item_id = $1
		`, itemID).Scan(&existingSnagger, &snagDeliveryMethod, &snagTracking)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		if existingSnagger != viewerID {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{"error": "item already snagged"})
		}
		_ = h.pool.QueryRow(c.Context(), `SELECT display_name FROM users WHERE id = $1`, viewerID).Scan(&snaggerName)
	}

	return c.JSON(models.WishlistItem{
		ID:         itemID,
		WishlistID: wishlistID,
		SnaggedBy: &models.SnaggedBy{
			ID:             viewerID,
			DisplayName:    snaggerName,
			DeliveryMethod: snagDeliveryMethod,
			TrackingNumber: snagTracking,
		},
	})
}

func hasActiveMatchWith(ctx context.Context, pool *pgxpool.Pool, userA, userB uuid.UUID) bool {
	var exists bool
	_ = pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM friendships f
			WHERE f.status = 'accepted'
			  AND (($1 = f.requester_id AND $2 = f.addressee_id)
			    OR ($1 = f.addressee_id AND $2 = f.requester_id))
		) OR EXISTS (
			SELECT 1 FROM snack_matches sm
			WHERE sm.status IN ('pending', 'active')
			  AND $1 IN (sm.user_a_id, sm.user_b_id)
			  AND $2 IN (sm.user_a_id, sm.user_b_id)
		)
	`, userA, userB).Scan(&exists)
	return exists
}

func (h *WishlistHandler) syncWishlistItemsIndex(ctx context.Context, wishlistID uuid.UUID) {
	rows, err := h.pool.Query(ctx, `SELECT id FROM wishlist_items WHERE wishlist_id = $1`, wishlistID)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var itemID uuid.UUID
		if err := rows.Scan(&itemID); err != nil {
			return
		}
		h.syncItemIndex(ctx, itemID)
	}
}

func (h *WishlistHandler) syncItemIndex(ctx context.Context, itemID uuid.UUID) {
	var (
		wishlistID                   uuid.UUID
		name, itemType, brand, notes string
		wishlistSlug, wishlistTitle  string
		isPublic                     bool
		userID                       uuid.UUID
		username, displayName        string
		emailVerified                bool
		deactivatedAt                interface{}
	)
	err := h.pool.QueryRow(ctx, `
		SELECT wi.wishlist_id, wi.name, wi.type, wi.brand, wi.notes,
		       w.slug, w.title, w.is_public,
		       u.id, u.username, u.display_name, u.email_verified, u.deactivated_at
		FROM wishlist_items wi
		JOIN wishlists w ON w.id = wi.wishlist_id
		JOIN users u ON u.id = w.user_id
		WHERE wi.id = $1
	`, itemID).Scan(
		&wishlistID, &name, &itemType, &brand, &notes,
		&wishlistSlug, &wishlistTitle, &isPublic,
		&userID, &username, &displayName, &emailVerified, &deactivatedAt,
	)
	if err != nil {
		return
	}

	itemKey := itemID.String()
	if !isPublic || !emailVerified || deactivatedAt != nil {
		_ = h.search.DeleteItem(ctx, itemKey)
		return
	}

	_ = h.search.IndexItem(ctx, search.IndexDocument{
		ID:            itemKey,
		WishlistID:    wishlistID.String(),
		WishlistSlug:  wishlistSlug,
		WishlistTitle: wishlistTitle,
		UserID:        userID.String(),
		Username:      username,
		UserName:      displayName,
		Name:          name,
		Type:          itemType,
		Brand:         brand,
		Notes:         notes,
	})
}

type MatchHandler struct {
	pool    *pgxpool.Pool
	storage *storage.Client
}

func NewMatchHandler(pool *pgxpool.Pool, s *storage.Client) *MatchHandler {
	return &MatchHandler{pool: pool, storage: s}
}

func (h *MatchHandler) RegisterRoutes(app fiber.Router) {
	m := app.Group("/matches")
	m.Get("/", middleware.RequireAuth(h.pool), h.ListMine)
	m.Post("/run", middleware.RequireAuth(h.pool), h.RunPairing)
}

func (h *MatchHandler) ListMine(c *fiber.Ctx) error {
	matches, err := matchingGet(c.Context(), h.pool, middleware.GetUserID(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	enrichMatches(c.Context(), h.storage, matches)
	return c.JSON(matches)
}

func (h *MatchHandler) RunPairing(c *fiber.Ctx) error {
	matches, err := matchingRun(c.Context(), h.pool)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"paired": len(matches), "matches": matches})
}

type UserHandler struct {
	pool    *pgxpool.Pool
	storage *storage.Client
}

func NewUserHandler(pool *pgxpool.Pool, s *storage.Client) *UserHandler {
	return &UserHandler{pool: pool, storage: s}
}

func (h *UserHandler) RegisterRoutes(app fiber.Router) {
	u := app.Group("/users")
	u.Put("/me", middleware.RequireAuth(h.pool), h.UpdateProfile)
	u.Post("/me/avatar", middleware.RequireAuth(h.pool), h.UploadAvatar)
	u.Post("/me/banner", middleware.RequireAuth(h.pool), h.UploadBanner)
	u.Put("/me/banner", middleware.RequireAuth(h.pool), h.SetBannerURL)
	u.Get("/:username", middleware.OptionalAuth(h.pool), h.GetProfile)
}

func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	rec, err := auth.GetUserByUsername(c.Context(), h.pool, c.Params("username"))
	if err != nil || !rec.EmailVerified || rec.DeactivatedAt != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	profile := models.PublicUser{
		ID:          rec.ID,
		Username:    rec.Username,
		DisplayName: rec.DisplayName,
		Bio:         rec.Bio,
		Country:     rec.Country,
	}
	if h.storage != nil {
		if avatarURL, err := h.storage.ResolveObjectURL(c.Context(), rec.AvatarURL, rec.AvatarKey); err == nil && avatarURL != "" {
			profile.AvatarURL = &avatarURL
		}
		profile.BannerURL = resolvePublicBannerURL(c.Context(), h.storage, rec.BannerURL, rec.BannerKey)
	} else if rec.BannerURL != nil && strings.TrimSpace(*rec.BannerURL) != "" {
		profile.BannerURL = rec.BannerURL
	}

	rows, err := h.pool.Query(c.Context(), `
		SELECT w.id, w.user_id, w.slug, w.title, w.description, w.is_public, w.banner_key, w.banner_url,
		       w.created_at, w.updated_at,
		       (SELECT COUNT(*)::int FROM wishlist_items wi WHERE wi.wishlist_id = w.id)
		FROM wishlists w
		WHERE w.user_id = $1 AND w.is_public = TRUE
		ORDER BY w.updated_at DESC
	`, rec.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer rows.Close()

	lists := make([]models.Wishlist, 0)
	for rows.Next() {
		var w models.Wishlist
		var bannerKey, bannerURL *string
		if err := rows.Scan(
			&w.ID, &w.UserID, &w.Slug, &w.Title, &w.Description, &w.IsPublic, &bannerKey, &bannerURL,
			&w.CreatedAt, &w.UpdatedAt, &w.ItemCount,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		populateWishlistBanner(c.Context(), h.storage, &w, bannerKey, bannerURL)
		lists = append(lists, w)
	}
	if err := rows.Err(); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	var friendship *models.FriendshipView
	viewerID := middleware.GetUserID(c)
	profileUserID := rec.ID
	if viewerID != uuid.Nil && viewerID != profileUserID {
		if friendshipRec, err := friends.LookupBetween(c.Context(), h.pool, viewerID, profileUserID); err == nil {
			friendship = friends.ViewFor(viewerID, profileUserID, friendshipRec)
		}
	}

	return c.JSON(models.UserProfile{User: profile, Wishlists: lists, Friendship: friendship})
}

func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var body struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
		Country     string `json:"country"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	rec, err := auth.GetUserByID(c.Context(), h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	country := strings.ToUpper(strings.TrimSpace(body.Country))
	if rec.DiscordID != nil && *rec.DiscordID != "" {
		_, err = h.pool.Exec(c.Context(), `
			UPDATE users SET bio = $2, country = $3 WHERE id = $1
		`, userID, body.Bio, country)
	} else {
		_, err = h.pool.Exec(c.Context(), `
			UPDATE users SET display_name = $2, bio = $3, country = $4 WHERE id = $1
		`, userID, body.DisplayName, body.Bio, country)
	}
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *UserHandler) UploadAvatar(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	rec, err := auth.GetUserByID(c.Context(), h.pool, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}
	if rec.DiscordID != nil && *rec.DiscordID != "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "avatar is synced from Discord"})
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "avatar file required"})
	}
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	key := fmt.Sprintf("avatars/%s/%s", userID, file.Filename)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if err := h.storage.UploadAvatar(c.Context(), key, bytes.NewReader(data), contentType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	_, err = h.pool.Exec(c.Context(), `
		UPDATE users SET avatar_key = $2, avatar_url = NULL, discord_avatar_hash = NULL WHERE id = $1
	`, userID, key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	avatarURL, err := h.storage.ResolveObjectURL(c.Context(), nil, &key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"avatar_url": avatarURL})
}

func (h *UserHandler) UploadBanner(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	if _, err := auth.GetUserByID(c.Context(), h.pool, userID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "user not found"})
	}

	file, err := c.FormFile("banner")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "banner file required"})
	}
	f, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	key := fmt.Sprintf("banners/%s/%s", userID, file.Filename)
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if !strings.HasPrefix(contentType, "image/") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "image file required"})
	}
	if err := h.storage.UploadBanner(c.Context(), key, bytes.NewReader(data), contentType); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	_, err = h.pool.Exec(c.Context(), `
		UPDATE users SET banner_key = $2, banner_url = NULL WHERE id = $1
	`, userID, key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	bannerURL, err := h.storage.ResolveObjectURL(c.Context(), nil, &key)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"banner_url": bannerURL})
}

func (h *UserHandler) SetBannerURL(c *fiber.Ctx) error {
	userID := middleware.GetUserID(c)
	var body struct {
		BannerURL string `json:"banner_url"`
	}
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid request"})
	}

	if strings.TrimSpace(body.BannerURL) == "" {
		_, err := h.pool.Exec(c.Context(), `
			UPDATE users SET banner_key = NULL, banner_url = NULL WHERE id = $1
		`, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"banner_url": nil})
	}

	if !isAllowedBannerURL(body.BannerURL) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "banner URL must be from Unsplash"})
	}

	url := strings.TrimSpace(body.BannerURL)
	_, err := h.pool.Exec(c.Context(), `
		UPDATE users SET banner_key = NULL, banner_url = $2 WHERE id = $1
	`, userID, url)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"banner_url": url})
}

type SearchHandler struct {
	pool        *pgxpool.Pool
	meili       *search.Client
	storage     *storage.Client
	snackSearch *snacksearch.Service
}

func NewSearchHandler(pool *pgxpool.Pool, meili *search.Client, storage *storage.Client, s *snacksearch.Service) *SearchHandler {
	return &SearchHandler{pool: pool, meili: meili, storage: storage, snackSearch: s}
}

func (h *SearchHandler) RegisterRoutes(app fiber.Router) {
	app.Get("/search", h.Search)
}

func (h *SearchHandler) Search(c *fiber.Ctx) error {
	q := strings.TrimSpace(c.Query("q"))
	resp, err := search.UniversalSearch(c.Context(), h.pool, h.meili, h.snackSearch, h.storage, q)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	if resp.Users == nil {
		resp.Users = []search.PublicUserHit{}
	}
	if resp.WishlistItems == nil {
		resp.WishlistItems = []search.WishlistItemHit{}
	}
	if resp.Products == nil {
		resp.Products = []snacksearch.ProductHit{}
	}
	return c.JSON(resp)
}

// avoid import cycle with matching package in handler signatures
func matchingGet(ctx context.Context, pool *pgxpool.Pool, userID uuid.UUID) ([]models.SnackMatch, error) {
	return matchPkgGet(ctx, pool, userID)
}

func matchingRun(ctx context.Context, pool *pgxpool.Pool) ([]models.SnackMatch, error) {
	return matchPkgRun(ctx, pool)
}
