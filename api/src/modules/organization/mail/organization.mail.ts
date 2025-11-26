import { sendMail } from '#src/modules/shared/mail/mail.service.ts';
import organizationInviteTemplate from './org-invite.mjml';

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
