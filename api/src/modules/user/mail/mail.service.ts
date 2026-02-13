import { sendMail } from "#src/modules/shared/mail";
import accountConfirmationTemplate from "./confirmation.js";
import passwordResetTemplate from "./password-reset.js";

export async function sendMailConfirmToUser(email: string, token: string) {
  await sendMail({
    to: email,
    subject: "Confirm your email address",
    body: accountConfirmationTemplate({
      FRONTEND_URL: process.env.FRONTEND_URL,
      token: token,
    }),
  });
}

export async function sendMailResetPassword(email: string, token: string) {
  await sendMail({
    to: email,
    subject: "Reset your password",
    body: passwordResetTemplate({
      FRONTEND_URL: process.env.FRONTEND_URL,
      token: token,
    }),
  });
}
