package game

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tormenta-bot/internal/models"
)

// ── PACOTES DE DIAMANTES ──────────────────────────────────

var DiamondPackages = []models.DiamondPackage{
	{ID: "pkg_30", Name: "Saquinho de Cristal", Emoji: "💎", Amount: 30, Bonus: 0, PriceBRL: 4.99, Price: "R$ 4,99"},
	{ID: "pkg_80", Name: "Bolsa de Gemas", Emoji: "💎", Amount: 80, Bonus: 5, PriceBRL: 9.99, Price: "R$ 9,99"},
	{ID: "pkg_180", Name: "Baú de Diamantes", Emoji: "💎", Amount: 180, Bonus: 20, PriceBRL: 19.99, Price: "R$ 19,99"},
	{ID: "pkg_400", Name: "Tesouro do Dragão", Emoji: "💎", Amount: 400, Bonus: 50, PriceBRL: 39.99, Price: "R$ 39,99"},
}

func GetDiamondPackage(id string) *models.DiamondPackage {
	for i := range DiamondPackages {
		if DiamondPackages[i].ID == id {
			return &DiamondPackages[i]
		}
	}
	return nil
}

// ── ABACATEPAY PIX ────────────────────────────────────────

const abacateBaseURL = "https://api.abacatepay.com"

// MPPaymentResult é retornado para os handlers após criar o pagamento.
// Mantemos o mesmo nome para não precisar alterar handlers_pix.go.
type MPPaymentResult struct {
	PaymentID int64  // não usado no AbacatePay, mantido por compatibilidade
	QRCode    string // copia e cola (BRCode)
	QRCodeB64 string // base64 PNG (baixado da qrCodeUrl do AbacatePay)
	QRCodeURL string // URL original da imagem QR (para fallback)
	TxID      string // ID do pixQrCode no AbacatePay (ex: "pix_abc123")
	ExpiresAt time.Time
}

// abacateCreateRequest é o body para POST /v1/pixQrCode/create
type abacateCreateRequest struct {
	Amount      int    `json:"amount"`    // em centavos
	ExpiresIn   int    `json:"expiresIn"` // segundos até expirar
	Description string `json:"description"`
}

// abacateCreateResponse é a resposta do AbacatePay ao criar QR code PIX
type abacateCreateResponse struct {
	Data struct {
		ID        string `json:"id"` // ID do pixQrCode, ex: "pix_abc123"
		Status    string `json:"status"`
		BRCode    string `json:"brCode"` // copia e cola
		QRCodeURL string `json:"qrCodeUrl"`
		ExpiresAt string `json:"expiresAt"`
		DevMode   bool   `json:"devMode"`
	} `json:"data"`
	Error interface{} `json:"error"`
}

// abacateStatusResponse é a resposta de GET /v1/pixQrCode/check?id=...
type abacateStatusResponse struct {
	Data struct {
		Status    string `json:"status"` // PENDING, PAID, EXPIRED, CANCELLED
		ExpiresAt string `json:"expiresAt"`
	} `json:"data"`
	Error interface{} `json:"error"`
}

// CreateMPPixPayment cria um QR code PIX via AbacatePay.
// Requer env: ABACATEPAY_TOKEN
func CreateMPPixPayment(pkg *models.DiamondPackage, charID int) (*MPPaymentResult, error) {
	token := os.Getenv("ABACATEPAY_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("ABACATEPAY_TOKEN não configurado")
	}

	txID := GenerateTxID()
	total := pkg.Amount + pkg.Bonus
	desc := fmt.Sprintf("Tormenta RPG - %s (%d diamantes) char%d", pkg.Name, total, charID)

	// Preço em centavos (AbacatePay usa centavos)
	amountCents := int(math.Round(pkg.PriceBRL * 100))

	// Expira em 30 minutos = 1800 segundos
	expiresInSec := 1800

	body := abacateCreateRequest{
		Amount:      amountCents,
		ExpiresIn:   expiresInSec,
		Description: desc,
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest("POST", abacateBaseURL+"/v1/pixQrCode/create", bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("abacatepay api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("abacatepay status %d: %s", resp.StatusCode, string(respBody))
	}

	var apResp abacateCreateResponse
	if err := json.Unmarshal(respBody, &apResp); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	if apResp.Data.BRCode == "" {
		return nil, fmt.Errorf("abacatepay response missing brCode: %s", string(respBody))
	}

	brt := time.FixedZone("BRT", -3*3600)
	expiresAt := time.Now().In(brt).Add(30 * time.Minute)

	abacateID := apResp.Data.ID
	if abacateID == "" {
		abacateID = "abacate_" + txID
	}

	// Tenta baixar o QR code PNG da qrCodeUrl e encodar em base64
	qrB64 := ""
	qrURL := apResp.Data.QRCodeURL
	if qrURL != "" {
		if imgData, err2 := fetchQRCodeImage(qrURL); err2 == nil {
			qrB64 = base64.StdEncoding.EncodeToString(imgData)
		}
	}

	return &MPPaymentResult{
		PaymentID: 0,
		QRCode:    apResp.Data.BRCode,
		QRCodeB64: qrB64,
		QRCodeURL: qrURL,
		TxID:      abacateID,
		ExpiresAt: expiresAt,
	}, nil
}

// CheckMPPaymentStatus mantido por compatibilidade.
func CheckMPPaymentStatus(mpPaymentID int64) (string, error) {
	return "pending", nil
}

// CheckAbacatePayStatus consulta o status pelo ID string do AbacatePay.
func CheckAbacatePayStatus(abacateID string) (string, error) {
	token := os.Getenv("ABACATEPAY_TOKEN")
	if token == "" {
		return "", fmt.Errorf("ABACATEPAY_TOKEN não configurado")
	}

	url := fmt.Sprintf("%s/v1/pixQrCode/check?id=%s", abacateBaseURL, abacateID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var apResp abacateStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&apResp); err != nil {
		return "", fmt.Errorf("abacatepay decode status response: %w", err)
	}

	// Normaliza status AbacatePay → formato usado pelo sistema
	switch strings.ToUpper(apResp.Data.Status) {
	case "PAID":
		return "approved", nil
	case "EXPIRED", "CANCELLED":
		return "cancelled", nil
	default:
		return "pending", nil
	}
}

// ── HELPERS ───────────────────────────────────────────────

// GenerateTxID retorna uma string hex aleatória de 25 caracteres.
func GenerateTxID() string {
	b := make([]byte, 13)
	if _, err := rand.Read(b); err != nil {
		// Fallback (muito improvável), mas garante um ID sempre retornado.
		fb := hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
		return strings.ToUpper(fb)[:25]
	}
	return strings.ToUpper(hex.EncodeToString(b))[:25]
}

// fetchQRCodeImage baixa a imagem PNG do QR code a partir da URL fornecida pela AbacatePay.
func fetchQRCodeImage(url string) ([]byte, error) {
	return FetchQRCodeImagePublic(url)
}

// FetchQRCodeImagePublic é a versão exportada para uso nos handlers.
func FetchQRCodeImagePublic(url string) ([]byte, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("fetchQRCodeImage GET: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetchQRCodeImage status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024)) // max 2MB
	if err != nil {
		return nil, fmt.Errorf("fetchQRCodeImage read: %w", err)
	}
	return data, nil
}

// FormatPixCode quebra o código PIX em linhas de 60 chars para melhor leitura.
func FormatPixCode(code string) string {
	if len(code) <= 60 {
		return code
	}
	var parts []string
	for i := 0; i < len(code); i += 60 {
		end := i + 60
		if end > len(code) {
			end = len(code)
		}
		parts = append(parts, code[i:end])
	}
	return strings.Join(parts, "\n")
}
