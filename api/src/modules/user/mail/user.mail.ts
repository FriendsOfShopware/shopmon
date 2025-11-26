import { sendMail } from '#src/modules/shared/mail/mail.service.ts';
import accountConfirmationTemplate from './confirmation.mjml';
import passwordResetTemplate from './password-reset.mjml';

export async function sendMailConfirmToUser(email: string, token: string) {
    await sendMail({
        to: email,
        subject: 'Confirm your email address',
        body: accountConfirmationTemplate({
            FRONTEND_URL: process.env.FRONTEND_URL,
            token: token,
        }),
    });
}

export async function sendMailResetPassword(email: string, token: string) {
    await sendMail({
        to: email,
        subject: 'Reset your password',
        body: passwordResetTemplate({
            FRONTEND_URL: process.env.FRONTEND_URL,
            token: token,
        }),
    });
}
