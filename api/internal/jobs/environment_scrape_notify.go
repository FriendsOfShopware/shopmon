package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/jackc/pgx/v5/pgtype"
)

// handleStatusChange detects degradation and sends notifications.
func (h *EnvironmentScrapeHandler) handleStatusChange(ctx context.Context, env queries.GetAllEnvironmentsRow, newStatus string) {
	envDetail, err := h.queries.GetEnvironmentByID(ctx, env.ID)
	if err != nil {
		slog.Warn("failed to get environment details for status comparison", "environmentId", env.ID, "error", err)
		return
	}

	oldWeight := statusWeight[envDetail.Status]
	newWeight := statusWeight[newStatus]

	if newWeight <= oldWeight {
		return // not degraded
	}

	// Status degraded - send notifications
	subscribers, err := h.queries.GetEnvironmentNotificationSubscribers(ctx, queries.GetEnvironmentNotificationSubscribersParams{
		OrganizationID: env.OrganizationID,
		EnvironmentID:  strconv.Itoa(int(env.ID)),
	})
	if err != nil {
		slog.Warn("failed to get notification subscribers", "environmentId", env.ID, "error", err)
		return
	}

	statusChangeKey := fmt.Sprintf("environment.change-status.%d", env.ID)

	// Check alert dedup lock — don't send emails if already alerted within the last hour
	alertLockKey := fmt.Sprintf("alert_%s", statusChangeKey)
	locked, _ := h.queries.IsLocked(ctx, alertLockKey)
	sendEmails := !locked

	if sendEmails {
		// Acquire lock for 1 hour to prevent duplicate alerts
		if err := h.queries.AcquireLock(ctx, queries.AcquireLockParams{
			Key:     alertLockKey,
			Expires: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
		}); err != nil {
			slog.Error("failed to acquire lock", "key", alertLockKey, "error", err)
		}
	}

	linkJSON, _ := json.Marshal(notificationLink{
		Name: "account.environments.detail",
		Params: map[string]string{
			"environmentId": strconv.Itoa(int(env.ID)),
		},
	})

	alertMessage := fmt.Sprintf("Status changed from %s to %s", envDetail.Status, newStatus)

	for _, user := range subscribers {
		if err := h.queries.UpsertNotification(ctx, queries.UpsertNotificationParams{
			UserID:  user.ID,
			Key:     statusChangeKey,
			Level:   "warning",
			Title:   fmt.Sprintf("Environment: %s status changed", env.Name),
			Message: alertMessage,
			Link:    linkJSON,
		}); err != nil {
			slog.Warn("failed to send status change notification", "userId", user.ID, "environmentId", env.ID, "error", err)
		}

		// Send email alert (only if not deduped)
		if sendEmails && h.mail != nil {
			subject := fmt.Sprintf("Environment %s status changed to %s", env.Name, newStatus)
			body := mail.BuildShopAlertEmail(user.Name, env.Name, alertMessage)
			if err := h.mail.Send(user.Email, subject, body); err != nil {
				slog.Warn("failed to send status change email", "userId", user.ID, "email", user.Email, "error", err)
			}
		}
	}
}

// notifyAuthError sends notifications when authentication fails.
func (h *EnvironmentScrapeHandler) notifyAuthError(ctx context.Context, env queries.GetAllEnvironmentsRow, authErr error) {
	subscribers, err := h.queries.GetEnvironmentNotificationSubscribers(ctx, queries.GetEnvironmentNotificationSubscribersParams{
		OrganizationID: env.OrganizationID,
		EnvironmentID:  strconv.Itoa(int(env.ID)),
	})
	if err != nil {
		slog.Warn("failed to get notification subscribers", "environmentId", env.ID, "error", err)
		return
	}

	notifKey := fmt.Sprintf("environment.update-auth-error.%d", env.ID)

	// Check alert dedup lock
	alertLockKey := fmt.Sprintf("alert_%s", notifKey)
	locked, _ := h.queries.IsLocked(ctx, alertLockKey)
	sendEmails := !locked

	if sendEmails {
		if err := h.queries.AcquireLock(ctx, queries.AcquireLockParams{
			Key:     alertLockKey,
			Expires: pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
		}); err != nil {
			slog.Error("failed to acquire lock", "key", alertLockKey, "error", err)
		}
	}

	linkJSON, _ := json.Marshal(notificationLink{
		Name: "account.environments.detail",
		Params: map[string]string{
			"environmentId": strconv.Itoa(int(env.ID)),
		},
	})

	errMsg := fmt.Sprintf("%v", authErr)
	if len(errMsg) > 50 {
		errMsg = errMsg[:50] + "..."
	}

	alertMessage := fmt.Sprintf("Could not connect to environment. Please check your credentials and try again. %s", errMsg)

	for _, user := range subscribers {
		if err := h.queries.UpsertNotification(ctx, queries.UpsertNotificationParams{
			UserID:  user.ID,
			Key:     notifKey,
			Level:   "error",
			Title:   fmt.Sprintf("Environment: %s could not be updated", env.Name),
			Message: alertMessage,
			Link:    linkJSON,
		}); err != nil {
			slog.Warn("failed to send auth error notification", "userId", user.ID, "environmentId", env.ID, "error", err)
		}

		// Send email alert (only if not deduped)
		if sendEmails && h.mail != nil {
			subject := fmt.Sprintf("Environment %s connection failed", env.Name)
			body := mail.BuildShopAlertEmail(user.Name, env.Name, alertMessage)
			if err := h.mail.Send(user.Email, subject, body); err != nil {
				slog.Warn("failed to send auth error email", "userId", user.ID, "error", err)
			}
		}
	}
}

// notifyDataFetchError sends notifications when data fetching fails.
func (h *EnvironmentScrapeHandler) notifyDataFetchError(ctx context.Context, env queries.GetAllEnvironmentsRow, errMsg string) {
	subscribers, err := h.queries.GetEnvironmentNotificationSubscribers(ctx, queries.GetEnvironmentNotificationSubscribersParams{
		OrganizationID: env.OrganizationID,
		EnvironmentID:  strconv.Itoa(int(env.ID)),
	})
	if err != nil {
		slog.Warn("failed to get notification subscribers", "environmentId", env.ID, "error", err)
		return
	}

	notifKey := fmt.Sprintf("environment.not.updated_%d", env.ID)

	linkJSON, _ := json.Marshal(notificationLink{
		Name: "account.environments.detail",
		Params: map[string]string{
			"environmentId": strconv.Itoa(int(env.ID)),
		},
	})

	for _, user := range subscribers {
		if err := h.queries.UpsertNotification(ctx, queries.UpsertNotificationParams{
			UserID:  user.ID,
			Key:     notifKey,
			Level:   "error",
			Title:   fmt.Sprintf("Environment: %s could not be updated", env.Name),
			Message: "Could not connect to environment. Please check your credentials and try again.",
			Link:    linkJSON,
		}); err != nil {
			slog.Warn("failed to send data fetch error notification", "userId", user.ID, "environmentId", env.ID, "error", err)
		}
	}
}
