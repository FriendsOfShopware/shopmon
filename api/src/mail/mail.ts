import type { Shop, ShopAlert, User } from '../repository/shops';
// @ts-expect-error
import passwordResetTemplate from './sources/password-reset.js';
// @ts-expect-error
import alertTemplate from './sources/alert.js';
// @ts-expect-error
import accountConfirmationTemplate from './sources/confirmation.js';
import { Bindings } from '../router';
import { Resend } from 'resend';

interface MaiLRequest {
    to: string;
    subject: string;
    body: string;
}

async function sendMail(env: Bindings, mail: MaiLRequest) {
    if (env.MAIL_ACTIVE === 'false') {
        console.log(`Sending mail to ${mail.to} with subject ${mail.subject}`);
        console.log(mail.body);
        return;
    }

    const resend = new Resend(env.RESEND_API_KEY);
    const resp = await resend.emails.send({
        from: env.MAIL_FROM,
        to: [mail.to],
        subject: mail.subject,
        html: mail.body,
    });

    if (resp.error) {
        console.error(`Failed to send mail to ${mail.to} with subject ${mail.subject}`, resp.error);
        throw new Error(`Failed to send mail: ${resp.error.message}`);
    }

    console.log(`Mail sent to ${mail.to} with subject ${mail.subject}`);
}

export async function sendMailConfirmToUser(
    env: Bindings,
    email: string,
    token: string,
) {
    await sendMail(env, {
        to: email,
        subject: 'Confirm your email address',
        body: accountConfirmationTemplate({
            FRONTEND_URL: env.FRONTEND_URL,
            token: token,
        }),
    });
}

export async function sendMailResetPassword(
    env: Bindings,
    email: string,
    token: string,
) {
    await sendMail(env, {
        to: email,
        subject: 'Reset your password',
        body: passwordResetTemplate({
            FRONTEND_URL: env.FRONTEND_URL,
            token: token,
        }),
    });
}

export async function sendAlert(
    env: Bindings,
    shop: Shop,
    user: User,
    alert: ShopAlert,
) {
    await sendMail(env, {
        to: user.email,
        subject: `Shop Alert: ${alert.subject} - ${shop.name}`,
        body: alertTemplate({
            FRONTEND_URL: env.FRONTEND_URL,
            shop,
            alert,
            user,
        }),
    });
}
