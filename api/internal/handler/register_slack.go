package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

// registerSlack exposes the same four operations for the organization webhook
// and for a shop webhook.
func registerSlack(api huma.API, h *Handler) {
	const orgPath = "/organizations/{orgId}/slack-webhook"
	const shopPath = "/organizations/{orgId}/shops/{shopId}/slack-webhook"

	huma.Register(api, slackGetOperation("getSlackWebhook", orgPath, "the organization"), h.GetSlackWebhook)
	huma.Register(api, slackUpdateOperation("updateSlackWebhook", orgPath, "the organization"), h.UpdateSlackWebhook)
	huma.Register(api, slackDeleteOperation("deleteSlackWebhook", orgPath, "the organization"), h.DeleteSlackWebhook)
	huma.Register(api, slackTestOperation("testSlackWebhook", orgPath, "the organization"), h.TestSlackWebhook)

	huma.Register(api, slackGetOperation("getShopSlackWebhook", shopPath, "a shop"), h.GetShopSlackWebhook)
	huma.Register(api, slackUpdateOperation("updateShopSlackWebhook", shopPath, "a shop"), h.UpdateShopSlackWebhook)
	huma.Register(api, slackDeleteOperation("deleteShopSlackWebhook", shopPath, "a shop"), h.DeleteShopSlackWebhook)
	huma.Register(api, slackTestOperation("testShopSlackWebhook", shopPath, "a shop"), h.TestShopSlackWebhook)
}

func slackGetOperation(id, path, owner string) huma.Operation {
	return huma.Operation{
		OperationID: id,
		Method:      http.MethodGet,
		Path:        path,
		Summary:     "Get whether " + owner + " has a Slack webhook",
		Tags:        []string{"Slack"},
		Errors:      []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}
}

func slackUpdateOperation(id, path, owner string) huma.Operation {
	return huma.Operation{
		OperationID:   id,
		Method:        http.MethodPut,
		Path:          path,
		Summary:       "Set the Slack webhook of " + owner,
		Tags:          []string{"Slack"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusBadRequest},
	}
}

func slackDeleteOperation(id, path, owner string) huma.Operation {
	return huma.Operation{
		OperationID:   id,
		Method:        http.MethodDelete,
		Path:          path,
		Summary:       "Remove the Slack webhook of " + owner,
		Tags:          []string{"Slack"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound},
	}
}

func slackTestOperation(id, path, owner string) huma.Operation {
	return huma.Operation{
		OperationID:   id,
		Method:        http.MethodPost,
		Path:          path + "/test",
		Summary:       "Post a test message to the Slack webhook of " + owner,
		Tags:          []string{"Slack"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound, http.StatusBadGateway},
	}
}
