package mail

import "github.com/matcornic/hermes"

func newHermes() hermes.Hermes {
	return hermes.Hermes{
		Product: hermes.Product{
			Name:      "Shopmon",
			Link:      "https://shopmon.shop",
			Copyright: "Best Regards, FriendsOfShopware",
		},
	}
}

// BuildConfirmationEmail returns HTML for the email verification email.
func BuildConfirmationEmail(name, verifyURL string) string {
	return generate(hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				"Thank you for registering with us. Please confirm your email address by clicking the button below.",
			},
			Actions: []hermes.Action{
				{
					Button: hermes.Button{
						Color: "#2563eb",
						Text:  "Confirm Email",
						Link:  verifyURL,
					},
				},
			},
			Outros: []string{
				"If you didn't create an account, you can safely ignore this email.",
			},
		},
	})
}

// BuildPasswordResetEmail returns HTML for the password reset email.
func BuildPasswordResetEmail(name, resetURL string) string {
	return generate(hermes.Email{
		Body: hermes.Body{
			Name: name,
			Intros: []string{
				"We received a request to reset your password. Please click the button below to set a new password.",
			},
			Actions: []hermes.Action{
				{
					Button: hermes.Button{
						Color: "#2563eb",
						Text:  "Reset Password",
						Link:  resetURL,
					},
				},
			},
			Outros: []string{
				"This link will expire in 1 hour. If you didn't request a password reset, you can safely ignore this email.",
			},
		},
	})
}

// BuildOrgInviteEmail returns HTML for the organization invitation email.
func BuildOrgInviteEmail(inviterName, orgName, acceptURL, rejectURL string) string {
	return generate(hermes.Email{
		Body: hermes.Body{
			Intros: []string{inviterName + " has invited you to join **" + orgName + "** on Shopmon."},
			Actions: []hermes.Action{
				{
					Button: hermes.Button{
						Color: "#2563eb",
						Text:  "Accept Invitation",
						Link:  acceptURL,
					},
				},
				{
					Button: hermes.Button{
						Color: "#6b7280",
						Text:  "Decline",
						Link:  rejectURL,
					},
				},
			},
		},
	})
}

// BuildShopAlertEmail returns HTML for the shop alert email.
func BuildShopAlertEmail(userName, shopName, alertMessage string) string {
	return generate(hermes.Email{
		Body: hermes.Body{
			Name: userName,
			Intros: []string{
				"There is an alert for shop **" + shopName + "**:",
				alertMessage,
			},
		},
	})
}

func generate(email hermes.Email) string {
	h := newHermes()
	body, _ := h.GenerateHTML(email)
	return body
}
