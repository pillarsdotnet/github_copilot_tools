package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	prNumber := flag.String("pr", "", "GitHub PR number (e.g., 116)")
	flag.Parse()

	if *prNumber == "" {
		fmt.Fprintf(os.Stderr, "Usage: request-copilot-review-windows -pr <PR_NUMBER>\n")
		fmt.Fprintf(os.Stderr, "Example: request-copilot-review-windows -pr 116\n")
		os.Exit(1)
	}

	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("  GitHub Copilot Review Request (Windows Side)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Connect to the Chrome debugging session
	// Chrome is listening on 127.0.0.1:9222 on the Windows side
	fmt.Println("🔍 Connecting to Chrome debugging session on port 9222...")
	fmt.Println()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use the devtools URL to connect to the existing Chrome instance
	allocatorCtx, cancel := chromedp.NewRemoteAllocator(ctx, "http://127.0.0.1:9222")
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocatorCtx)
	defer cancel()

	// Step 1: Navigate to the PR
	prURL := fmt.Sprintf("https://github.com/owner/repo/pull/%s", *prNumber)
	fmt.Printf("📄 Opening PR #%s: %s\n", *prNumber, prURL)
	fmt.Println()

	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(prURL),
	); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error navigating to PR: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Wait for the page to load
	fmt.Println("⏳ Waiting for page load...")
	time.Sleep(3 * time.Second)

	// Step 3: Find and click the review button
	// XPath: //*[@id="re-request-review-copilot-pull-request-reviewer"]
	fmt.Println("🔘 Locating 'Request Copilot Review' button...")
	fmt.Println("   XPath: //*[@id=\"re-request-review-copilot-pull-request-reviewer\"]")
	fmt.Println()

	if err := chromedp.Run(taskCtx,
		chromedp.Click(`//*[@id="re-request-review-copilot-pull-request-reviewer"]`, chromedp.BySearch),
	); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error clicking review button: %v\n", err)
		fmt.Fprintf(os.Stderr, "  Note: Button may not be visible or XPath may have changed\n")
		fmt.Fprintf(os.Stderr, "  Check GitHub PR interface for current button location\n")
		os.Exit(1)
	}

	// Step 4: Confirm success
	fmt.Println("✅ Review request button clicked!")
	fmt.Println()
	fmt.Println("⏳ Waiting for GitHub to process the request...")
	time.Sleep(2 * time.Second)

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("  ✓ Copilot review request submitted successfully")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("GitHub will queue a new Copilot review for PR #" + *prNumber)
	fmt.Println()
}
