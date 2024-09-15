import type { Shop, ShopAlert, User } from '../repository/shops';
// @ts-expect-error
import passwordResetTemplate from './sources/password-reset.js';
// @ts-expect-error
import alertTemplate from './sources/alert.js';
// @ts-expect-error
import accountConfirmationTemplate from './sources/confirmation.js';
import { Bindings } from '../router';

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

    const formData = new FormData();
    formData.append('from', env.MAIL_FROM);
    formData.append('to', mail.to);
    formData.append('subject', mail.subject);
    formData.append('html', mail.body);

    const response = await fetch(
        `https://api.eu.mailgun.net/v3/${env.MAILGUN_DOMAIN}/messages`,
        {
            method: 'POST',
            body: formData,
            headers: {
                Authorization: 'Basic ' + btoa('api:' + env.MAILGUN_KEY),
            },
        },
    );

    if (!response.ok) {
        throw new Error(`Failed to send mail: ${await response.text()}`);
    }
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
