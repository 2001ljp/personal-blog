package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type headerFlags []string

func (h *headerFlags) String() string {
	return strings.Join(*h, ", ")
}

func (h *headerFlags) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func (h headerFlags) Header() (http.Header, error) {
	if len(h) == 0 {
		return make(http.Header), nil
	}
	headers := make(http.Header)
	for _, raw := range h {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format %q, expected Key: Value", raw)
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("empty header key in %q", raw)
		}
		headers.Add(key, val)
	}
	return headers, nil
}

type requestFactory func(ctx context.Context) (*http.Request, error)

func newRequestFactory(method, target string, body []byte, headers http.Header) requestFactory {
	return func(ctx context.Context) (*http.Request, error) {
		var reader io.Reader
		if len(body) > 0 {
			reader = bytes.NewReader(body)
		}
		req, err := http.NewRequestWithContext(ctx, method, target, reader)
		if err != nil {
			return nil, err
		}
		for k, vals := range headers {
			for _, v := range vals {
				req.Header.Add(k, v)
			}
		}
		return req, nil
	}
}

type sample struct {
	latency time.Duration
	status  int
	err     error
}

type stepResult struct {
	RPS           float64
	Duration      time.Duration
	Attempts      int
	Success       int
	Failures      int
	StatusCounts  map[int]int
	ErrorCounts   map[string]int
	Latencies     []time.Duration
	AvgLatency    time.Duration
	P50Latency    time.Duration
	P90Latency    time.Duration
	P95Latency    time.Duration
	P99Latency    time.Duration
	MaxLatency    time.Duration
	ThroughputRPS float64
}

func (sr *stepResult) finalize() {
	if len(sr.Latencies) == 0 {
		return
	}
	sort.Slice(sr.Latencies, func(i, j int) bool {
		return sr.Latencies[i] < sr.Latencies[j]
	})
	var total time.Duration
	for _, l := range sr.Latencies {
		total += l
		if l > sr.MaxLatency {
			sr.MaxLatency = l
		}
	}
	sr.AvgLatency = total / time.Duration(len(sr.Latencies))
	sr.P50Latency = percentile(sr.Latencies, 50)
	sr.P90Latency = percentile(sr.Latencies, 90)
	sr.P95Latency = percentile(sr.Latencies, 95)
	sr.P99Latency = percentile(sr.Latencies, 99)
	if sr.Duration > 0 {
		sr.ThroughputRPS = float64(sr.Success) / sr.Duration.Seconds()
	}
}

func percentile(latencies []time.Duration, percent int) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	if percent <= 0 {
		return latencies[0]
	}
	if percent >= 100 {
		return latencies[len(latencies)-1]
	}
	index := (percent*len(latencies) + 99) / 100 // ceil
	if index <= 0 {
		index = 1
	}
	if index > len(latencies) {
		index = len(latencies)
	}
	return latencies[index-1]
}

type stepConfig struct {
	rps            float64
	duration       time.Duration
	burst          int
	workers        int
	requestTimeout time.Duration
}

func runStep(ctx context.Context, client *http.Client, factory requestFactory, cfg stepConfig) stepResult {
	stepCtx, cancel := context.WithTimeout(ctx, cfg.duration)
	defer cancel()

	result := stepResult{
		RPS:          cfg.rps,
		Duration:     cfg.duration,
		StatusCounts: make(map[int]int),
		ErrorCounts:  make(map[string]int),
	}
	limiter := rate.NewLimiter(rate.Limit(cfg.rps), cfg.burst)
	jobs := make(chan struct{}, cfg.workers*2)
	results := make(chan sample, cfg.workers*4)
	var wg sync.WaitGroup

	for w := 0; w < cfg.workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stepCtx.Done():
					return
				case _, ok := <-jobs:
					if !ok {
						return
					}
					reqCtx, cancelReq := context.WithTimeout(stepCtx, cfg.requestTimeout)
					req, err := factory(reqCtx)
					if err != nil {
						cancelReq()
						results <- sample{latency: 0, err: err}
						continue
					}
					start := time.Now()
					resp, err := client.Do(req)
					if err != nil {
						cancelReq()
						results <- sample{latency: time.Since(start), err: err}
						continue
					}
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
					cancelReq()
					results <- sample{latency: time.Since(start), status: resp.StatusCode}
				}
			}
		}()
	}

	var aggWG sync.WaitGroup
	aggWG.Add(1)
	go func() {
		defer aggWG.Done()
		for res := range results {
			result.Attempts++
			if res.err != nil {
				result.Failures++
				result.ErrorCounts[res.err.Error()]++
			} else if res.status >= 200 && res.status < 300 {
				result.Success++
				result.Latencies = append(result.Latencies, res.latency)
			} else {
				result.Failures++
				result.StatusCounts[res.status]++
				result.Latencies = append(result.Latencies, res.latency)
			}
		}
	}()

	sendDone := make(chan struct{})
	go func() {
		defer close(sendDone)
		defer close(jobs)
		for {
			if err := limiter.Wait(stepCtx); err != nil {
				return
			}
			select {
			case <-stepCtx.Done():
				return
			case jobs <- struct{}{}:
			}
		}
	}()

	<-stepCtx.Done()
	<-sendDone
	wg.Wait()
	close(results)
	aggWG.Wait()

	result.finalize()
	return result
}

func loadBody(raw string) ([]byte, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	if strings.HasPrefix(raw, "@") {
		path := strings.TrimPrefix(raw, "@")
		return os.ReadFile(path)
	}
	return []byte(raw), nil
}

func main() {
	var (
		targetURL      = flag.String("url", "http://127.0.0.1:8081/api/v1/posts/", "target URL to benchmark")
		method         = flag.String("method", http.MethodGet, "HTTP method")
		bodyRaw        = flag.String("body", "", "request body or @path/to/body.json")
		startRPS       = flag.Float64("start-rps", 50, "starting RPS")
		maxRPS         = flag.Float64("max-rps", 500, "maximum RPS to test")
		stepRPS        = flag.Float64("step-rps", 50, "RPS increment per step")
		stepDuration   = flag.Duration("step-duration", 10*time.Second, "duration of each RPS step")
		burstSize      = flag.Int("burst", 100, "token bucket burst size per step")
		workers        = flag.Int("workers", 32, "number of concurrent workers")
		requestTimeout = flag.Duration("request-timeout", 3*time.Second, "per-request timeout")
		maxP95Latency  = flag.Duration("max-p95", 500*time.Millisecond, "max acceptable P95 latency for stable RPS")
		maxErrorRate   = flag.Float64("max-error-rate", 0.02, "max acceptable error rate for stable RPS")
		clientTimeout  = flag.Duration("client-timeout", 5*time.Second, "HTTP client timeout")
	)
	var h headerFlags
	flag.Var(&h, "H", "additional request headers, repeatable, format 'Key: Value'")
	flag.Parse()

	body, err := loadBody(*bodyRaw)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load body failed: %v\n", err)
		os.Exit(1)
	}
	headers, err := h.Header()
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse headers failed: %v\n", err)
		os.Exit(1)
	}
	client := &http.Client{
		Timeout: *clientTimeout,
	}
	factory := newRequestFactory(*method, *targetURL, body, headers)

	var results []stepResult
	ctx := context.Background()
	var baselineP95 time.Duration
	for rps := *startRPS; rps <= *maxRPS; rps += *stepRPS {
		fmt.Printf("\n=== Running step: target RPS %.2f for %s ===\n", rps, stepDuration.String())
		res := runStep(ctx, client, factory, stepConfig{
			rps:            rps,
			duration:       *stepDuration,
			burst:          *burstSize,
			workers:        *workers,
			requestTimeout: *requestTimeout,
		})
		if baselineP95 == 0 && res.P95Latency > 0 {
			baselineP95 = res.P95Latency
		}
		results = append(results, res)
		printStep(res, baselineP95)
		if res.Success == 0 {
			fmt.Println("No successful requests, aborting further steps.")
			break
		}
	}
	if len(results) == 0 {
		fmt.Println("no results collected")
		return
	}
	stable := calculateStableRPS(results, *maxP95Latency, *maxErrorRate)
	fmt.Printf("\nStable burst RPS (p95<=%s, error<=%.2f%%): %.2f\n",
		maxP95Latency, *maxErrorRate*100, stable)
}

func printStep(res stepResult, baselineP95 time.Duration) {
	errorRate := 0.0
	if res.Attempts > 0 {
		errorRate = float64(res.Failures) / float64(res.Attempts) * 100
	}
	fmt.Printf("Attempts: %d, Success: %d, Failures: %d (error %.2f%%)\n",
		res.Attempts, res.Success, res.Failures, errorRate)
	fmt.Printf("Throughput: %.2f req/s, Avg latency: %s, P50/P90/P95/P99: %s / %s / %s / %s (max %s)\n",
		res.ThroughputRPS, res.AvgLatency, res.P50Latency, res.P90Latency, res.P95Latency, res.P99Latency, res.MaxLatency)
	if baselineP95 > 0 {
		delta := res.P95Latency - baselineP95
		label := "flat"
		if delta > 0 {
			label = "+" + delta.String()
		} else if delta < 0 {
			label = delta.String()
		}
		fmt.Printf("Latency delta vs baseline P95 (%s): %s\n", baselineP95, label)
	}
	if len(res.StatusCounts) > 0 {
		fmt.Println("Non-2xx statuses:")
		for code, count := range res.StatusCounts {
			fmt.Printf("  %d => %d\n", code, count)
		}
	}
	if len(res.ErrorCounts) > 0 {
		fmt.Println("Client errors:")
		for msg, count := range res.ErrorCounts {
			fmt.Printf("  %s => %d\n", msg, count)
		}
	}
}

func calculateStableRPS(results []stepResult, maxP95 time.Duration, maxError float64) float64 {
	var stable float64
	for _, res := range results {
		errorRate := 1.0
		if res.Attempts > 0 {
			errorRate = float64(res.Failures) / float64(res.Attempts)
		}
		if res.P95Latency <= maxP95 && errorRate <= maxError {
			stable = res.RPS
		} else {
			break
		}
	}
	return stable
}
