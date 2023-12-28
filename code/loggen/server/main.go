package main

import (
	"context"
	"embed"
	_ "embed"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/redis/go-redis/v9"
)

//go:embed views/*
var viewsfs embed.FS

func main() {
	ctx := context.Background()

	logJson := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	slog.SetDefault(slog.New(logJson))

	Port := 4000 // default port
	{
		port := strings.TrimSpace(os.Getenv("PORT"))
		if port == "" {
			slog.ErrorContext(ctx, "port is empty in environment variable")
			return
		}

		var err error
		Port, err = strconv.Atoi(port)
		if err != nil {
			slog.ErrorContext(ctx, "port is not valid number", slog.Any("error", err))
			return
		}
	}

	RedisAddr := "redis-service.app-engprod:6379" // default redis address
	{
		redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDRESS"))
		if redisAddr != "" {
			RedisAddr = redisAddr
		}

	}

	rdb := redis.NewClient(&redis.Options{
		Addr: RedisAddr,
		DB:   0,
	})

	defer func() {
		if rdb == nil {
			return
		}
		if _err := rdb.Close(); _err != nil {
			slog.ErrorContext(ctx, "cannot close redis", slog.Any("error", _err))
		}
	}()

	if rdb == nil {
		slog.ErrorContext(ctx, "redis client is nil")
		return
	}

	if _err := rdb.Ping(ctx).Err(); _err != nil {
		slog.ErrorContext(ctx, "redis client ping error", slog.Any("error", _err))
		return
	}

	genLogHandler := &GenerateLogHandler{
		RedisClient: rdb,
	}

	engine := html.NewFileSystem(http.FS(viewsfs), ".html")
	engine.AddFunc(
		// add unescape function
		"unescape", func(s string) template.JS {
			return template.JS(s)
		},
	)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		Views:                 engine,
	})

	app.Get("/ready", func(c *fiber.Ctx) error {
		return c.JSON(map[string]any{
			"ok": true,
		})
	})
	app.Get("/statistic/:metricPrefix/:traceID", genLogHandler.Statistic)
	app.Get("/statistic-report/:metricPrefix/:traceID", genLogHandler.StatisticHTML)

	var apiErrChan = make(chan error, 1)
	go func() {
		slog.InfoContext(ctx, fmt.Sprintf("server start at port %d", Port))
		err := app.Listen(fmt.Sprintf(":%d", Port))
		if err != nil {
			apiErrChan <- fmt.Errorf("cannot start server")
		}

	}()

	// ** listen for sigterm signal
	var signalChan = make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalChan:
		slog.InfoContext(ctx, "http transport: exiting")
		if _err := app.ShutdownWithContext(ctx); _err != nil {
			slog.ErrorContext(ctx, "http transport error", slog.Any("error", _err))
		}

	case err := <-apiErrChan:
		if err != nil {
			slog.ErrorContext(ctx, "http server error", slog.Any("error", err))
		}
	}
}

type GenerateLogHandler struct {
	RedisClient *redis.Client
}

func (h *GenerateLogHandler) Statistic(c *fiber.Ctx) error {
	var ctx context.Context = c.Context()
	metricPrefix := strings.TrimSpace(c.Params("metricPrefix"))
	traceID := strings.TrimSpace(c.Params("traceID"))

	resp, err := h.GetMetric(ctx, metricPrefix, traceID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

func (h *GenerateLogHandler) StatisticHTML(c *fiber.Ctx) error {
	var ctx context.Context = c.Context()
	metricPrefix := strings.TrimSpace(c.Params("metricPrefix"))
	traceID := strings.TrimSpace(c.Params("traceID"))

	resp, err := h.GetMetric(ctx, metricPrefix, traceID)
	if err != nil {
		return err
	}

	labels := make([]string, 0)
	bpsData := make([]string, 0)
	logLineData := make([]string, 0)
	for _, bps := range resp.BytesPerSecond {
		labels = append(labels, fmt.Sprintf(`"%s"`, bps.Date))
		bpsData = append(bpsData, fmt.Sprintf("%d", bps.TotalBytes))
		logLineData = append(logLineData, fmt.Sprintf("%d", bps.TotalLines))
	}

	return c.Render("views/statistic", fiber.Map{
		"LABELS":               fmt.Sprintf("[%s]", strings.Join(labels, ", ")),
		"TOTAL_BYTES":          fmt.Sprintf("%d", resp.TotalBytes),
		"TOTAL_BYTES_HUMANIZE": resp.TotalBytesHumanize,
		"BPS_DATA":             fmt.Sprintf("[%s]", strings.Join(bpsData, ", ")),
		"TOTAL_LINES":          fmt.Sprintf("%d", resp.TotalLines),
		"TOTAL_LINES_HUMANIZE": humanize.SI(float64(resp.TotalLines), "Entries"),
		"LOGLINES_DATA":        fmt.Sprintf("[%s]", strings.Join(logLineData, ", ")),
	})
}

type BytesPerSecond struct {
	Date               string `json:"date"`
	TotalLines         int    `json:"totalLines"`
	TotalBytes         int    `json:"totalBytes"`
	TotalBytesHumanize string `json:"totalBytesHumanize"`
}

type StatusResp struct {
	TraceID            string           `json:"traceID"`
	TimeStarted        time.Time        `json:"timeStarted"`
	TimeEnded          time.Time        `json:"timeEnded"`
	TotalLines         int              `json:"totalLines"`
	TotalBytes         int              `json:"totalBytes"`
	TotalBytesHumanize string           `json:"totalBytesHumanize"`
	BytesPerSecond     []BytesPerSecond `json:"bytesPerSecond"`
}

func (h *GenerateLogHandler) GetMetric(ctx context.Context, metricPrefix, traceID string) (StatusResp, error) {
	logLineMetricName := fmt.Sprintf("%s:log_printed:%s", metricPrefix, traceID)
	logLineMetric, err := h.RedisClient.HGetAll(ctx, logLineMetricName).Result()
	if err != nil {
		return StatusResp{}, fmt.Errorf("cannot get metric log printed '%s': %w", logLineMetricName, err)
	}

	byteLineMetricName := fmt.Sprintf("%s:bytes_printed:%s", metricPrefix, traceID)
	byteLineMetric, err := h.RedisClient.HGetAll(ctx, byteLineMetricName).Result()
	if err != nil {
		return StatusResp{}, fmt.Errorf("cannot get metric bytes printed '%s': %w", byteLineMetricName, err)
	}

	bps := make(map[string]BytesPerSecond)
	totalLine := 0
	for timestamp, countStr := range logLineMetric {
		data, exist := bps[timestamp]
		if !exist {
			bps[timestamp] = BytesPerSecond{}
		}

		count, _ := strconv.Atoi(countStr)
		totalLine += count

		data.Date = timestamp
		data.TotalLines = count

		bps[timestamp] = data
	}

	totalBytes := 0
	for timestamp, bytesPrinted := range byteLineMetric {
		data, exist := bps[timestamp]
		if !exist {
			bps[timestamp] = BytesPerSecond{}
		}

		bytesInt, _ := strconv.Atoi(bytesPrinted)
		totalBytes += bytesInt

		data.Date = timestamp
		data.TotalBytes = bytesInt
		data.TotalBytesHumanize = humanize.Bytes(uint64(bytesInt))

		bps[timestamp] = data
	}

	bytesPerSecond := make([]BytesPerSecond, 0)
	for _, data := range bps {
		bytesPerSecond = append(bytesPerSecond, data)
	}

	sort.Slice(bytesPerSecond, func(i, j int) bool {
		return bytesPerSecond[i].Date < bytesPerSecond[j].Date
	})

	var startTime, endTime time.Time
	if len(bytesPerSecond) >= 1 {
		startTime, _ = time.Parse(time.RFC3339, bytesPerSecond[0].Date)
		endTime, _ = time.Parse(time.RFC3339, bytesPerSecond[len(bytesPerSecond)-1].Date)
	}

	resp := StatusResp{
		TraceID:            traceID,
		TimeStarted:        startTime,
		TimeEnded:          endTime,
		TotalLines:         totalLine,
		TotalBytes:         totalBytes,
		TotalBytesHumanize: humanize.Bytes(uint64(totalBytes)),
		BytesPerSecond:     bytesPerSecond,
	}

	return resp, nil
}
