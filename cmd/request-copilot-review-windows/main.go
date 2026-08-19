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
	fmt.Printf("📄 Opening PR #%s\n", *prNumber)
	fmt.Printf("   URL: %s\n", prURL)
	fmt.Println()

	if err := chromedp.Run(taskCtx,
		chromedp.Navigate(prURL),
	); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error navigating to PR: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Wait for the page to load and for the button to be visible
	fmt.Println("⏳ Waiting for page load (3 seconds)...")
	time.Sleep(3 * time.Second)

	// Step 3: Try to find and click the review button
	// We'll try multiple selector methods to be robust
	fmt.Println("🔘 Locating 'Request Copilot Review' button...")
	fmt.Println()

	// Method 1: Click by ID using XPath (most reliable)
	buttonID := "re-request-review-copilot-pull-request-reviewer"
	fmt.Printf("   Attempt 1: XPath by ID (//*[@id=\"%s\"])\n", buttonID)

	err := chromedp.Run(taskCtx,
		chromedp.Click(fmt.Sprintf(`//*[@id="%s"]`, buttonID), chromedp.BySearch),
	)

	if err != nil {
		// Method 2: Try CSS selector by ID
		fmt.Printf("   Attempt 2: CSS selector (#%s)\n", buttonID)
		err = chromedp.Run(taskCtx,
			chromedp.Click(fmt.Sprintf(`#%s`, buttonID), chromedp.ByQuery),
		)

		if err != nil {
			// Method 3: Try finding button by name attribute
			fmt.Printf("   Attempt 3: Button by name attribute (re_request_reviewer_id)\n")
			err = chromedp.Run(taskCtx,
				chromedp.Click(`button[name="re_request_reviewer_id"]`, chromedp.ByQuery),
			)

			if err != nil {
				// Method 4: Try submitting the form directly
				fmt.Printf("   Attempt 4: Submit form (pull-request-reviewers-form-*)\n")
				err = chromedp.Run(taskCtx,
					chromedp.Submit(`form[id*="pull-request-reviewers-form"]`, chromedp.ByQuery),
				)

				if err != nil {
					fmt.Fprintf(os.Stderr, "✗ Error: Could not find or click the review button\n")
					fmt.Fprintf(os.Stderr, "  Last error: %v\n", err)
					fmt.Fprintf(os.Stderr, "\n")
					fmt.Fprintf(os.Stderr, "  Troubleshooting:\n")
					fmt.Fprintf(os.Stderr, "  1. The button may not be visible on the PR page\n")
					fmt.Fprintf(os.Stderr, "  2. GitHub's button selector may have changed\n")
					fmt.Fprintf(os.Stderr, "  3. You may need to scroll to reveal the review section\n")
					fmt.Fprintf(os.Stderr, "\n")
					fmt.Fprintf(os.Stderr, "  Expected button:\n")
					fmt.Fprintf(os.Stderr, "    ID: re-request-review-copilot-pull-request-reviewer\n")
					fmt.Fprintf(os.Stderr, "    Name: re_request_reviewer_id\n")
					fmt.Fprintf(os.Stderr, "    Type: submit\n")
					fmt.Fprintf(os.Stderr, "    Form: pull-request-reviewers-form-* (for re-requesting reviews)\n")
					os.Exit(1)
				}
			}
		}
	}

	// Step 4: Confirm success
	fmt.Println()
	fmt.Println("✅ Review request button clicked!")
	fmt.Println()
	fmt.Println("⏳ Waiting for GitHub to process the request...")
	time.Sleep(2 * time.Second)

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("  ✓ Copilot review request submitted successfully")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("GitHub will queue a new Copilot review for PR #%s\n", *prNumber)
	fmt.Println()
	fmt.Println("Expected timeline:")
	fmt.Println("  • 1-2 minutes: Copilot analyzes the PR")
	fmt.Println("  • Refresh GitHub to see the new review")
	fmt.Println()
}
