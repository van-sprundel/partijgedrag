import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as sgMail from '@sendgrid/mail';

@Injectable()
export class SendgridService {
  private sgMail: sgMail.MailService;

  constructor(private readonly configService: ConfigService) {
    this.sgMail = sgMail.default;
    this.sgMail.setApiKey(configService.get('SENDGRID_API_TOKEN')!);
  }

  async sendWelcomeEmail(email: string): Promise<void> {
    const msg = {
      to: email,
      from: 'no-reply@merlijnmac.nl',
      subject: '_indexWelcome to NestJS Remix Template',
      text: 'and easy to do anywhere, even with Node.js',
      html: '<strong>and easy to do anywhere, even with Node.js</strong>',
    };

    await this.sgMail.send(msg);
  }

  async sendPasswordResetEmail(email: string, accessToken: string, redirectUrl: string): Promise<void> {
    const msg = {
      to: email,
      from: 'no-reply@merlijnmac.nl',
      subject: 'Reset your password for NestJS Remix Template',
      text: 'Reset Your Password for NestJS Remix Template',
      html: `<!doctype html>
<html>
  <body>
    <div
      style='background-color:#0B0F17;color:#ffffff;font-family:ui-rounded, "Hiragino Maru Gothic ProN", Quicksand, Comfortaa, Manjari, "Arial Rounded MT Bold", Calibri, source-sans-pro, sans-serif;font-size:16px;font-weight:400;letter-spacing:0.15008px;line-height:1.5;margin:0;padding:32px 0;min-height:100%;width:100%'
    >
      <table
        align="center"
        width="100%"
        style="margin:0 auto;max-width:600px;background-color:#0B0F17"
        role="presentation"
        cellspacing="0"
        cellpadding="0"
        border="0"
      >
        <tbody>
          <tr style="width:100%">
            <td>
              <div style="padding:16px 24px 16px 24px">
                <div
                  style="font-size:16px;text-align:center;padding:16px 24px 16px 24px"
                >
                  <svg
                    width="328.79999999999995"
                    height="102.8553846153846"
                    viewBox="0 0 390 122"
                    class="looka-1j8o68f"
                  >
                    <defs id="SvgjsDefs1011"></defs>
                    <g
                      id="SvgjsG1012"
                      featurekey="rootContainer"
                      transform="matrix(1,0,0,1,0,0)"
                      fill="#9334E9"
                    >
                      <rect
                        xmlns="http://www.w3.org/2000/svg"
                        width="390"
                        height="122"
                        rx="10"
                        ry="10"
                      ></rect>
                    </g>
                    <g
                      id="SvgjsG1013"
                      featurekey="q4o0QG-0"
                      transform="matrix(5.492781797784733,0,0,5.492781797784733,19.121154931998163,-8.562464300815627)"
                      fill="#09252c"
                    >
                      <path
                        d="M0.16 5.300000000000001 l0.6 -0.06 l6.18 0.06 l0.22 2.58 l-2.78 0.42 l0.52 11.7 l-3.02 -0.02 l0.08 -11.28 l-1.58 -0.14 z M8.120000000000001 19.92 l-0.04 -6.94 l0.08 -0.74 l-0.04 -0.74 l0.1 -4.42 l0.18 -1.72 l6.5 0.12 l0.1 3.9 l-4.26 0 l-0.08 2.5 l3.26 -0.06 l0 0.22 l0.1 1.7 l-3.04 0.06 l0.1 2.96 l4.06 0.2 l0 3.08 z M15.940000000000001 20 l0.24 -14.5 l2.38 -0.18 l1.08 0.12 l1.02 6.94 l0.6 -7.16 l4.1 0.5 c0.02 0.76 0.04 1.54 0.04 2.36 c0.08 1.38 0.18 3.18 0.18 4.84 c0.06 2.68 -0.02 4.6 0.16 7.12 l-1.76 -0.04 l-0.14 -8.98 l-0.38 -0.08 l-1 7.96 l-0.34 0.98 l-2.12 0.12 l-1.02 -8.4 l-0.6 0.14 l0.22 8.38 z M33.84 6.16 l0.38 2.7 l-0.02 0.08 c0 0.02 -0.08 0.96 -0.12 1.9 c-0.04 0.54 -0.08 1.18 -0.1 1.88 l-0.72 1.28 l-1.92 0.64 l-1.7 0.08 l0.28 3.36 l0.1 1.7 l-1.38 0.16 c-0.14 0 -0.38 0.02 -0.6 0.04 c-0.14 0.02 -0.28 0.02 -0.42 0.02 l-0.7 -0.12 l0 -3.04 l-0.12 -1.98 c-0.02 -0.42 -0.02 -0.8 -0.02 -1.14 c-0.02 -0.62 -0.02 -1.22 -0.02 -1.44 l0 -4.62 l0.08 -2.36 l0.78 0 c0.46 0 1 -0.02 1.5 -0.02 c0.98 0 2.02 -0.04 2.28 0.02 c0.22 0.06 0.66 0.1 1.04 0.16 l0.7 0.12 z M31.020000000000003 12.68 l0.3 -0.3 c0 -0.32 0.1 -2.02 0.1 -2.22 s-0.18 -0.9 -0.22 -1.1 l-0.48 -0.52 l-0.68 -0.16 l-1.36 0 l-0.12 0.34 l-0.08 1.44 l0.16 2.52 l0.8 0.16 z M38.74 5.199999999999999 l0 12.2 l1.5 0.28 l1.46 0.04 l-0.14 2.26 l-1.14 0 l-3.8 0.14 l-1.26 -0.04 l-0.26 -13.28 l0.12 -1.44 z M44.300000000000004 13.379999999999999 l2 -0.1 l-0.06 -2.78 l-0.5 -1.94 l-1.24 -0.1 z M44.84 5.199999999999999 l1.28 0 l2.62 0.18 l0.54 7.26 l0.2 7.24 l-0.6 0.04 l-2.26 0.1 l-0.24 0 l-0.18 -4.5 l-1.88 -0.12 l-0.5 4.6 l-1.96 0.06 l0.64 -7.48 l0.58 -4.28 l0.42 -3 z M49.32 5.359999999999999 l4.06 -0.12 l2.12 0.08 l0.6 0.04 l-0.22 2.9 l-1.6 -0.1 l0.12 11.84 l-3.54 0.04 l0.18 -7.04 l0.08 -4.88 l-1.94 -0.18 z M56.86 19.92 l-0.04 -6.94 l0.08 -0.74 l-0.04 -0.74 l0.1 -4.42 l0.18 -1.72 l6.5 0.12 l0.1 3.9 l-4.26 0 l-0.08 2.5 l3.26 -0.06 l0 0.22 l0.1 1.7 l-3.04 0.06 l0.1 2.96 l4.06 0.2 l0 3.08 z"
                      ></path>
                    </g>
                  </svg>
                </div>
              </div>
              <h3
                style="font-weight:bold;text-align:left;margin:0;font-size:20px;padding:32px 24px 0px 24px"
              >
                Reset your password?
              </h3>
              <div
                style="color:#828384;font-size:14px;font-weight:normal;text-align:left;padding:8px 24px 16px 24px"
              >
                If you didn&#x27;t request a reset, don&#x27;t worry. You can
                safely ignore this email.
              </div>
              <div style="text-align:left;padding:12px 24px 32px 24px">
                <a
                  href="${redirectUrl}?token=${accessToken}"
                  style="color:#FFFFFF;font-size:14px;font-weight:bold;background-color:#9333EA;border-radius:64px;display:inline-block;padding:12px 20px;text-decoration:none"
                  target="_blank"
                  ><span
                    ><!--[if mso
                      ]><i
                        style="letter-spacing: 20px;mso-font-width:-100%;mso-text-raise:30"
                        hidden
                        >&nbsp;</i
                      ><!
                    [endif]--></span
                  ><span>Reset my password</span
                  ><span
                    ><!--[if mso
                      ]><i
                        style="letter-spacing: 20px;mso-font-width:-100%"
                        hidden
                        >&nbsp;</i
                      ><!
                    [endif]--></span
                  ></a
                >
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </body>
</html>`,
    };
    await this.sgMail.send(msg);
  }
}
