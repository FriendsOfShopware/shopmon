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

    // https://blog.cloudflare.com/sending-email-from-workers-with-mailchannels
    const response = await fetch('https://api.mailchannels.net/tx/v1/send', {
        method: 'POST',
        headers: {
            'content-type': 'application/json',
        },
        body: JSON.stringify({
            personalizations: [
                {
                    to: [{ email: mail.to }],
                    ...(() => {
                        if (env.MAIL_DKIM_PRIVATE_KEY === undefined) {
                            return {};
                        }

                        return {
                            dkim_domain: env.MAIL_DKIM_DOMAIN,
                            dkim_selector: env.MAIL_DKIM_SELECTOR,
                            dkim_private_key: env.MAIL_DKIM_PRIVATE_KEY,
                        };
                    })(),
                },
            ],
            from: {
                email: env.MAIL_FROM,
                name: env.MAIL_FROM_NAME,
            },
            subject: mail.subject,
            content: [
                {
                    type: 'text/html',
                    value: mail.body,
                },
            ],
        }),
    });

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
