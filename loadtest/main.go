package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"
	"github.com/shirou/gopsutil/v4/net"
)

type ingestPayload struct {
	Logs []string `json:"logs"`
}

func main() {
	var (
		endpoint = flag.String("endpoint", "http://localhost:3333/api/v1/logs/ingest", "HTTP endpoint")
		apiKey   = flag.String("api-key", "", "API key (optional)")
		duration = flag.Duration("duration", 1*time.Minute, "test duration")
		outFile  = flag.String("out", "load_metrics.csv", "csv output file")
		rps      = flag.Int("rps", 10, "requests per second")
	)
	flag.Parse()

	f, err := os.Create(*outFile)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	w.Comma = ';'
	defer w.Flush()

	_ = w.Write([]string{
		"ts", "cpu_percent", "mem_used_mb",
		"net_sent_kbps", "net_recv_kbps",
		"req_ok", "req_fail", "last_req_ms",
	})

	client := &http.Client{Timeout: 10 * time.Second}
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	if *rps <= 0 {
		log.Fatal("rps must be > 0")
	}
	requestInterval := time.Second / time.Duration(*rps)

	var reqOK, reqFail int64
	var lastReqMs int64

	// для расчета сетевой скорости
	prevIO, _ := net.IOCounters(false)
	var prevSent, prevRecv uint64
	if len(prevIO) > 0 {
		prevSent = prevIO[0].BytesSent
		prevRecv = prevIO[0].BytesRecv
	}

	// Частота запросов регулируется параметром -rps
	sendTicker := time.NewTicker(requestInterval)
	metricsTicker := time.NewTicker(1 * time.Second)
	defer sendTicker.Stop()
	defer metricsTicker.Stop()

	log.Println("load test started...")

	for {
		select {
		case <-ctx.Done():
			log.Println("done")
			return

		case <-sendTicker.C:
			start := time.Now()

			payload := ingestPayload{
				Logs: []string{
					genLog("INFO"),
					genLog("WARN"),
					genLog("ERROR"),
					genLog("DEBUG"),
					genLog("INFO"),
				},
			}

			body, _ := json.Marshal(payload)
			req, _ := http.NewRequest(http.MethodPost, *endpoint, bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			if *apiKey != "" {
				// если у тебя другой заголовок — замени здесь
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", *apiKey))
			}

			resp, err := client.Do(req)
			if err != nil {
				atomic.AddInt64(&reqFail, 1)
				atomic.StoreInt64(&lastReqMs, time.Since(start).Milliseconds())
				continue
			}
			_ = resp.Body.Close()

			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				atomic.AddInt64(&reqOK, 1)
			} else {
				atomic.AddInt64(&reqFail, 1)
			}
			atomic.StoreInt64(&lastReqMs, time.Since(start).Milliseconds())

		case t := <-metricsTicker.C:
			cpuP := 0.0
			if vals, err := cpu.Percent(0, false); err == nil && len(vals) > 0 {
				cpuP = vals[0]
			}

			vm, _ := mem.VirtualMemory()

			curIO, _ := net.IOCounters(false)
			var sentKBps, recvKBps float64
			if len(curIO) > 0 {
				sentDelta := curIO[0].BytesSent - prevSent
				recvDelta := curIO[0].BytesRecv - prevRecv
				sentKBps = float64(sentDelta) / 1024.0
				recvKBps = float64(recvDelta) / 1024.0
				prevSent = curIO[0].BytesSent
				prevRecv = curIO[0].BytesRecv
			}

			_ = w.Write([]string{
				t.Format(time.RFC3339),
				toLocaleNumber(cpuP),
				toLocaleNumber(float64(vm.Used) / 1024.0 / 1024.0),
				toLocaleNumber(sentKBps),
				toLocaleNumber(recvKBps),
				fmt.Sprintf("%d", atomic.LoadInt64(&reqOK)),
				fmt.Sprintf("%d", atomic.LoadInt64(&reqFail)),
				fmt.Sprintf("%d", atomic.LoadInt64(&lastReqMs)),
			})
			w.Flush()
		}
	}
}

func genLog(level string) string {
	timestamp := time.Now().Format(time.RFC3339)

	// ~75% логов — разнообразные, ~25% — из «повторяемого» паттерна (для схожести)
	useSimilarPattern := rand.Intn(100) < 25

	services := []string{"api", "auth", "billing", "orders", "notifications", "search"}
	envs := []string{"dev", "stage", "prod"}
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	paths := []string{"/api/v1/users", "/api/v1/projects", "/api/v1/applications", "/api/v1/logs/ingest", "/api/v1/auth/login"}
	hosts := []string{"app-01", "app-02", "worker-01", "edge-01"}
	userAgents := []string{"curl/8.7", "PostmanRuntime/7.43", "Go-http-client/1.1", "Mozilla/5.0"}
	events := []string{"request_processed", "validation_failed", "db_timeout", "cache_miss", "auth_failed", "ingest_accepted"}

	service := services[rand.Intn(len(services))]
	env := envs[rand.Intn(len(envs))]
	method := methods[rand.Intn(len(methods))]
	path := paths[rand.Intn(len(paths))]
	host := hosts[rand.Intn(len(hosts))]
	ua := userAgents[rand.Intn(len(userAgents))]
	event := events[rand.Intn(len(events))]
	status := []int{200, 201, 202, 400, 401, 403, 404, 429, 500, 503}[rand.Intn(10)]
	duration := []int{8, 12, 25, 40, 75, 120, 240, 520}[rand.Intn(8)]
	ip := fmt.Sprintf("10.%d.%d.%d", rand.Intn(32)+1, rand.Intn(256), rand.Intn(256))

	if useSimilarPattern {
		// Ограничиваем вариативность, чтобы часть логов была похожей
		service = []string{"api", "auth"}[rand.Intn(2)]
		env = "prod"
		method = []string{"POST", "GET"}[rand.Intn(2)]
		path = []string{"/api/v1/logs/ingest", "/api/v1/auth/login"}[rand.Intn(2)]
		host = []string{"app-01", "edge-01"}[rand.Intn(2)]
		ua = []string{"Go-http-client/1.1", "PostmanRuntime/7.43"}[rand.Intn(2)]
		event = []string{"ingest_accepted", "auth_failed"}[rand.Intn(2)]
		status = []int{202, 401, 429}[rand.Intn(3)]
		duration = []int{25, 40, 75}[rand.Intn(3)]
	}

	parts := []string{fmt.Sprintf("%s %s", timestamp, level)}

	// Канонические поля добавляются случайно (не все сразу)
	if rand.Intn(100) < 95 {
		parts = append(parts, fmt.Sprintf("service=%s", service))
	}
	if rand.Intn(100) < 70 {
		parts = append(parts, fmt.Sprintf("environment=%s", env))
	}
	if rand.Intn(100) < 85 {
		parts = append(parts, fmt.Sprintf("event=%s", event))
	}
	if rand.Intn(100) < 80 {
		parts = append(parts, fmt.Sprintf("status_code=%d", status))
	}
	if rand.Intn(100) < 75 {
		parts = append(parts, fmt.Sprintf("duration=%dms", duration))
	}
	if rand.Intn(100) < 65 {
		parts = append(parts, fmt.Sprintf("ip=%s", ip))
	}
	if rand.Intn(100) < 65 {
		parts = append(parts, fmt.Sprintf("method=%s", method))
	}
	if rand.Intn(100) < 65 {
		parts = append(parts, fmt.Sprintf("path=%s", path))
	}
	if rand.Intn(100) < 55 {
		parts = append(parts, fmt.Sprintf("useragent=%s", ua))
	}
	if rand.Intn(100) < 50 {
		parts = append(parts, fmt.Sprintf("hostname=%s", host))
	}
	if rand.Intn(100) < 35 {
		parts = append(parts, fmt.Sprintf("error_message=%s", []string{"timeout", "invalid_token", "upstream_unavailable", "none"}[rand.Intn(4)]))
	}

	msg := "ok"
	if status >= 400 {
		msg = "failed"
	}
	parts = append(parts, fmt.Sprintf("message=%s", msg))

	return fmt.Sprint(parts[0], " ", joinTail(parts[1:]))
}

func joinTail(items []string) string {
	if len(items) == 0 {
		return ""
	}
	result := items[0]
	for i := 1; i < len(items); i++ {
		result += " " + items[i]
	}
	return result
}

func toLocaleNumber(v float64) string {
	return strings.ReplaceAll(fmt.Sprintf("%.2f", v), ".", ",")
}
