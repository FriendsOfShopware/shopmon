package mail

// shopmonTheme implements hermes.Theme with a modern, high-contrast visual layout.
type shopmonTheme struct{}

func (t *shopmonTheme) Name() string {
	return "shopmon"
}

func (t *shopmonTheme) HTMLTemplate() string {
	return `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" dir="{{ if .Hermes.TextDirection }}{{ .Hermes.TextDirection }}{{ else }}ltr{{ end }}">
<head>
  <meta name="viewport" content="width=device-width, initial-scale=1.0" />
  <meta http-equiv="Content-Type" content="text/html; charset=UTF-8" />
  <style type="text/css">
    body {
      margin: 0;
      padding: 0;
      width: 100% !important;
      -webkit-text-size-adjust: 100%;
      -ms-text-size-adjust: 100%;
      background-color: #f8fafc;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif;
    }
    p {
      margin-top: 0;
      margin-bottom: 12px;
      line-height: 1.6;
    }
    ul, ol {
      margin-top: 8px;
      margin-bottom: 16px;
      padding-left: 20px;
    }
    li {
      margin-bottom: 8px;
      line-height: 1.5;
    }
    a {
      color: #2563eb;
      text-decoration: underline;
    }
    code {
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
      font-size: 13px;
      background-color: #f1f5f9;
      padding: 2px 6px;
      border-radius: 4px;
      color: #0f172a;
    }
    strong {
      color: #0f172a;
      font-weight: 600;
    }
  </style>
</head>
<body style="margin: 0; padding: 0; width: 100%; background-color: #f8fafc; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; -webkit-font-smoothing: antialiased;">
  <table class="email-wrapper" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; margin: 0; padding: 32px 16px; background-color: #f8fafc;">
    <tr>
      <td align="center">
        <table class="email-content" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; max-width: 580px; margin: 0 auto; padding: 0;">
          
          <!-- Logo Header -->
          <tr>
            <td class="email-header" style="padding: 0 0 24px 0; text-align: center;">
              <a href="{{ .Hermes.Product.Link }}" target="_blank" style="display: inline-block; text-decoration: none;">
                {{ if .Hermes.Product.Logo }}
                  <img src="{{ .Hermes.Product.Logo }}" alt="{{ .Hermes.Product.Name }}" height="44" style="height: 44px; max-height: 44px; border: 0; vertical-align: middle;" />
                {{ else }}
                  <span style="font-size: 22px; font-weight: 700; color: #0f172a; letter-spacing: -0.5px;">{{ .Hermes.Product.Name }}</span>
                {{ end }}
              </a>
            </td>
          </tr>

          <!-- Main Card -->
          <tr>
            <td class="email-body" width="100%" style="width: 100%; margin: 0; padding: 0; background-color: #ffffff; border-radius: 12px; border: 1px solid #e2e8f0;">
              <!-- Top Accent Bar -->
              <div style="height: 4px; background-color: #2563eb; border-top-left-radius: 11px; border-top-right-radius: 11px;"></div>
              
              <table class="email-body_inner" align="center" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; margin: 0 auto; padding: 32px 28px;">
                <tr>
                  <td class="content-cell" style="color: #334155; font-size: 15px; line-height: 1.6;">
                    <!-- Title / Greeting -->
                    {{ if .Email.Body.Title }}
                      <h1 style="margin: 0 0 20px 0; color: #0f172a; font-size: 20px; font-weight: 700; line-height: 1.3; letter-spacing: -0.3px;">{{ .Email.Body.Title | safe }}</h1>
                    {{ else if .Email.Body.Name }}
                      <h1 style="margin: 0 0 20px 0; color: #0f172a; font-size: 19px; font-weight: 700; line-height: 1.3; letter-spacing: -0.3px;">
                        {{ if .Email.Body.Greeting }}{{ .Email.Body.Greeting }}{{ else }}Hi{{ end }} {{ .Email.Body.Name }},
                      </h1>
                    {{ end }}

                    <!-- Intros -->
                    {{ range $line := .Email.Body.Intros }}
                      <div style="margin: 0 0 16px 0; color: #334155; font-size: 15px; line-height: 1.6;">
                        {{ $line | safe }}
                      </div>
                    {{ end }}

                    <!-- FreeMarkdown -->
                    {{ if .Email.Body.FreeMarkdown }}
                      <div style="margin: 16px 0; color: #334155; font-size: 15px; line-height: 1.6;">
                        {{ .Email.Body.FreeMarkdown | safe }}
                      </div>
                    {{ end }}

                    <!-- Dictionary (Key-Value Grid) -->
                    {{ if .Email.Body.Dictionary }}
                      <table class="dictionary" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; margin: 20px 0; border: 1px solid #f1f5f9; border-radius: 8px; background-color: #f8fafc; overflow: hidden;">
                        {{ range $entry := .Email.Body.Dictionary }}
                          <tr>
                            <td style="padding: 10px 16px; font-size: 13px; font-weight: 600; color: #64748b; border-bottom: 1px solid #f1f5f9; width: 35%;">{{ $entry.Key }}</td>
                            <td style="padding: 10px 16px; font-size: 14px; color: #0f172a; border-bottom: 1px solid #f1f5f9;">{{ $entry.Value | safe }}</td>
                          </tr>
                        {{ end }}
                      </table>
                    {{ end }}

                    <!-- Table -->
                    {{ if .Email.Body.Table.Data }}
                      <table class="data-table" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; margin: 20px 0; border: 1px solid #e2e8f0; border-radius: 8px; overflow: hidden; border-collapse: collapse;">
                        {{ if .Email.Body.Table.Columns.Header }}
                          <tr style="background-color: #f8fafc;">
                            {{ range $cell := .Email.Body.Table.Columns.Header }}
                              <th style="padding: 10px 14px; font-size: 12px; font-weight: 700; text-transform: uppercase; color: #64748b; text-align: left; border-bottom: 1px solid #e2e8f0;">{{ $cell }}</th>
                            {{ end }}
                          </tr>
                        {{ end }}
                        {{ range $row := .Email.Body.Table.Data }}
                          <tr style="border-bottom: 1px solid #f1f5f9;">
                            {{ range $cell := $row }}
                              <td style="padding: 12px 14px; font-size: 14px; color: #334155;">{{ $cell | safe }}</td>
                            {{ end }}
                          </tr>
                        {{ end }}
                      </table>
                    {{ end }}

                    <!-- Actions / Buttons -->
                    {{ if .Email.Body.Actions }}
                      <table class="body-action" align="center" width="100%" cellpadding="0" cellspacing="0" role="presentation" style="width: 100%; margin: 28px 0; text-align: center;">
                        <tr>
                          <td align="center">
                            {{ range $action := .Email.Body.Actions }}
                              <a href="{{ $action.Button.Link }}" target="_blank" style="display: inline-block; padding: 12px 24px; margin: 6px; font-size: 14px; font-weight: 600; text-decoration: none; border-radius: 8px; color: #ffffff; background-color: {{ if $action.Button.Color }}{{ $action.Button.Color }}{{ else }}#2563eb{{ end }}; box-shadow: 0 2px 4px rgba(37, 99, 235, 0.15);">
                                {{ $action.Button.Text }}
                              </a>
                            {{ end }}
                          </td>
                        </tr>
                      </table>
                    {{ end }}

                    <!-- Outros -->
                    {{ range $line := .Email.Body.Outros }}
                      <div style="margin: 16px 0 0 0; color: #64748b; font-size: 14px; line-height: 1.5;">
                        {{ $line | safe }}
                      </div>
                    {{ end }}

                    <!-- Signature -->
                    <div style="margin-top: 28px; padding-top: 20px; border-top: 1px solid #f1f5f9; color: #475569; font-size: 14px; line-height: 1.5;">
                      {{ if .Email.Body.Signature }}
                        {{ .Email.Body.Signature }}
                      {{ else }}
                        Best regards,
                      {{ end }}
                      <br />
                      <strong style="color: #0f172a; font-weight: 600;">{{ .Hermes.Product.Name }} Team</strong>
                    </div>

                    <!-- Sub / Trouble with button -->
                    {{ if .Email.Body.Actions }}
                      {{ range $action := .Email.Body.Actions }}
                        <table class="body-sub" style="width: 100%; margin-top: 24px; padding-top: 20px; border-top: 1px solid #f1f5f9;">
                          <tr>
                            <td>
                              <p style="margin: 0 0 6px 0; color: #94a3b8; font-size: 12px; line-height: 1.5;">
                                If you're having trouble with the "{{ $action.Button.Text }}" button, copy and paste the URL below into your browser:
                              </p>
                              <p style="margin: 0; font-size: 12px; line-height: 1.5;">
                                <a href="{{ $action.Button.Link }}" style="color: #2563eb; word-break: break-all; text-decoration: underline;">{{ $action.Button.Link }}</a>
                              </p>
                            </td>
                          </tr>
                        </table>
                      {{ end }}
                    {{ end }}

                  </td>
                </tr>
              </table>
            </td>
          </tr>

          <!-- Footer -->
          <tr>
            <td style="padding: 24px 0 0 0; text-align: center;">
              <p style="margin: 0 0 8px 0; color: #94a3b8; font-size: 12px; line-height: 1.5;">
                {{ .Hermes.Product.Copyright }}
              </p>
            </td>
          </tr>

        </table>
      </td>
    </tr>
  </table>
</body>
</html>
`
}

func (t *shopmonTheme) PlainTextTemplate() string {
	return `{{ if .Email.Body.Title }}{{ .Email.Body.Title }}
{{ else if .Email.Body.Name }}{{ if .Email.Body.Greeting }}{{ .Email.Body.Greeting }}{{ else }}Hi{{ end }} {{ .Email.Body.Name }},
{{ end }}
{{ range $line := .Email.Body.Intros }}
{{ $line }}
{{ end }}
{{ if .Email.Body.FreeMarkdown }}{{ .Email.Body.FreeMarkdown }}{{ end }}
{{ if .Email.Body.Dictionary }}
{{ range $entry := .Email.Body.Dictionary }}{{ $entry.Key }}: {{ $entry.Value }}
{{ end }}
{{ end }}
{{ if .Email.Body.Table.Data }}
{{ range $row := .Email.Body.Table.Data }}{{ range $cell := $row }}{{ $cell }} | {{ end }}
{{ end }}
{{ end }}
{{ if .Email.Body.Actions }}
{{ range $action := .Email.Body.Actions }}
{{ $action.Button.Text }}: {{ $action.Button.Link }}
{{ end }}
{{ end }}
{{ range $line := .Email.Body.Outros }}
{{ $line }}
{{ end }}
{{ if .Email.Body.Signature }}{{ .Email.Body.Signature }}{{ else }}Best regards,{{ end }}
{{ .Hermes.Product.Name }} Team
`
}
