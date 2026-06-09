package notification

import (
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/abdulwahidkahar/notification-service/internal/model"
)

type EmailContent struct {
	Subject string
	Body    string
}

type emailTemplateData struct {
	Subject       string
	Title         string
	Subtitle      string
	StatusLabel   string
	StatusColor   string
	Amount        string
	UserID        string
	EventID       string
	CreatedAt     string
	Transaction   string
	PrimaryAction string
}

func BuildEmailContent(event model.NotificationEvent) (EmailContent, error) {
	data := emailTemplateData{
		UserID:        html.EscapeString(event.UserID),
		EventID:       html.EscapeString(event.EventID),
		CreatedAt:     formatEmailTime(event.CreatedAt),
		Amount:        formatRupiah(event.Amount),
		StatusColor:   "#0064D2",
		PrimaryAction: "Lihat detail transaksi",
	}

	switch event.Type {
	case model.EventTransferSuccess:
		data.Subject = "Transfer Berhasil"
		data.Title = "Transfer berhasil"
		data.Subtitle = "Transaksi kamu sudah berhasil diproses."
		data.StatusLabel = "Berhasil"
		data.Transaction = "Transfer"
	case model.EventTopUpConfirmed:
		data.Subject = "Top Up Dikonfirmasi"
		data.Title = "Top up dikonfirmasi"
		data.Subtitle = "Saldo sudah berhasil ditambahkan."
		data.StatusLabel = "Dikonfirmasi"
		data.Transaction = "Top up"
	default:
		return EmailContent{}, fmt.Errorf("unknown event type: %s", event.Type)
	}

	return EmailContent{
		Subject: data.Subject,
		Body:    buildHTML(data),
	}, nil
}

func buildHTML(data emailTemplateData) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="id">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s</title>
</head>
<body style="margin:0;padding:0;background-color:#F5F7FA;font-family:Arial,Helvetica,sans-serif;color:#1F2937;">
  <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="background-color:#F5F7FA;margin:0;padding:24px 12px;">
    <tr>
      <td align="center">
        <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="max-width:560px;background-color:#FFFFFF;border-radius:12px;overflow:hidden;border:1px solid #E6EAF0;">
          <tr>
            <td style="height:5px;background-color:#FFCB05;font-size:0;line-height:0;">&nbsp;</td>
          </tr>
          <tr>
            <td style="padding:22px 24px 16px 24px;">
              <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0">
                <tr>
                  <td style="font-size:20px;font-weight:700;color:#0064D2;line-height:1.2;">notif<span style="color:#FFCB05;">pay</span></td>
                  <td align="right">
                    <span style="display:inline-block;background-color:#EAF4FF;color:%s;border-radius:999px;padding:6px 10px;font-size:12px;font-weight:700;">%s</span>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td style="padding:2px 24px 20px 24px;">
              <h1 style="margin:0 0 8px 0;font-size:22px;line-height:1.3;color:#111827;font-weight:700;">%s</h1>
              <p style="margin:0;font-size:14px;line-height:1.6;color:#6B7280;">%s</p>
            </td>
          </tr>
          <tr>
            <td style="padding:0 24px 8px 24px;">
              <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0" style="border-collapse:separate;border-spacing:0;background-color:#FFFFFF;border:1px solid #E6EAF0;border-radius:10px;overflow:hidden;">
                <tr>
                  <td style="padding:18px 18px 6px 18px;font-size:12px;color:#6B7280;">Total transaksi</td>
                </tr>
                <tr>
                  <td style="padding:0 18px 18px 18px;font-size:28px;line-height:1.15;font-weight:700;color:#111827;">%s</td>
                </tr>
                <tr>
                  <td style="padding:0 18px 6px 18px;">
                    <table role="presentation" width="100%%" cellspacing="0" cellpadding="0" border="0">
                      <tr>
                        <td style="padding:11px 0;border-top:1px solid #EEF2F7;font-size:13px;color:#6B7280;">Jenis transaksi</td>
                        <td align="right" style="padding:11px 0;border-top:1px solid #EEF2F7;font-size:13px;font-weight:700;color:#111827;">%s</td>
                      </tr>
                      <tr>
                        <td style="padding:11px 0;border-top:1px solid #EEF2F7;font-size:13px;color:#6B7280;">User ID</td>
                        <td align="right" style="padding:11px 0;border-top:1px solid #EEF2F7;font-size:13px;font-weight:700;color:#111827;">%s</td>
                      </tr>
                      <tr>
                        <td style="padding:11px 0;border-top:1px solid #EEF2F7;font-size:13px;color:#6B7280;">Waktu</td>
                        <td align="right" style="padding:11px 0;border-top:1px solid #EEF2F7;font-size:13px;font-weight:700;color:#111827;">%s</td>
                      </tr>
                      <tr>
                        <td style="padding:11px 0;border-top:1px solid #EEF2F7;font-size:13px;color:#6B7280;">Event ID</td>
                        <td align="right" style="padding:11px 0;border-top:1px solid #EEF2F7;font-size:12px;font-weight:700;color:#111827;">%s</td>
                      </tr>
                    </table>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          <tr>
            <td style="padding:14px 24px 24px 24px;">
              <a href="#" style="display:block;text-align:center;background-color:#0064D2;color:#FFFFFF;text-decoration:none;border-radius:8px;padding:13px 16px;font-size:14px;font-weight:700;">%s</a>
              <p style="margin:16px 0 0 0;font-size:12px;line-height:1.6;color:#8A94A6;text-align:center;">Email otomatis dari notification-service. Abaikan jika transaksi sudah sesuai.</p>
            </td>
          </tr>
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`,
		html.EscapeString(data.Subject),
		data.StatusColor,
		html.EscapeString(data.StatusLabel),
		html.EscapeString(data.Title),
		html.EscapeString(data.Subtitle),
		data.Amount,
		html.EscapeString(data.Transaction),
		data.UserID,
		data.CreatedAt,
		data.EventID,
		html.EscapeString(data.PrimaryAction),
	)
}

func formatRupiah(amount int64) string {
	if amount == 0 {
		return "Rp0"
	}

	isNegative := amount < 0
	if isNegative {
		amount = -amount
	}

	raw := fmt.Sprintf("%d", amount)
	parts := make([]string, 0, len(raw)/3+1)
	for len(raw) > 3 {
		parts = append([]string{raw[len(raw)-3:]}, parts...)
		raw = raw[:len(raw)-3]
	}
	parts = append([]string{raw}, parts...)

	prefix := "Rp"
	if isNegative {
		prefix = "-Rp"
	}

	return prefix + strings.Join(parts, ".")
}

func formatEmailTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}

	return t.Local().Format("02 Jan 2006 15:04")
}
