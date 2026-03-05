package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"github.com/tormenta-bot/internal/assets"
	botruntime "github.com/tormenta-bot/internal/bot"
	"github.com/tormenta-bot/internal/database"
	"github.com/tormenta-bot/internal/handlers"
)

func main() {
	// Load .env — procura no diretório atual e no diretório do executável
	loaded := false
	for _, envPath := range []string{
		".env",
		"/root/tormenta-bot/.env",
		"/app/.env",
		func() string {
			exe, _ := os.Executable()
			return filepath.Join(filepath.Dir(exe), ".env")
		}(),
	} {
		if err := godotenv.Load(envPath); err == nil {
			log.Printf("✅ .env carregado de: %s", envPath)
			loaded = true
			break
		}
	}
	if !loaded {
		log.Println("⚠️  .env não encontrado — usando variáveis de ambiente do sistema")
	}
	log.Printf("🔑 GM_IDS=%q | TELEGRAM_TOKEN presente=%v",
		os.Getenv("GM_IDS"), os.Getenv("TELEGRAM_TOKEN") != "")

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	database.Migrate()

	// Initialize image asset manager
	assetsDir := os.Getenv("ASSETS_DIR")
	if assetsDir == "" {
		assetsDir = filepath.Join(".", "assets", "images")
	}
	imageCache := database.ImageCache{}
	if err := assets.Init(assetsDir, imageCache); err != nil {
		log.Printf("\u26a0\ufe0f  Image system init failed: %v \u2014 running in text-only mode", err)
	}

	// Initialize Telegram bot
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("\u274c TELEGRAM_TOKEN not set!")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("\u274c Failed to create bot: %v", err)
	}
	handlers.Bot = bot
	bot.Debug = false
	log.Printf("\u2705 Bot @%s started | Images: %s", bot.Self.UserName, assetsDir)

	// Start Pix polling goroutine (polls AbacatePay every 15s)
	if os.Getenv("ABACATEPAY_TOKEN") != "" {
		handlers.StartPixPolling()
		log.Println("\U0001f4b3 AbacatePay Pix polling started")
	}

	// Start VIP auto hunt worker (ticks every 5 minutes)
	handlers.StartAutoHuntWorker()
	handlers.StartEnergyRegenWorker()

	// Start HTTP server for Mercado Pago webhooks (optional, but recommended)
	webhookPort := os.Getenv("MP_WEBHOOK_PORT")
	if webhookPort == "" {
		webhookPort = "8080"
	}
	go startWebhookServer(webhookPort)

	// Poll for Telegram updates
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	workers := envInt("UPDATE_WORKERS", 1) // safe default preserves current behavior
	log.Printf("🧵 Update worker pool started with %d worker(s)", workers)
	botruntime.StartUpdateWorkerPool(updates, workers, botruntime.UpdateHandlers{
		OnMessage:  handlers.HandleMessage,
		OnCallback: handlers.HandleCallback,
	})
	select {}
}

func envInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return def
	}
	return n
}

// =============================================
// ABACATEPAY WEBHOOK HTTP SERVER
// =============================================
// AbacatePay envia POST /pix/webhook quando um pagamento é confirmado.

func startWebhookServer(port string) {
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// AbacatePay webhook
	mux.HandleFunc("/pix/webhook", handleAbacateWebhook)

	// Mantemos rota antiga por compatibilidade
	mux.HandleFunc("/mp/webhook", handleAbacateWebhook)

	addr := ":" + port
	log.Printf("\U0001f310 HTTP server listening on %s (AbacatePay webhook: /pix/webhook)", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Printf("\u26a0\ufe0f  HTTP server error: %v", err)
	}
}

// abacateWebhookPayload é o body que o AbacatePay envia nos eventos billing.paid.
// O campo "data" contém os produtos cujo externalId é o txID do nosso pagamento.
type abacateWebhookPayload struct {
	Event string `json:"event"` // "billing.paid"
	Data  struct {
		ID       string `json:"id"`     // ID da cobrança (billing), ex: "bill_abc123"
		Status   string `json:"status"` // "PAID"
		Products []struct {
			ExternalID string `json:"externalId"` // nosso txID
		} `json:"products"`
	} `json:"data"`
}

func handleAbacateWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusOK)
		return
	}
	defer r.Body.Close()

	body, err := io.ReadAll(io.LimitReader(r.Body, 64*1024))
	if err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	var payload abacateWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	log.Printf("[AbacatePay Webhook] Event=%s ID=%s Status=%s", payload.Event, payload.Data.ID, payload.Data.Status)

	// Só processa billing.paid
	if payload.Event != "billing.paid" || payload.Data.Status != "PAID" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
		return
	}

	// O ID para confirmar é o ID do pixQrCode (TxID guardado no banco)
	// O AbacatePay envia o billing ID — usamos ele direto pois guardamos como TxID
	abacateID := payload.Data.ID

	// Processa de forma assíncrona para responder rápido ao AbacatePay
	go func(id string) {
		playerID, diamonds, err := handlers.HandleAbacateWebhookNotification(id)
		if err != nil {
			log.Printf("[AbacatePay Webhook] Error processing %s: %v", id, err)
			return
		}
		if diamonds > 0 {
			log.Printf("[AbacatePay Webhook] Credited %d diamonds, playerID %d", diamonds, playerID)
		}
	}(abacateID)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
