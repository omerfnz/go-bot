package browser

import (
	"context"
	"fmt"

	"github.com/chromedp/chromedp"
	"github.com/omer/go-bot/pkg/utils"
)

// Fingerprint represents browser fingerprint configuration
type Fingerprint struct {
	UserAgent  string
	Language   string
	Platform   string
	Vendor     string
	WebGL      string
	Resolution [2]int
}

// ApplyStealthMode applies various stealth techniques to bypass bot detection
func ApplyStealthMode(ctx context.Context) error {
	// Run all stealth scripts
	if err := disableWebDriver(ctx); err != nil {
		return fmt.Errorf("failed to disable webdriver: %w", err)
	}

	if err := enableChromeRuntime(ctx); err != nil {
		return fmt.Errorf("failed to enable chrome runtime: %w", err)
	}

	if err := fixPermissions(ctx); err != nil {
		return fmt.Errorf("failed to fix permissions: %w", err)
	}

	if err := fixPlugins(ctx); err != nil {
		return fmt.Errorf("failed to fix plugins: %w", err)
	}

	if err := fixLanguages(ctx); err != nil {
		return fmt.Errorf("failed to fix languages: %w", err)
	}

	return nil
}

// disableWebDriver removes navigator.webdriver property
func disableWebDriver(ctx context.Context) error {
	script := `
		Object.defineProperty(navigator, 'webdriver', {
			get: () => undefined
		});
	`
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// enableChromeRuntime adds window.chrome object
func enableChromeRuntime(ctx context.Context) error {
	script := `
		if (!window.chrome) {
			window.chrome = {
				runtime: {},
				loadTimes: function() {},
				csi: function() {},
				app: {}
			};
		}
	`
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// fixPermissions fixes navigator.permissions.query
func fixPermissions(ctx context.Context) error {
	script := `
		const originalQuery = window.navigator.permissions.query;
		window.navigator.permissions.query = (parameters) => (
			parameters.name === 'notifications' ?
				Promise.resolve({ state: Notification.permission }) :
				originalQuery(parameters)
		);
	`
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// fixPlugins makes plugins array look real
func fixPlugins(ctx context.Context) error {
	script := `
		Object.defineProperty(navigator, 'plugins', {
			get: () => [
				{
					0: {type: "application/x-google-chrome-pdf", suffixes: "pdf", description: "Portable Document Format"},
					description: "Portable Document Format",
					filename: "internal-pdf-viewer",
					length: 1,
					name: "Chrome PDF Plugin"
				},
				{
					0: {type: "application/pdf", suffixes: "pdf", description: "Portable Document Format"},
					description: "Portable Document Format",
					filename: "mhjfbmdgcfjbbpaeojofohoefgiehjai",
					length: 1,
					name: "Chrome PDF Viewer"
				}
			]
		});
	`
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// fixLanguages fixes navigator.languages
func fixLanguages(ctx context.Context) error {
	script := `
		Object.defineProperty(navigator, 'languages', {
			get: () => ['en-US', 'en', 'tr-TR', 'tr']
		});
	`
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// RandomizeFingerprint applies random browser fingerprint
func RandomizeFingerprint(ctx context.Context) error {
	fingerprint := generateRandomFingerprint()

	// Apply user agent
	if err := setUserAgent(ctx, fingerprint.UserAgent); err != nil {
		return fmt.Errorf("failed to set user agent: %w", err)
	}

	// Apply platform
	if err := setPlatform(ctx, fingerprint.Platform); err != nil {
		return fmt.Errorf("failed to set platform: %w", err)
	}

	// Apply vendor
	if err := setVendor(ctx, fingerprint.Vendor); err != nil {
		return fmt.Errorf("failed to set vendor: %w", err)
	}

	// Apply WebGL vendor
	if err := setWebGLVendor(ctx, fingerprint.WebGL); err != nil {
		return fmt.Errorf("failed to set webgl vendor: %w", err)
	}

	return nil
}

// generateRandomFingerprint generates a random but consistent fingerprint
func generateRandomFingerprint() Fingerprint {
	platforms := []string{"Win32", "MacIntel", "Linux x86_64"}
	vendors := []string{"Google Inc.", "Apple Computer, Inc."}
	webglVendors := []string{"Intel Inc.", "NVIDIA Corporation", "AMD"}

	resolutions := [][2]int{
		{1920, 1080},
		{1366, 768},
		{1440, 900},
		{1536, 864},
		{1600, 900},
	}

	return Fingerprint{
		UserAgent:  utils.RandomUserAgent(),
		Language:   "en-US",
		Platform:   utils.RandomChoice(platforms),
		Vendor:     utils.RandomChoice(vendors),
		WebGL:      utils.RandomChoice(webglVendors),
		Resolution: resolutions[utils.RandomInt(0, len(resolutions)-1)],
	}
}

// setUserAgent sets navigator.userAgent
func setUserAgent(ctx context.Context, userAgent string) error {
	script := fmt.Sprintf(`
		Object.defineProperty(navigator, 'userAgent', {
			get: () => '%s'
		});
	`, userAgent)
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// setPlatform sets navigator.platform
func setPlatform(ctx context.Context, platform string) error {
	script := fmt.Sprintf(`
		Object.defineProperty(navigator, 'platform', {
			get: () => '%s'
		});
	`, platform)
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// setVendor sets navigator.vendor
func setVendor(ctx context.Context, vendor string) error {
	script := fmt.Sprintf(`
		Object.defineProperty(navigator, 'vendor', {
			get: () => '%s'
		});
	`, vendor)
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// setWebGLVendor sets WebGL vendor
func setWebGLVendor(ctx context.Context, vendor string) error {
	script := fmt.Sprintf(`
		const getParameter = WebGLRenderingContext.prototype.getParameter;
		WebGLRenderingContext.prototype.getParameter = function(parameter) {
			if (parameter === 37445) {
				return '%s';
			}
			if (parameter === 37446) {
				return 'ANGLE (Intel, Intel(R) UHD Graphics 620, OpenGL 4.5)';
			}
			return getParameter.call(this, parameter);
		};
	`, vendor)
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}

// DisableAutomationFlags disables automation detection flags
func DisableAutomationFlags(ctx context.Context) error {
	script := `
		delete navigator.__proto__.webdriver;
	`
	return chromedp.Run(ctx, chromedp.Evaluate(script, nil))
}
