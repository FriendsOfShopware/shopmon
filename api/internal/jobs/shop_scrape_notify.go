package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
)

// handleStatusChange detects degradation and sends notifications.
func (h *ShopScrapeHandler) handleStatusChange(ctx context.Context, shop queries.GetAllShopsRow, newStatus string) {
	// We need to get the current shop status from the database since GetAllShopsRow doesn't include status.
	// We compare based on what we know: if the new status is worse than the old status stored in DB.
	shopDetail, err := h.queries.GetShopByID(ctx, shop.ID)
	if err != nil {
		slog.Warn("failed to get shop details for status comparison", "shopId", shop.ID, "error", err)
		return
	}

	oldWeight := statusWeight[shopDetail.Status]
	newWeight := statusWeight[newStatus]

	if newWeight <= oldWeight {
		return // not degraded
	}

	// Status degraded - send notifications
	subscribers, err := h.queries.GetShopNotificationSubscribers(ctx, shop.OrganizationID)
	if err != nil {
		slog.Warn("failed to get notification subscribers", "shopId", shop.ID, "error", err)
		return
	}

	statusChangeKey := fmt.Sprintf("shop.change-status.%d", shop.ID)

	// Check alert dedup lock — don't send emails if already alerted within the last hour
	alertLockKey := fmt.Sprintf("alert_%s", statusChangeKey)
	locked, _ := h.queries.IsLocked(ctx, alertLockKey)
	sendEmails := !locked

	if sendEmails {
		// Acquire lock for 1 hour to prevent duplicate alerts
		h.queries.AcquireLock(ctx, queries.AcquireLockParams{
			Key:     alertLockKey,
			Expires: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
		})
	}

	linkJSON, _ := json.Marshal(notificationLink{
		Name: "account.shops.detail",
		Params: map[string]string{
			"shopId": strconv.Itoa(int(shop.ID)),
		},
	})

	alertMessage := fmt.Sprintf("Status changed from %s to %s", shopDetail.Status, newStatus)

	for _, user := range subscribers {
		if !isUserSubscribedToShop(user, shop.ID) {
			continue
		}

		if err := h.queries.UpsertNotification(ctx, queries.UpsertNotificationParams{
			UserID:  user.ID,
			Key:     statusChangeKey,
			Level:   "warning",
			Title:   fmt.Sprintf("Shop: %s status changed", shop.Name),
			Message: alertMessage,
			Link:    linkJSON,
		}); err != nil {
			slog.Warn("failed to send status change notification", "userId", user.ID, "shopId", shop.ID, "error", err)
		}

		// Send email alert (only if not deduped)
		if sendEmails && h.mail != nil {
			subject := fmt.Sprintf("Shop %s status changed to %s", shop.Name, newStatus)
			body := mail.BuildShopAlertEmail(user.Name, shop.Name, alertMessage)
			if err := h.mail.Send(user.Email, subject, body); err != nil {
				slog.Warn("failed to send status change email", "userId", user.ID, "email", user.Email, "error", err)
			}
		}
	}
}

// notifyAuthError sends notifications when authentication fails.
func (h *ShopScrapeHandler) notifyAuthError(ctx context.Context, shop queries.GetAllShopsRow, authErr error) {
	subscribers, err := h.queries.GetShopNotificationSubscribers(ctx, shop.OrganizationID)
	if err != nil {
		slog.Warn("failed to get notification subscribers", "shopId", shop.ID, "error", err)
		return
	}

	notifKey := fmt.Sprintf("shop.update-auth-error.%d", shop.ID)

	// Check alert dedup lock
	alertLockKey := fmt.Sprintf("alert_%s", notifKey)
	locked, _ := h.queries.IsLocked(ctx, alertLockKey)
	sendEmails := !locked

	if sendEmails {
		h.queries.AcquireLock(ctx, queries.AcquireLockParams{
			Key:     alertLockKey,
			Expires: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
		})
	}

	linkJSON, _ := json.Marshal(notificationLink{
		Name: "account.shops.detail",
		Params: map[string]string{
			"shopId": strconv.Itoa(int(shop.ID)),
		},
	})

	errMsg := fmt.Sprintf("%v", authErr)
	if len(errMsg) > 50 {
		errMsg = errMsg[:50] + "..."
	}

	alertMessage := fmt.Sprintf("Could not connect to shop. Please check your credentials and try again. %s", errMsg)

	for _, user := range subscribers {
		if !isUserSubscribedToShop(user, shop.ID) {
			continue
		}

		if err := h.queries.UpsertNotification(ctx, queries.UpsertNotificationParams{
			UserID:  user.ID,
			Key:     notifKey,
			Level:   "error",
			Title:   fmt.Sprintf("Shop: %s could not be updated", shop.Name),
			Message: alertMessage,
			Link:    linkJSON,
		}); err != nil {
			slog.Warn("failed to send auth error notification", "userId", user.ID, "shopId", shop.ID, "error", err)
		}

		// Send email alert (only if not deduped)
		if sendEmails && h.mail != nil {
			subject := fmt.Sprintf("Shop %s connection failed", shop.Name)
			body := mail.BuildShopAlertEmail(user.Name, shop.Name, alertMessage)
			if err := h.mail.Send(user.Email, subject, body); err != nil {
				slog.Warn("failed to send auth error email", "userId", user.ID, "error", err)
			}
		}
	}
}

// notifyDataFetchError sends notifications when data fetching fails.
func (h *ShopScrapeHandler) notifyDataFetchError(ctx context.Context, shop queries.GetAllShopsRow, errMsg string) {
	subscribers, err := h.queries.GetShopNotificationSubscribers(ctx, shop.OrganizationID)
	if err != nil {
		slog.Warn("failed to get notification subscribers", "shopId", shop.ID, "error", err)
		return
	}

	notifKey := fmt.Sprintf("shop.not.updated_%d", shop.ID)

	linkJSON, _ := json.Marshal(notificationLink{
		Name: "account.shops.detail",
		Params: map[string]string{
			"shopId": strconv.Itoa(int(shop.ID)),
		},
	})

	for _, user := range subscribers {
		if !isUserSubscribedToShop(user, shop.ID) {
			continue
		}

		if err := h.queries.UpsertNotification(ctx, queries.UpsertNotificationParams{
			UserID:  user.ID,
			Key:     notifKey,
			Level:   "error",
			Title:   fmt.Sprintf("Shop: %s could not be updated", shop.Name),
			Message: "Could not connect to shop. Please check your credentials and try again.",
			Link:    linkJSON,
		}); err != nil {
			slog.Warn("failed to send data fetch error notification", "userId", user.ID, "shopId", shop.ID, "error", err)
		}
	}
}

// isUserSubscribedToShop checks if a user has subscribed to notifications for a specific shop.
func isUserSubscribedToShop(user queries.GetShopNotificationSubscribersRow, shopID int32) bool {
	if user.Notifications == nil {
		return false
	}

	var notifications []string
	if err := json.Unmarshal(user.Notifications, &notifications); err != nil {
		return false
	}

	shopKey := fmt.Sprintf("shop-%d", shopID)
	for _, n := range notifications {
		if n == shopKey {
			return true
		}
	}
	return false
}
