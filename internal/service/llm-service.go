package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/invopop/jsonschema"
	"github.com/labstack/gommon/log"
	"github.com/markojerkic/spring-planing/internal/database"
	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/option"
	"gorm.io/gorm"
)

type LLMService struct {
	openRouterClient *openai.Client
	requestChan      chan LLMRequest
	db               *database.Database
	webSocketService *WebSocketService
}

type LLMRequest struct {
	TicketKey   string
	Description string
	RoomID      uint
	TicketID    uint
	RetryCount  uint
}

type RecommendedEstimate struct {
	WeekEstimate int32 `json:"weekEstimate" form:"weekEstimate" default:"0" jsonschema_description:"Estimate for the ticket in weeks. Minimum is 0."`
	DayEstimate  int32 `json:"dayEstimate" form:"dayEstimate" default:"0" jsonschema_description:"Estimate for the ticket in days. Minimum is 0, maximum is 7."`
	HourEstimate int32 `json:"hourEstimate" form:"hourEstimate" default:"0" jsonschema_description:"Estimate for the ticket in hours. Minimum is 0, maximum is 8."`
}

func GenerateSchema[T any]() any {
	// Structured Outputs uses a subset of JSON schema
	// These flags are necessary to comply with the subset
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	var v T
	schema := reflector.Reflect(v)
	return schema
}

var RecommendedEstimateSchema = GenerateSchema[RecommendedEstimate]()

func (l *LLMService) processRequests() {
	for req := range l.requestChan {
		log.Debug("Processing LLM request", "ticket", req.TicketKey, "description", req.Description)

		var isLLMEnabled bool
		if err := l.db.DB.WithContext(context.Background()).
			Model(&database.Room{}).
			Select("allow_llm_estimation").
			Where("id = ?", req.RoomID).
			Scan(&isLLMEnabled).Error; err != nil {
			slog.Error("Error reading room", "error", err)
			return
		}

		if !isLLMEnabled {
			slog.Debug("LLM estimation is disabled for room", "room", req.RoomID)
			return
		}

		slog.Debug("Processing LLM request", "ticket", req.TicketKey, "description", req.Description)

		llmCtx, cancelLlm := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancelLlm()

		estimate, err := l.generateEstimate(llmCtx, req.TicketKey, req.Description)
		if err != nil {
			slog.Error("Error generating estimate", "ticket", req.TicketKey, "error", err)
			if req.RetryCount < 3 {
				slog.Debug("Retrying LLM request", "ticket", req.TicketKey, "description", req.Description, "retryCount", req.RetryCount)
				go func() {
					l.requestChan <- LLMRequest{
						TicketKey:   req.TicketKey,
						Description: req.Description,
						RoomID:      req.RoomID,
						TicketID:    req.TicketID,
						RetryCount:  req.RetryCount + 1,
					}
				}()
			}
			return
		}

		timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = l.db.DB.WithContext(timeoutCtx).Transaction(func(tx *gorm.DB) error {

			estimate := database.Estimate{
				TicketID: req.TicketID,
				Estimate: int(estimate.WeekEstimate*5*8 + estimate.DayEstimate*8 + estimate.HourEstimate),
			}
			if err := tx.Create(&estimate).Error; err != nil {
				return err
			}
			slog.Debug("Estimate saved", "estimate", estimate.ID)

			if err := tx.Model(&database.Ticket{}).
				Where("id = ?", req.TicketID).
				Update("llm_estimate_id", estimate.ID).Error; err != nil {
				return err
			}

			return nil
		})
		if err != nil {
			slog.Error("Error saving estimate", "error", err)
		}

		formatedEstimate := fmt.Sprintf("LLM estimate: %dw %dd %dh", estimate.WeekEstimate, estimate.DayEstimate, estimate.HourEstimate)

		go func() {
			l.webSocketService.SendLLMRecommendation(req.TicketID, &req.TicketKey, req.RoomID, formatedEstimate)
		}()
	}
}

func (l *LLMService) generateEstimate(ctx context.Context, ticketKey string, ticketDescription string) (RecommendedEstimate, error) {
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "ticket_estimate",
		Description: openai.String("Estimate for the required work for a Jira ticket."),
		Schema:      RecommendedEstimateSchema,
		Strict:      openai.Bool(true),
	}

	systemPrompt := `You are a ticket recommender. You will be given a Jira ticket key and description.
	You will be asked to estimate the work required for the ticket.
	Your response should be a JSON object with the following keys:

- weekEstimate: Estimate for the ticket in weeks. Minimum is 0.
- dayEstimate: Estimate for the ticket in days. Minimum is 0, maximum is 7.
- hourEstimate: Estimate for the ticket in hours. Minimum is 0, maximum is 8.

You will be given the following information:

- ticketKey: The Jira ticket key.
- ticketDescription: The Jira ticket description.

You will be asked to estimate the work required for the ticket. Your response should be a JSON object with the following keys:

- weekEstimate: Estimate for the ticket in weeks. Minimum is 0.
- dayEstimate: Estimate for the ticket in days. Minimum is 0, maximum is 7.
- hourEstimate: Estimate for the ticket in hours. Minimum is 0, maximum is 8.
`

	chat, err := l.openRouterClient.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(fmt.Sprintf(`Ticket key: %s.
				Ticket description: %s
				`, ticketKey, ticketDescription)),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: schemaParam},
		},
		Model: "google/gemini-2.5-flash-lite",
	})

	if err != nil {
		slog.Error("Error generating recomendation", "error", err)
		return RecommendedEstimate{}, err
	}

	var estimateRecommendation RecommendedEstimate
	err = json.Unmarshal([]byte(chat.Choices[0].Message.Content), &estimateRecommendation)
	if err != nil {
		slog.Error("Error parsing generated recomendation", "error", err)
		return RecommendedEstimate{}, err
	}

	slog.Debug("Generated estimate recommendation", "estimate", estimateRecommendation)

	return estimateRecommendation, nil
}

func NewLLMService(webSocketService *WebSocketService, db *database.Database) *LLMService {
	if webSocketService == nil {
		panic("webSocketService cannot be nil")
	}
	if db == nil {
		panic("db cannot be nil")
	}

	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENROUTER_API_KEY")),
		option.WithBaseURL(os.Getenv("OPENROUTER_BASE_URL")),
	)

	service := &LLMService{
		openRouterClient: &client,
		requestChan:      make(chan LLMRequest, 100),
		webSocketService: webSocketService,
		db:               db,
	}

	for range 10 {
		go service.processRequests()
	}

	return service
}

func (l *LLMService) GetRequestChannel() chan<- LLMRequest {
	return l.requestChan
}
