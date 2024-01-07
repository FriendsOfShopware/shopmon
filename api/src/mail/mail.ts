import type { Shop, ShopAlert, User } from "../repository/shops";
import { createSentry } from "../toucan";

interface MaiLRequest {
    to: string;
    subject: string;
    body: string;
}

async function sendMail(env: Env, mail: MaiLRequest) {
    if (env.MAIL_ACTIVE === 'false') {
        console.log(`Sending mail to ${mail.to} with subject ${mail.subject}`)
        console.log(mail.body);
        return;
    }

    // https://blog.cloudflare.com/sending-email-from-workers-with-mailchannels
    const response = await fetch(`https://api.mailchannels.net/tx/v1/send`, {
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

export async function sendMailConfirmToUser(env: Env, email: string, token: string) {
    const mailBody = `
        <p>Hi,</p>
        <p>Thank you for registering with us. Please click the link below to confirm your email address.</p>

        <p><a href="${env.FRONTEND_URL}/account/confirm/${token}">Confirm email</a></p>

        <p>Best regards,</p>
        <p>FriendsOfShopware</p>
    `;

    await sendMail(env, {
        to: email,
        subject: 'Confirm your email address',
        body: mailBody
    });
}

export async function sendMailResetPassword(env: Env, email: string, token: string) {
    const mailBody = `
        <p>Hi,</p>
        <p>Please click the link below to reset your password.</p>

        <p><a href="${env.FRONTEND_URL}/account/forgot-password/${token}">Reset password</a></p>

        <p>Best regards,</p>
        <p>FriendsOfShopware</p>
    `;

    await sendMail(env, {
        to: email,
        subject: 'Reset your password',
        body: mailBody
    });
}

export async function sendAlert(env: Env, shop: Shop, user: User, alert: ShopAlert) {
    await sendMail(env, {
        to: user.email,
        subject: `Shop Alert: ${alert.subject} - ${shop.name}`,
        body: `Hello ${user.username},<br><br>

        The Shop ${shop.name} has an new alert: <br><br>

        ${alert.message}.
        `
    })
}
