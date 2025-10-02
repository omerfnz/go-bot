# Go-SERP-Bot Teknik Tasarım Dokümanı

**Doküman Durumu:** Taslak  
**Oluşturulma Tarihi:** 1 Ekim 2025  
**Versiyon:** 1.0

---

## 1. Sistem Mimarisi

### 1.1. Genel Mimari

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                            │
│                    (cobra commands)                          │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                   Task Scheduler                             │
│              (Sonsuz döngü yönetimi)                        │
└──────────────────────┬──────────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────────┐
│                   Worker Pool                                │
│        (Goroutine pool, concurrency control)                │
└─────┬────────────────┬────────────────┬────────────────┬───┘
      │                │                │                │
┌─────▼────┐    ┌─────▼────┐    ┌─────▼────┐    ┌─────▼────┐
│ Worker 1 │    │ Worker 2 │    │ Worker 3 │    │ Worker N │
└─────┬────┘    └─────┬────┘    └─────┬────┘    └─────┬────┘
      │                │                │                │
┌─────▼────────────────▼────────────────▼────────────────▼───┐
│                    Task Execution                           │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐  │
│  │  Proxy   │  │ Browser  │  │  SERP    │  │  Stats   │  │
│  │  Manager │  │  Manager │  │  Scraper │  │ Collector│  │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘  │
└──────────────────────────────────────────────────────────────┘
```

### 1.2. Veri Akışı

```
1. Config Load → 2. Proxy Pool Init → 3. Task Queue Create
                                            ↓
4. Worker Pool Start ← ───────────────────────
                ↓
5. Worker gets Task → 6. Select Proxy → 7. Launch Browser
                                            ↓
8. Navigate to Google → 9. Search Keyword → 10. Find Target
                                            ↓
11. Record Rank → 12. Click Target → 13. Browse Site
                                            ↓
14. Record Stats → 15. Close Browser → 16. Return to Pool
                                            ↓
17. Wait Interval → 18. Loop Back to Step 4
```

---

## 2. Modül Detayları

### 2.1. Config Module (`internal/config`)

**Sorumluluklar:**
- `config.json` ve `.env` dosyalarını okuma
- Konfigürasyon validasyonu
- Environment variable override

**Struct Tasarımı:**
```go
type Config struct {
    // Genel ayarlar
    Headless    bool   `json:"headless" env:"HEADLESS"`
    Workers     int    `json:"workers" env:"WORKERS"`
    Interval    int    `json:"interval" env:"INTERVAL"`
    
    // Task ayarları
    Keywords    []Keyword `json:"keywords"`
    Proxies     []string  `json:"proxies"`
    
    // Timeout ayarları
    PageTimeout    int `json:"page_timeout"`
    SearchTimeout  int `json:"search_timeout"`
    
    // Retry ayarları
    MaxRetries     int `json:"max_retries"`
    RetryDelay     int `json:"retry_delay"`
    
    // Selectors
    Selectors   SelectorConfig `json:"selectors"`
}

type Keyword struct {
    Term       string `json:"term"`
    TargetURL  string `json:"target_url"`
}

type SelectorConfig struct {
    SearchBox    string `json:"search_box"`
    SearchButton string `json:"search_button"`
    ResultItem   string `json:"result_item"`
    ResultLink   string `json:"result_link"`
    NextButton   string `json:"next_button"`
}
```

**Fonksiyonlar:**
```go
func Load(configPath string) (*Config, error)
func (c *Config) Validate() error
func (c *Config) LoadEnv() error
```

---

### 2.2. Proxy Module (`internal/proxy`)

**Sorumluluklar:**
- Proxy pool yönetimi
- Proxy rotation (round-robin/random)
- Proxy validation ve health check
- Blacklist yönetimi

**Struct Tasarımı:**
```go
type Proxy struct {
    URL          string
    Host         string
    Port         int
    Username     string // Optional (v1.3)
    Password     string // Optional (v1.3)
    Type         ProxyType // HTTP, HTTPS, SOCKS5
    LastUsed     time.Time
    FailCount    int
    SuccessCount int
    IsBlacklisted bool
}

type ProxyPool struct {
    proxies      []*Proxy
    current      int
    strategy     RotationStrategy // RoundRobin, Random
    mu           sync.Mutex
    blacklist    map[string]bool
    validator    *ProxyValidator
}

type ProxyValidator struct {
    testURL     string
    timeout     time.Duration
    httpClient  *http.Client
}
```

**Fonksiyonlar:**
```go
func NewProxyPool(proxies []string, strategy RotationStrategy) *ProxyPool
func (pp *ProxyPool) Get() (*Proxy, error)
func (pp *ProxyPool) Release(proxy *Proxy, success bool)
func (pp *ProxyPool) Blacklist(proxy *Proxy)
func (pp *ProxyPool) ValidateAll(ctx context.Context) error
func (pv *ProxyValidator) Validate(proxy *Proxy) error
```

---

### 2.3. Browser Module (`internal/browser`)

**Sorumluluklar:**
- chromedp wrapper ve lifecycle management
- Stealth mode (bot detection bypass)
- Human-like interactions
- Screenshot (opsiyonel)

**Struct Tasarımı:**
```go
type Browser struct {
    ctx         context.Context
    cancel      context.CancelFunc
    allocCtx    context.Context
    allocCancel context.CancelFunc
    proxy       *Proxy
    userAgent   string
    headless    bool
}

type BrowserOptions struct {
    Headless      bool
    Proxy         *Proxy
    UserAgent     string
    WindowSize    [2]int // [width, height]
    Timezone      string
    Language      string
    Stealth       bool
}

type HumanSimulator struct {
    typingSpeed  time.Duration // 50-200ms per char
    scrollSpeed  time.Duration
    clickDelay   time.Duration
    rand         *rand.Rand
}
```

**Fonksiyonlar:**
```go
func NewBrowser(opts BrowserOptions) (*Browser, error)
func (b *Browser) Navigate(url string) error
func (b *Browser) TypeHumanLike(selector string, text string) error
func (b *Browser) Click(selector string) error
func (b *Browser) ScrollRandom() error
func (b *Browser) WaitRandom(min, max time.Duration) error
func (b *Browser) Screenshot(filename string) error
func (b *Browser) Close() error

// Stealth functions
func applyStealthMode(ctx context.Context) error
func randomizeFingerprint(ctx context.Context) error
```

---

### 2.4. SERP Module (`internal/serp`)

**Sorumluluklar:**
- Google arama yapma
- SERP sonuçlarını parse etme
- Hedef URL'in sıralamasını bulma
- Sayfa navigasyonu (next page)

**Struct Tasarımı:**
```go
type SearchResult struct {
    Title    string
    URL      string
    Position int
    Page     int
}

type Searcher struct {
    browser   *Browser
    selectors SelectorConfig
    logger    *logrus.Logger
}

type RankingInfo struct {
    Keyword   string
    TargetURL string
    Position  int
    Page      int
    Found     bool
    Timestamp time.Time
}
```

**Fonksiyonlar:**
```go
func NewSearcher(browser *Browser, selectors SelectorConfig) *Searcher
func (s *Searcher) Search(keyword string) error
func (s *Searcher) GetResults() ([]SearchResult, error)
func (s *Searcher) FindTarget(targetURL string) (*RankingInfo, error)
func (s *Searcher) NextPage() error
func (s *Searcher) ClickResult(position int) error
func (s *Searcher) BrowseTarget(duration time.Duration) error
```

---

### 2.5. Task Module (`internal/task`)

**Sorumluluklar:**
- Task struct ve görev yönetimi
- Worker pool implementation
- Sonsuz döngü ve scheduling
- Context yönetimi

**Struct Tasarımı:**
```go
type Task struct {
    ID        string
    Keyword   Keyword
    Proxy     *Proxy
    Status    TaskStatus // Pending, Running, Success, Failed
    Attempts  int
    Error     error
    Result    *TaskResult
    CreatedAt time.Time
    StartedAt time.Time
    EndedAt   time.Time
}

type TaskResult struct {
    Ranking    *RankingInfo
    Duration   time.Duration
    ProxyUsed  string
    Success    bool
    ErrorMsg   string
}

type WorkerPool struct {
    workers      int
    taskQueue    chan *Task
    resultQueue  chan *TaskResult
    wg           *sync.WaitGroup
    ctx          context.Context
    cancel       context.CancelFunc
    proxyPool    *ProxyPool
    config       *Config
}

type Scheduler struct {
    pool      *WorkerPool
    config    *Config
    stats     *StatsCollector
    interval  time.Duration
    running   bool
    mu        sync.Mutex
}
```

**Fonksiyonlar:**
```go
func NewTask(keyword Keyword) *Task
func NewWorkerPool(config *Config, proxyPool *ProxyPool) *WorkerPool
func (wp *WorkerPool) Start(ctx context.Context) error
func (wp *WorkerPool) Stop() error
func (wp *WorkerPool) Submit(task *Task)
func (wp *WorkerPool) worker(workerID int)

func NewScheduler(config *Config, pool *WorkerPool, stats *StatsCollector) *Scheduler
func (s *Scheduler) Start(ctx context.Context) error
func (s *Scheduler) Stop() error
func (s *Scheduler) runCycle() error
```

---

### 2.6. Stats Module (`internal/stats`)

**Sorumluluklar:**
- İstatistik toplama
- JSON dosyasına kaydetme
- Zaman serisi verisi yönetimi

**Struct Tasarımı:**
```go
type StatsCollector struct {
    stats    *Statistics
    filePath string
    mu       sync.RWMutex
}

type Statistics struct {
    TotalTasks      int                    `json:"total_tasks"`
    SuccessfulTasks int                    `json:"successful_tasks"`
    FailedTasks     int                    `json:"failed_tasks"`
    Keywords        map[string]*KeywordStats `json:"keywords"`
    Proxies         map[string]*ProxyStats   `json:"proxies"`
    StartTime       time.Time              `json:"start_time"`
    LastUpdate      time.Time              `json:"last_update"`
}

type KeywordStats struct {
    Keyword         string            `json:"keyword"`
    TotalSearches   int               `json:"total_searches"`
    SuccessfulClicks int              `json:"successful_clicks"`
    RankingHistory  []RankingSnapshot `json:"ranking_history"`
    AverageDuration float64           `json:"average_duration"`
}

type RankingSnapshot struct {
    Position  int       `json:"position"`
    Page      int       `json:"page"`
    Found     bool      `json:"found"`
    Timestamp time.Time `json:"timestamp"`
}

type ProxyStats struct {
    ProxyURL     string  `json:"proxy_url"`
    TotalUses    int     `json:"total_uses"`
    SuccessCount int     `json:"success_count"`
    FailCount    int     `json:"fail_count"`
    SuccessRate  float64 `json:"success_rate"`
}
```

**Fonksiyonlar:**
```go
func NewStatsCollector(filePath string) (*StatsCollector, error)
func (sc *StatsCollector) RecordTask(result *TaskResult)
func (sc *StatsCollector) RecordRanking(keyword string, ranking *RankingInfo)
func (sc *StatsCollector) RecordProxy(proxy *Proxy, success bool)
func (sc *StatsCollector) Save() error
func (sc *StatsCollector) Load() error
func (sc *StatsCollector) GetSummary() string
```

---

### 2.7. Logger Module (`internal/logger`)

**Sorumluluklar:**
- Yapılandırılabilir loglama (console + file)
- Log seviyeleri (DEBUG, INFO, WARN, ERROR)
- Structured logging

**Implementation:**
```go
func NewLogger(level string, logFile string) (*logrus.Logger, error)
func SetupLogger(config *Config) (*logrus.Logger, error)
```

---

## 3. Concurrency Stratejisi

### 3.1. Worker Pool Pattern

```go
// Worker pool ile goroutine sayısını sınırlama
type WorkerPool struct {
    workers     int
    taskQueue   chan *Task
    resultQueue chan *TaskResult
    wg          *sync.WaitGroup
}

// Her worker bir goroutine
func (wp *WorkerPool) worker(workerID int) {
    defer wp.wg.Done()
    
    for task := range wp.taskQueue {
        // Task execution
        result := wp.executeTask(task)
        wp.resultQueue <- result
    }
}
```

### 3.2. Context Kullanımı

```go
// Her task için timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

// Graceful shutdown
func (s *Scheduler) Stop() error {
    s.cancel() // Context cancel
    s.wg.Wait() // Tüm goroutine'lerin bitmesini bekle
    return nil
}
```

### 3.3. Resource Management

```go
// Her browser instance için defer cleanup
func executeTask(task *Task) (*TaskResult, error) {
    browser, err := NewBrowser(opts)
    if err != nil {
        return nil, err
    }
    defer browser.Close() // Mutlaka close
    
    // Task logic
    ...
}
```

---

## 4. Hata Yönetimi Stratejisi

### 4.1. Retry Mekanizması

```go
func retryWithBackoff(maxRetries int, fn func() error) error {
    var err error
    for i := 0; i < maxRetries; i++ {
        err = fn()
        if err == nil {
            return nil
        }
        
        // Exponential backoff
        delay := time.Duration(math.Pow(2, float64(i))) * time.Second
        log.Warnf("Attempt %d failed: %v. Retrying in %v...", i+1, err, delay)
        time.Sleep(delay)
    }
    return fmt.Errorf("max retries exceeded: %w", err)
}
```

### 4.2. Error Types

```go
type ErrorType int

const (
    ErrTypeProxy ErrorType = iota
    ErrTypeBrowser
    ErrTypeSelector
    ErrTypeTimeout
    ErrTypeCaptcha
    ErrTypeNetwork
)

type TaskError struct {
    Type    ErrorType
    Message string
    Err     error
}

func (e *TaskError) Error() string {
    return fmt.Sprintf("%s: %v", e.Message, e.Err)
}
```

### 4.3. Panic Recovery

```go
func (wp *WorkerPool) worker(workerID int) {
    defer func() {
        if r := recover(); r != nil {
            log.Errorf("Worker %d panicked: %v\n%s", workerID, r, debug.Stack())
        }
        wp.wg.Done()
    }()
    
    // Worker logic
}
```

---

## 5. Bot Detection Bypass Stratejileri

### 5.1. Stealth Mode Implementation

```go
func applyStealthMode(ctx context.Context) error {
    return chromedp.Run(ctx,
        // navigator.webdriver = undefined
        chromedp.Evaluate(`
            Object.defineProperty(navigator, 'webdriver', {
                get: () => undefined
            });
        `, nil),
        
        // Chrome detection bypass
        chromedp.Evaluate(`
            window.chrome = {
                runtime: {},
            };
        `, nil),
        
        // Permissions
        chromedp.Evaluate(`
            const originalQuery = window.navigator.permissions.query;
            window.navigator.permissions.query = (parameters) => (
                parameters.name === 'notifications' ?
                    Promise.resolve({ state: Notification.permission }) :
                    originalQuery(parameters)
            );
        `, nil),
    )
}
```

### 5.2. Fingerprint Randomization

```go
type Fingerprint struct {
    UserAgent  string
    Timezone   string
    Language   string
    Resolution [2]int
    Platform   string
}

func generateFingerprint() Fingerprint {
    userAgents := []string{
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36...",
        "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36...",
        // ... more
    }
    
    return Fingerprint{
        UserAgent:  userAgents[rand.Intn(len(userAgents))],
        Timezone:   "Europe/Istanbul",
        Language:   "tr-TR,tr;q=0.9,en-US;q=0.8,en;q=0.7",
        Resolution: [2]int{1920, 1080},
        Platform:   "Win32",
    }
}
```

### 5.3. Human-like Behavior

```go
func (hs *HumanSimulator) TypeHumanLike(text string) []chromedp.Action {
    actions := []chromedp.Action{}
    
    for _, char := range text {
        actions = append(actions,
            chromedp.SendKeys(`input[name="q"]`, string(char)),
            chromedp.Sleep(time.Duration(50+rand.Intn(150)) * time.Millisecond),
        )
    }
    
    return actions
}

func (hs *HumanSimulator) ScrollPattern() []chromedp.Action {
    return []chromedp.Action{
        chromedp.Evaluate(`window.scrollBy(0, 300 + Math.random() * 200)`, nil),
        chromedp.Sleep(time.Duration(1000+rand.Intn(2000)) * time.Millisecond),
        chromedp.Evaluate(`window.scrollBy(0, 400 + Math.random() * 300)`, nil),
        chromedp.Sleep(time.Duration(1500+rand.Intn(2500)) * time.Millisecond),
    }
}
```

---

## 6. Performans Optimizasyonu

### 6.1. Memory Management

```go
// Browser pool için reuse
type BrowserPool struct {
    pool chan *Browser
    max  int
}

func (bp *BrowserPool) Get() *Browser {
    select {
    case browser := <-bp.pool:
        return browser
    default:
        return NewBrowser()
    }
}

func (bp *BrowserPool) Put(browser *Browser) {
    select {
    case bp.pool <- browser:
    default:
        browser.Close() // Pool doluysa kapat
    }
}
```

### 6.2. Connection Pooling

```go
// HTTP client ile proxy connection reuse
func newHTTPClient(proxy *Proxy) *http.Client {
    transport := &http.Transport{
        Proxy: http.ProxyURL(proxy.URL),
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    }
    
    return &http.Client{
        Transport: transport,
        Timeout:   30 * time.Second,
    }
}
```

---

## 7. Test Stratejisi Detayları

### 7.1. Unit Test Örnekleri

```go
// internal/proxy/pool_test.go
func TestProxyPool_Get(t *testing.T) {
    proxies := []string{"http://proxy1.com:8080", "http://proxy2.com:8080"}
    pool := NewProxyPool(proxies, RoundRobin)
    
    proxy1, err := pool.Get()
    assert.NoError(t, err)
    assert.Equal(t, "proxy1.com", proxy1.Host)
    
    proxy2, err := pool.Get()
    assert.NoError(t, err)
    assert.Equal(t, "proxy2.com", proxy2.Host)
}

func TestProxyPool_Blacklist(t *testing.T) {
    pool := NewProxyPool([]string{"http://bad-proxy.com:8080"}, RoundRobin)
    
    proxy, _ := pool.Get()
    pool.Blacklist(proxy)
    
    _, err := pool.Get()
    assert.Error(t, err)
}
```

### 7.2. Integration Test

```go
// internal/serp/search_integration_test.go
// +build integration

func TestSearcher_EndToEnd(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }
    
    browser, _ := NewBrowser(BrowserOptions{Headless: true})
    defer browser.Close()
    
    searcher := NewSearcher(browser, defaultSelectors)
    err := searcher.Search("golang tutorial")
    assert.NoError(t, err)
    
    results, err := searcher.GetResults()
    assert.NoError(t, err)
    assert.NotEmpty(t, results)
}
```

### 7.3. Mock Kullanımı

```go
// internal/task/worker_test.go
type MockBrowser struct {
    mock.Mock
}

func (m *MockBrowser) Navigate(url string) error {
    args := m.Called(url)
    return args.Error(0)
}

func TestWorker_ExecuteTask(t *testing.T) {
    mockBrowser := new(MockBrowser)
    mockBrowser.On("Navigate", "https://google.com").Return(nil)
    
    // Test logic
}
```

---

## 8. Deployment ve DevOps

### 8.1. Makefile Komutları

```makefile
# Build
build:
	go build -o bin/serp-bot cmd/serp-bot/main.go

# Test
test:
	go test -v -cover ./...

test-unit:
	go test -v -short ./...

test-integration:
	go test -v -run Integration ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Lint
lint:
	golangci-lint run ./...

fmt:
	gofmt -s -w .

# Run
run:
	go run cmd/serp-bot/main.go start

# Clean
clean:
	rm -rf bin/ logs/ data/stats.json

# Install dependencies
deps:
	go mod download
	go mod tidy
```

### 8.2. CI/CD Pipeline (GitHub Actions)

```yaml
name: CI

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - name: Install dependencies
        run: make deps
      - name: Lint
        run: make lint
      - name: Test
        run: make test-coverage
      - name: Upload coverage
        uses: codecov/codecov-action@v2
```

---

## 9. Güvenlik Önlemleri

### 9.1. Environment Variables

```go
// Hassas bilgileri environment'tan al
func loadProxyCredentials() (string, string, error) {
    username := os.Getenv("PROXY_USERNAME")
    password := os.Getenv("PROXY_PASSWORD")
    
    if username == "" || password == "" {
        return "", "", errors.New("proxy credentials not set")
    }
    
    return username, password, nil
}
```

### 9.2. Log Sanitization

```go
func sanitizeLog(logMsg string) string {
    // URL'lerden password'leri temizle
    re := regexp.MustCompile(`://([^:]+):([^@]+)@`)
    return re.ReplaceAllString(logMsg, "://$1:***@")
}
```

---

## 10. Monitoring ve Debugging

### 10.1. Health Check

```go
func healthCheck() error {
    checks := []struct{
        name string
        fn   func() error
    }{
        {"Chrome Available", checkChrome},
        {"Config Valid", checkConfig},
        {"Proxy Pool", checkProxyPool},
        {"Disk Space", checkDiskSpace},
    }
    
    for _, check := range checks {
        if err := check.fn(); err != nil {
            return fmt.Errorf("%s failed: %w", check.name, err)
        }
    }
    
    return nil
}
```

### 10.2. Metrics Collection

```go
type Metrics struct {
    TasksProcessed   prometheus.Counter
    TaskDuration     prometheus.Histogram
    ProxySuccess     prometheus.Gauge
    ActiveWorkers    prometheus.Gauge
}
```

---

## Sonuç

Bu teknik tasarım dokümanı, Go-SERP-Bot projesinin implementasyon detaylarını içermektedir. Her modül, single responsibility principle'a uygun şekilde tasarlanmış ve test edilebilirlik göz önünde bulundurulmuştur.

**Önemli Notlar:**
- Tüm public fonksiyonlar GoDoc formatında comment'lenmeli
- Error handling her yerde yapılmalı
- Context kullanımı ile timeout ve cancellation sağlanmalı
- Resource leak'leri önlemek için defer kullanılmalı
- Test coverage %100 hedeflenmeli

