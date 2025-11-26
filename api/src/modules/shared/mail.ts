import nodemailer from 'nodemailer';

export interface MailRequest {
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

export async function sendMail(mail: MailRequest) {
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
