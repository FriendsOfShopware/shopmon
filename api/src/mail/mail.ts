import nodemailer from 'nodemailer';
import type { Shop, ShopAlert, User } from '../repository/shops.ts';
import alertTemplate from './sources/alert.js';
import accountConfirmationTemplate from './sources/confirmation.js';
import organizationInviteTemplate from './sources/org-invite.js';
import passwordResetTemplate from './sources/password-reset.js';

interface MaiLRequest {
    to: string;
    subject: string;
    body: string;
}

const transport = nodemailer.createTransport({
    host: process.env.SMTP_HOST || 'localhost',
    port: Number.parseInt(process.env.SMTP_PORT || '1025', 10),
    secure: process.env.SMTP_SECURE === 'true',
    auth:
        process.env.SMTP_USER && process.env.SMTP_PASS
            ? {
                  user: process.env.SMTP_USER,
                  pass: process.env.SMTP_PASS,
              }
            : undefined,
});

async function sendMail(mail: MaiLRequest) {
    try {
        await transport.sendMail({
            from: process.env.MAIL_FROM,
            to: mail.to,
            subject: mail.subject,
            html: mail.body,
            ...(process.env.SMTP_REPLY_TO && {
                replyTo: process.env.SMTP_REPLY_TO,
            }),
        });
    } catch (error) {
        console.error(
            `Failed to send mail to ${mail.to} with subject ${mail.subject}`,
            error,
        );
        throw new Error(`Failed to send mail: ${error.message}`);
    }

    console.log(`Mail sent to ${mail.to} with subject ${mail.subject}`);
}

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

export async function sendAlert(shop: Shop, user: User, alert: ShopAlert) {
    await sendMail({
        to: user.email,
        subject: `Shop Alert: ${alert.subject} - ${shop.name}`,
        body: alertTemplate({
            FRONTEND_URL: process.env.FRONTEND_URL,
            shop,
            alert,
            user,
        }),
    });
}

export async function sendMailInviteToOrganization(
    email: string,
    organizationName: string,
    invitedByUsername: string,
    token: string,
) {
    await sendMail({
        to: email,
        subject: `You have been invited to join ${organizationName} at Shopmon`,
        body: organizationInviteTemplate({
            FRONTEND_URL: process.env.FRONTEND_URL,
            token,
            organizationName,
            invitedByUsername,
        }),
    });
}
