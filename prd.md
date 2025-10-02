# Proje Adı: Go-SERP-Bot Eğitim Aracı

**Doküman Durumu:** Taslak
**Oluşturulma Tarihi:** 1 Ekim 2025
**Versiyon:** 1.0
**Yazar:** Gemini

---

## 1. Giriş ve Amaç

### 1.1. Problem
Modern backend geliştirme, özellikle Go dilinde, eşzamanlılık (concurrency) ve yüksek performanslı ağ işlemleri gibi ileri düzey konuları anlamayı gerektirir. Bu teorik bilgileri pratiğe dökmenin en etkili yolu, bu konseptlerin temelini oluşturan gerçekçi bir proje geliştirmektir. Web otomasyonu ve scraping, bu becerileri kazanmak için ideal bir uygulama alanıdır.

### 1.2. Hedef ve Çözüm
Bu projenin temel amacı, Go dilini ve modern otomasyon tekniklerini öğrenmek için pratik bir araç geliştirmektir.

Geliştirilecek olan **Go-SERP-Bot**, komut satırından çalışan (CLI) bir uygulama olacak ve belirli anahtar kelimeler için arama motoru sonuç sayfalarında (SERP) otomatik olarak gezinti yapacaktır. Uygulama, hedef web sitelerini SERP sonuçlarında bulup, organik trafik simüle ederek (tıklama, sitede gezinme, bekleme) o kelimelerdeki sıralamalarını yükseltmeyi amaçlar. Bu süreç, baştan sona Go'nun güçlü eşzamanlılık özelliklerini ve `chromedp` gibi harici kütüphanelerin kullanımını öğretmek üzere tasarlanmıştır.

### 1.3. Hedef Kitle
* **Birincil:** Kendini Go ve otomasyon konularında geliştirmek isteyen yazılım geliştiriciler (Proje sahibi).

### 1.4. Ölçülebilir Başarı Kriterleri
* Go dilinde eşzamanlılık modellerinin (Goroutine, WaitGroup, Channels) başarılı bir şekilde uygulanması.
* `chromedp` kütüphanesi kullanılarak bir web tarayıcısının tam otomasyonunun gerçekleştirilmesi.
* Uygulamanın, harici bir konfigürasyon dosyasından (JSON veya YAML) görev parametrelerini (anahtar kelimeler, proxy'ler) okuyabilmesi.
* Uygulamanın, en az 10 paralel görevi aynı anda sorunsuz ve çökmeden yürütebilmesi.

---

## 2. Kapsam ve Özellikler

### 2.1. Versiyon 1.0 (Minimum Uygulanabilir Ürün - MVP)
Bu ilk sürüm, projenin temel işlevselliğini ve öğrenme hedeflerinin çekirdeğini oluşturacaktır.

* **Paralel Görev Yürütme:** Belirlenmiş bir anahtar kelime listesindeki her bir kelime için aynı anda bir arama görevi başlatma.
* **Temel Tarayıcı Otomasyonu:**
    * Google.com'a gitme.
    * Anahtar kelimeyi arama kutusuna yazma ve arama yapma.
    * Arama sonuçları sayfasında hedef web sitesini bulma (sayfa sayfa ilerleyerek).
    * Hedef web sitesine tıklama ve sayfada basit gezinme.
* **Sıralama Takibi:** Hedef web sitesinin SERP'teki mevcut pozisyonunu tespit etme ve loglama.
* **Basit Yapılandırma:** Aranacak anahtar kelimelerin, hedef web sitelerinin ve kullanılacak ücretsiz proxy listesinin kod içinde (hard-coded) tanımlanması.
* **Konsol Günlüğü (Logging):** Her görevin başlangıcı, bitişi, sıralama bilgisi ve olası hatalar hakkında temel bilgileri konsola yazdırma.
* **Temel Hata Yönetimi:** Başarısız işlemler için basit retry mekanizması (2-3 deneme).

### 2.2. Versiyon 1.1 (Yapılandırma ve Esneklik)
* **Harici Konfigürasyon:** Anahtar kelimelerin, hedef sitelerin, proxy listesinin ve diğer ayarların (örn: timeout süresi, retry sayısı) bir `config.json` veya `.env` dosyasından okunması.
* **Komut Satırı Argümanları (Flags):** Uygulamanın davranışını değiştirmek için temel komut satırı argümanları ekleme. Örneğin:
    * `--headless=false` : Tarayıcının görünür modda çalışmasını sağlamak.
    * `--workers=10` : Aynı anda çalışacak maksimum goroutine sayısını sınırlamak.
    * `--interval=300` : Döngüler arası bekleme süresi (saniye).
* **Sürekli Çalışma Modu:** Program, belirlenen aralıklarla sürekli olarak SERP kontrolü yapacak ve hedef siteye trafik gönderecek (sonsuz döngü). Her döngüde sıralama takibi yapılacak.
* **Gelişmiş Proxy Yönetimi:**
    * Ücretsiz proxy listesinden otomatik proxy çekme ve doğrulama.
    * Başarısız proxy'leri blacklist'e alma.
    * Round-robin veya random proxy rotation stratejisi.
    * Proxy başarısızlığında otomatik geçiş.
* **İstatistik Toplama:**
    * Her keyword için toplam tıklama sayısı.
    * Her keyword için bulunan sıralama geçmişi (zaman serisi).
    * Başarılı/başarısız görev oranları.
    * İstatistiklerin JSON formatında dosyaya kaydedilmesi (`stats.json`).
* **Dosya Bazlı Loglama:** Console'un yanı sıra log dosyasına da yazma (seviye: INFO, WARN, ERROR).

### 2.3. Versiyon 1.2 (Gelişmiş Otomasyon ve Taklit - Bot Detection Bypass)
* **User-Agent Rotasyonu:** Her tarayıcı oturumu için rastgele ve geçerli bir `User-Agent` ataması (güncel tarayıcı versiyonları).
* **İnsansı Etkileşim:**
    * Anahtar kelimeleri anında yapıştırmak yerine, harf harf, rastgele gecikmelerle (50-200ms) yazma.
    * Tıklama ve sayfa kaydırma eylemleri arasına rastgele bekleme süreleri ekleme (1-5 saniye).
    * Hedef sitede rastgele sayfa scroll, farklı linklere tıklama, sayfada 30-120 saniye bekleme.
    * Mouse movement simulasyonu (opsiyonel).
* **Gelişmiş Tarayıcı Fingerprinting:**
    * WebGL, Canvas, Audio fingerprint'lerini randomize etme.
    * Timezone, language, screen resolution gibi parametreleri gerçekçi kombinasyonlarla ayarlama.
    * JavaScript ile bot tespitini bypass etme (navigator.webdriver vb.).
* **CAPTCHA Yönetimi:**
    * CAPTCHA tespit edildiğinde kullanıcıya bildirim veya 2captcha gibi servislerle entegrasyon (opsiyonel).

### 2.4. Versiyon 1.3 (Production Özellikleri)
* **Authentication Proxy Desteği:** Ücretli residential/datacenter proxy'ler için username:password@host:port formatında authentication desteği.
* **Dashboard API:** İstatistikleri ve canlı durumu görüntülemek için basit REST API (opsiyonel).
* **Graceful Shutdown:** SIGTERM/SIGINT sinyalleri ile düzgün kapanma, yarım kalan görevlerin tamamlanması.
* **Health Check:** Uygulamanın sağlık durumunu kontrol eden endpoint veya komut.
* **Performance Optimizasyonu:**
    * Bellek kullanım optimizasyonu.
    * Context pooling ile tarayıcı instance'larını yeniden kullanma.

### 2.5. Kapsam Dışı (Şimdilik)
* Grafiksel Kullanıcı Arayüzü (GUI).
* Veritabanı entegrasyonu (PostgreSQL, MongoDB).
* Web tabanlı kontrol paneli.
* Bing, DuckDuckGo gibi diğer arama motorları desteği.

---

## 3. Teknik Gereksinimler

* **Programlama Dili:** Go (Golang) v1.21+
* **Ana Kütüphaneler:**
    * `github.com/chromedp/chromedp`: Tarayıcı otomasyonu için.
    * `sync`: Goroutine senkronizasyonu için (`WaitGroup`, `Mutex`).
    * `context`: Context yönetimi ve timeout kontrolü için.
    * `encoding/json`: Konfigürasyon ve istatistik dosyalarını okumak/yazmak için.
    * `github.com/joho/godotenv`: `.env` dosyası desteği için.
    * `github.com/sirupsen/logrus` veya `go.uber.org/zap`: Yapılandırılabilir loglama için.
    * `github.com/spf13/cobra`: CLI komutları ve flag yönetimi için.
    * `net/http`: Ücretsiz proxy listesi çekmek ve health check endpoint için.
* **Test Kütüphaneleri:**
    * `testing`: Go built-in test framework.
    * `github.com/stretchr/testify`: Assert ve mock desteği.
    * `github.com/chromedp/chromedp`: Test senaryoları için.
* **Linting ve Code Quality:**
    * `golangci-lint`: Kod kalitesi ve linting.
    * `gofmt`: Kod formatlama.
* **Platform:** Çapraz platform (Windows, macOS, Linux üzerinde çalışabilmeli).
* **Harici Bağımlılıklar:** 
    * Google Chrome veya Chromium tabanlı tarayıcı (Headless Chrome).
    * Minimum 4GB RAM (paralel görevler için).
    * İnternet bağlantısı.

---

## 4. Kullanıcı Akışı

### 4.1. İlk Kurulum ve Başlatma
1.  Kullanıcı, `config.json` dosyasını aramak istediği anahtar kelimeler, hedef web siteleri ve kullanmak istediği proxy listesi ile düzenler.
2.  Kullanıcı, `.env` dosyasında gerekli ayarları yapar (örn: `HEADLESS=true`, `WORKERS=5`, `INTERVAL=300`).
3.  Kullanıcı, terminal üzerinden uygulamayı çalıştırır: `go run cmd/serp-bot/main.go` veya derlenmiş dosyayı `./serp-bot start`.

### 4.2. Program Çalışma Döngüsü (Sonsuz Döngü)
4.  Uygulama başlar ve konsola başlangıç mesajını yazar.
5.  `config.json` ve `.env` dosyalarını okur ve görevleri hafızaya yükler.
6.  **Ana Döngü Başlangıcı:**
    * Her bir anahtar kelime için paralel olarak bir Goroutine (ve dolayısıyla bir tarayıcı örneği) başlatır.
    * Worker pool deseni ile maksimum paralel görev sayısını sınırlar.
7.  **Her Görev İçin:**
    * Proxy havuzundan bir proxy seçer ve tarayıcıyı başlatır.
    * Google.com'a gider ve anahtar kelimeyi arar.
    * SERP sonuçlarında hedef web sitesini arar (sayfa sayfa ilerleyerek).
    * Hedef sitenin sıralamasını tespit eder ve loglar/kaydeder.
    * Hedef siteye tıklar ve sitede insan gibi davranış sergiler (scroll, bekle, farklı sayfalara geç).
    * 30-120 saniye bekledikten sonra tarayıcıyı kapatır.
8.  **Döngü Sonrası:**
    * Tüm görevler tamamlandığında istatistikleri günceller (`stats.json`).
    * Konsola özet bilgi yazar (kaç görev başarılı, sıralama değişimleri).
    * Belirlenen süre kadar bekler (örn: 5 dakika).
    * **Döngüyü yeniden başlatır (6. adıma geri döner).**
9.  Program, kullanıcı Ctrl+C veya SIGTERM ile durdurana kadar çalışmaya devam eder.

---

## 5. Riskler ve Çözüm Önerileri

* **Risk:** Hedef web sitesinin (Google) HTML yapısını değiştirmesi, seçicilerin (selectors) çalışmamasına neden olabilir.
    * **Çözüm:** Seçicileri kodun içinde sabit olarak yazmak yerine konfigürasyon dosyasından okunabilir hale getirmek. Hata yönetimini (error handling) güçlü tutmak. Fallback selector stratejisi uygulamak.
* **Risk:** Bot tespit sistemleri tarafından engellenme (reCAPTCHA, Cloudflare vb.).
    * **Çözüm:** Eğitim projesi olduğu için bu risk kabul edilebilir. Ancak ileri düzey öğrenme hedefi olarak, insansı taklit özellikleri (Bkz. v1.2) ve kaliteli (residential) proxy kullanımı araştırılabilir. CAPTCHA tespit mekanizması eklemek.
* **Risk:** Çok fazla paralel görev başlatıldığında sistem kaynaklarının (CPU/RAM) aşırı tüketilmesi.
    * **Çözüm:** "Worker Pool" tasarım deseni kullanarak aynı anda çalışacak goroutine sayısını sınırlayan bir mekanizma geliştirmek (Bkz. v1.1'deki `--workers` flag'i). Bellek limitleri koymak.
* **Risk:** Ücretsiz proxy'lerin güvenilmez olması, çoğunun çalışmaması.
    * **Çözüm:** Proxy validasyon mekanizması eklemek. Başarısız proxy'leri blacklist'e almak. Birden fazla ücretsiz proxy kaynağı kullanmak.
* **Risk:** IP yasaklanması veya geçici ban.
    * **Çözüm:** Rate limiting eklemek. Her IP için günlük maksimum istek sayısı sınırı. Proxy rotasyonunu aktif kullanmak.
* **Risk:** Etik ve yasal sorunlar (ToS ihlali).
    * **Çözüm:** Bu proje sadece eğitim amaçlıdır. Gerçek kullanımda Google'ın kullanım şartlarını ihlal edebilir. Sorumluluk kullanıcıya aittir.
* **Risk:** Sonsuz döngüde hata birikimi, memory leak.
    * **Çözüm:** Her döngüde tüm kaynakları düzgün temizlemek (defer kullanımı). Periyodik health check. Panic recovery mekanizması.

---

## 6. Proje Yapısı

```
go-bot/
├── cmd/
│   └── serp-bot/
│       └── main.go                 # Entry point, CLI komutları
├── internal/
│   ├── config/
│   │   ├── config.go              # Konfigürasyon struct ve loader
│   │   └── config_test.go
│   ├── browser/
│   │   ├── browser.go             # chromedp wrapper
│   │   ├── actions.go             # Tarayıcı aksiyonları (scroll, click vb.)
│   │   ├── stealth.go             # Bot detection bypass
│   │   └── browser_test.go
│   ├── serp/
│   │   ├── search.go              # SERP arama ve ranking kontrolü
│   │   ├── navigation.go          # Sayfa navigasyonu
│   │   └── serp_test.go
│   ├── proxy/
│   │   ├── proxy.go               # Proxy yönetimi
│   │   ├── pool.go                # Proxy pool ve rotation
│   │   ├── validator.go           # Proxy validation
│   │   └── proxy_test.go
│   ├── task/
│   │   ├── task.go                # Task struct ve yönetimi
│   │   ├── worker.go              # Worker pool implementasyonu
│   │   ├── scheduler.go           # Sonsuz döngü ve zamanlama
│   │   └── task_test.go
│   ├── stats/
│   │   ├── stats.go               # İstatistik toplama ve kaydetme
│   │   └── stats_test.go
│   └── logger/
│       ├── logger.go              # Loglama utility
│       └── logger_test.go
├── pkg/
│   └── utils/
│       ├── random.go              # Rastgele değer üreticiler
│       └── utils_test.go
├── configs/
│   ├── config.json.example        # Örnek konfigürasyon
│   └── selectors.json             # Google selectors
├── logs/                          # Log dosyaları (gitignore)
├── data/                          # İstatistik dosyaları (gitignore)
│   └── stats.json
├── .env.example                   # Örnek environment dosyası
├── .gitignore
├── go.mod
├── go.sum
├── Makefile                       # Build, test, lint komutları
├── README.md                      # Proje dokümantasyonu
├── prd.md                         # Bu doküman
└── TASKLIST.md                    # Geliştirme görev listesi
```

---

## 7. Test Stratejisi

### 7.1. Test Hedefi
* **Hedef Coverage:** %100 [[memory:6562668]]
* **Test Türleri:** Unit test, integration test

### 7.2. Unit Test
Her bir package için ayrı test dosyaları:
* `config_test.go`: Konfigürasyon okuma/validasyon testleri
* `proxy_test.go`: Proxy rotation, validation testleri
* `browser_test.go`: Tarayıcı aksiyonları testleri (mock ile)
* `task_test.go`: Worker pool, görev yönetimi testleri
* `stats_test.go`: İstatistik hesaplama ve kaydetme testleri

### 7.3. Integration Test
* Gerçek tarayıcı ile end-to-end senaryolar
* Proxy ile bağlantı testleri
* SERP arama ve sıralama tespiti testleri

### 7.4. Test Komutları
```bash
make test           # Tüm testleri çalıştır
make test-unit      # Sadece unit testler
make test-coverage  # Coverage raporu
make test-integration  # Integration testler
```

---

## 8. Performans Metrikleri

### 8.1. Hedef Metrikler
* **Paralel Görev Kapasitesi:** Minimum 10, maksimum 50 eşzamanlı görev
* **Bellek Kullanımı:** Görev başına maksimum 200MB
* **CPU Kullanımı:** Görev başına maksimum %20 (4 çekirdek üzerinde)
* **Görev Tamamlanma Süresi:** 60-180 saniye arası (sitede bekleme hariç)
* **Hata Toleransı:** %80+ başarı oranı (ücretsiz proxy ile)

### 8.2. İzlenecek Metrikler
* Toplam görev sayısı
* Başarılı/başarısız görev oranı
* Ortalama görev tamamlanma süresi
* Proxy başarı oranları
* SERP sıralama değişimleri
* Bellek ve CPU kullanımı

---

## 9. Geliştirme Süreci

### 9.1. Development Workflow
1. Feature branch oluştur (`feature/xxx`)
2. Kod yaz ve testlerini yaz (TDD yaklaşımı)
3. `make lint` ile kod kalitesini kontrol et
4. `make test` ile tüm testleri çalıştır
5. Coverage %100 olduğundan emin ol
6. Pull request oluştur
7. Code review sonrası merge

### 9.2. Kod Kalitesi Standartları
* Her fonksiyon için comment/docstring yazılmalı
* Hata yönetimi her zaman yapılmalı (nil check)
* Context kullanımı (timeout, cancellation)
* Defer ile resource cleanup
* Goroutine leak olmaması
* Single Responsibility Principle [[memory:3743208]]

### 9.3. Linting Rules
* `golangci-lint` ile otomatik kontrol
* `gofmt` ile otomatik formatlama
* Import organization
* Unused variable/import kontrolü
* Cyclomatic complexity limiti (max 15)

---

## 10. Dokümantasyon Gereksinimleri

### 10.1. README.md İçeriği
* Proje tanımı ve amaç
* Kurulum adımları
* Konfigürasyon örnekleri
* Kullanım kılavuzu
* CLI komutları ve flagler
* Troubleshooting
* Lisans ve disclaimer (eğitim amaçlı)

### 10.2. Kod Dokümantasyonu
* Her public fonksiyon için GoDoc formatında comment
* Karmaşık algoritmalar için inline comment
* Package düzeyinde overview comment

---

## 11. Güvenlik ve Etik Konular

### 11.1. Güvenlik
* Hassas bilgileri (proxy credentials) `.env` dosyasında sakla
* `.env` dosyasını `.gitignore`'a ekle
* API key'leri environment variable olarak al
* Log dosyalarında hassas bilgi loglanmaması

### 11.2. Etik ve Yasal Uyarılar
⚠️ **ÖNEMLİ UYARI:**
* Bu proje **sadece eğitim amaçlıdır**.
* Google'ın Terms of Service'ini ihlal edebilir.
* Ticari kullanım için uygun değildir.
* Kullanıcı, yasalar ve ToS'a uymaktan sorumludur.
* Aşırı kullanım IP ban'ına sebep olabilir.
* Sorumluluk kullanıcıya aittir.

---

## 12. Gelecek Geliştirmeler (v2.0+)

* Bing, DuckDuckGo, Yandex desteği
* Web dashboard (React + REST API)
* PostgreSQL/MongoDB ile istatistik saklama
* Docker container desteği
* Kubernetes deployment
* Prometheus/Grafana entegrasyonu
* Telegram bot ile bildirimler
* Machine learning ile CAPTCHA bypass
* Distributed architecture (multi-node)