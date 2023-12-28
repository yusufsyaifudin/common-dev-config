package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/dustin/go-humanize"
	"github.com/jessevdk/go-flags"
)

type Args struct {
	StartTime        string `required:"true" long:"start-time" description:"Start time"`
	EndTime          string `required:"true" long:"end-time" description:"End time"`
	MaxBytePerSecond string `required:"true" long:"max-byte-per-second" description:"Max byte per second"`
	TraceID          string `required:"true" long:"trace-id" description:"Trace ID"`
	MetricPrefix     string `required:"true" long:"metric-prefix" description:"Metric prefix"`
	StatsdAddress    string `required:"true" long:"statsd-address" description:"StatsD address"`
	RedisAddress     string `required:"false" long:"redis-address" description:"Redis address"`
}

func main() {
	logJson := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	})
	slog.SetDefault(slog.New(logJson))

	ctx := context.Background()

	// Parsing Command Argument
	var argsVal Args
	_, err := flags.ParseArgs(&argsVal, os.Args)
	if err != nil {
		slog.ErrorContext(ctx, "failed parsing flag", slog.Any("error", err))
		return
	}

	traceID := argsVal.TraceID

	var startTime time.Time
	{
		startTimeStr := strings.TrimSpace(argsVal.StartTime)
		if startTimeStr == "" {
			slog.ErrorContext(ctx, "empty start_time arg")
			return
		}

		// example: 2023-11-06T20:45:00+07:00
		// in URL, + must encode as %2b such as: 2023-11-06T20:45:00%2b07:00
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			slog.ErrorContext(ctx, "start time is empty", slog.Any("error", err))
			return
		}
	}

	var endTime time.Time
	{
		endTimeStr := strings.TrimSpace(argsVal.EndTime)
		if endTimeStr == "" {
			slog.ErrorContext(ctx, "empty end_time arg")
			return
		}

		// example: 2023-11-06T20:45:00+07:00
		// in URL, + must encode as %2b such as: 2023-11-06T20:45:00%2b07:00
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			slog.ErrorContext(ctx, "end time is empty", slog.Any("error", err))
			return
		}
	}

	var bytePerSecond int
	{
		bytePerSecondStr := strings.TrimSpace(argsVal.MaxBytePerSecond)
		if bytePerSecondStr == "" {
			slog.ErrorContext(ctx, "empty bytes_per_second params")
			return
		}

		var bps uint64
		bps, err = humanize.ParseBytes(bytePerSecondStr)
		if err != nil {
			slog.ErrorContext(ctx, "bytes_per_second is not valid number", slog.Any("error", err))
			return
		}

		if bps > humanize.MByte {
			slog.ErrorContext(ctx, fmt.Sprintf("one instance cannot generate more than 1 MB per second, you request for %s", humanize.Bytes(bps)))
			return
		}

		bytePerSecond = int(bps)
	}

	goStatsd, err := statsd.New(argsVal.StatsdAddress, statsd.WithNamespace(argsVal.MetricPrefix))
	if err != nil {
		slog.ErrorContext(ctx, "statsd client error", slog.Any("error", err), slog.String("statsd_address", argsVal.StatsdAddress))
		return
	}
	defer func() {
		if goStatsd == nil {
			return
		}

		if _err := goStatsd.Close(); _err != nil {
			slog.ErrorContext(ctx, "cannot close statsd", slog.Any("error", _err))
		}
	}()

	if argsVal.RedisAddress == "" {
		var localRedis *miniredis.Miniredis
		localRedis, err = miniredis.Run()
		if err != nil {
			slog.ErrorContext(ctx, "miniredis error", slog.Any("error", err))
			return
		}

		if localRedis == nil {
			slog.ErrorContext(ctx, "miniredis is nil, it may required when redis is not set")
			return
		}

		argsVal.RedisAddress = localRedis.Addr()
	}

	rdb := redis.NewClient(&redis.Options{
		Addr: argsVal.RedisAddress,
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

	type logFunc func(ctx context.Context, msg string, args ...any)
	listFuncLog := []logFunc{
		slog.DebugContext,
		slog.InfoContext,
		slog.WarnContext,
		slog.ErrorContext,
	}

	random := rand.New(rand.NewSource(time.Now().Unix()))
	faker := gofakeit.NewCustom(random)

	tickerLog := time.NewTicker(1 * time.Second)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			select {

			case t := <-tickerLog.C:
				if t.UnixNano() < startTime.UnixNano() {
					// don't print before start time
					continue
				}

				if t.UnixNano() >= endTime.UnixNano() {
					// stop goroutine after end time
					// Not print log so we can count exact log line printed by this service using the request id.
					// slog.InfoContext(ctx, "stopped log line", slog.String("traceID", traceID), slog.Time("endTime", endTime))
					wg.Done()
					return
				}

				// generate log until meet bytes_per_second
				batchLogs, batchLogsErr := GenerateBatch(faker, bytePerSecond)
				if batchLogsErr != nil {
					slog.ErrorContext(ctx, "cannot generate batch",
						slog.String("traceID", traceID),
						slog.Time("time", t),
						slog.Any("error", batchLogsErr),
					)
					wg.Done()
					return
				}

				lineCount := 0
				lineBytes := 0
				for _, logLine := range batchLogs {
					lineCount++
					lineBytes += logLine.CurrentBytes
					logWriter := listFuncLog[rand.Intn(len(listFuncLog))]

					logWriter(ctx,
						fmt.Sprintf("iteration=%d filled bytes=%s current bytes=%s", logLine.Iter, logLine.FilledBytesHuman, logLine.CurrentBytesHuman),
						slog.String("traceID", traceID),
						slog.Any("payload", logLine.Payload),
					)

					go func() {
						redisFieldKey := t.Format(time.RFC3339)
						logLineMetricName := fmt.Sprintf("%s:log_printed:%s", argsVal.MetricPrefix, argsVal.TraceID)
						if _err := rdb.HIncrBy(ctx, logLineMetricName, redisFieldKey, 1).Err(); _err != nil {
							slog.ErrorContext(ctx, "hincr log_printed redis error", slog.String("key", logLineMetricName), slog.Any("error", _err))
						}

						byteLineMetricName := fmt.Sprintf("%s:bytes_printed:%s", argsVal.MetricPrefix, argsVal.TraceID)
						if _err := rdb.HIncrBy(ctx, byteLineMetricName, redisFieldKey, int64(logLine.CurrentBytes)).Err(); _err != nil {
							slog.ErrorContext(ctx, "hincr bytes_printed redis error", slog.String("key", byteLineMetricName), slog.Any("error", _err))
						}
					}()

					statsdTags := []string{
						fmt.Sprintf("traceid:%s", argsVal.TraceID),
					}

					if _err := goStatsd.Incr("log_printed", statsdTags, 1); _err != nil {
						slog.ErrorContext(ctx, "cannot push increment log lines metric to statsd", slog.Any("error", _err))
					}

					if _err := goStatsd.Histogram("bytes_printed", float64(logLine.CurrentBytes), statsdTags, 1); _err != nil {
						slog.ErrorContext(ctx, "cannot push bytes printed metric to statsd", slog.Any("error", _err))
					}
				}

			}
		}
	}()

	wg.Wait()
}

type BatchLog struct {
	Iter              int
	CurrentBytes      int
	CurrentBytesHuman string
	FilledBytes       int
	FilledBytesHuman  string
	Payload           any
}

type BatchLogs []BatchLog

const JSONLogFormat = `
{
  "user": {
	"is_active": %t,
	"name": {
		"first": "%s",
		"middle": "%s",
		"last": "%s"
	},
	"username": "%s",
	"emails": [
		"%s@gmail.com"
	],
	"placeOfBirth": "%s",
	"dateOfBirth": "%s",
	"phoneNumber": "%s",
	"location": {
		"street": "%s",
		"city": "%s",
		"state": "%s",
		"country": "%s",
		"zip": "%s",
		"coordinates": {
		  "latitude": %.3f,
		  "longitude": %.3f
		}
	},
	"website": "%s",
	"job": {
		"title": "%s",
		"descriptor": "%s",
		"level": "%s",
		"company": "%s"
	},
    "pet": [
		"%s",
		"%s"
	],
	"bitcoinAddress": "%s",
    "bookGenre": [
		"%s",
		"%s",
		"%s"
	]
  },
  "request": {
	"ip": "%s",
	"version": "%s",
	"status": %d,
    "userAgent": "%s",
	"requestID": "%s"
  }
}
`

func GenerateBatch(fakeIt *gofakeit.Faker, maxByte int) (BatchLogs, error) {
	out := make(BatchLogs, 0)

	var logLine string
	var currFilledByte, currByte, iter int
	for ; currFilledByte < maxByte; currFilledByte += currByte {
		iter++

		nameFirst, nameMiddle, nameLast := fakeIt.FirstName(), fakeIt.MiddleName(), fakeIt.LastName()
		userName := strings.ToLower(nameFirst) + "_" + strings.ToLower(nameLast)
		userJob := fakeIt.Job()
		logLine = fmt.Sprintf(
			JSONLogFormat,
			fakeIt.Bool(),
			nameFirst,
			nameMiddle,
			nameLast,
			userName,
			userName,
			fakeIt.City()+", "+fakeIt.Country(),
			fakeIt.Date().Format(time.DateOnly),
			fakeIt.PhoneFormatted(),
			fakeIt.StreetName(),
			fakeIt.City(),
			fakeIt.State(),
			fakeIt.Country(),
			fakeIt.Zip(),
			fakeIt.Latitude(),
			fakeIt.Longitude(),
			fakeIt.URL(),
			userJob.Title,
			userJob.Descriptor,
			userJob.Level,
			userJob.Company,
			fakeIt.PetName(),
			fakeIt.PetName(),
			fakeIt.BitcoinAddress(),
			fakeIt.BookGenre(),
			fakeIt.BookGenre(),
			fakeIt.BookGenre(),

			fakeIt.IPv4Address(),
			fakeIt.HTTPVersion(),
			fakeIt.HTTPStatusCodeSimple(),
			fakeIt.UserAgent(),
			fakeIt.UUID(),
		)

		var logLineI any
		if err := json.Unmarshal([]byte(logLine), &logLineI); err != nil {
			return nil, fmt.Errorf("cannot encode json: %w", err)
		}

		currByte = len(logLine)
		out = append(out, BatchLog{
			Iter:              iter,
			CurrentBytes:      currByte,
			CurrentBytesHuman: humanize.Bytes(uint64(currByte)),
			FilledBytes:       currFilledByte + currByte,
			FilledBytesHuman:  humanize.Bytes(uint64(currFilledByte + currByte)),
			Payload:           logLineI,
		})
	}

	return out, nil
}
