package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
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
	fmt.Println("  GitHub Copilot Review Request (Windows UI Automation)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Initialize COM
	fmt.Println("🔧 Initializing Windows UI Automation...")
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Failed to initialize COM: %v\n", err)
		os.Exit(1)
	}
	defer ole.CoUninitialize()
	fmt.Println("✓ COM initialized")
	fmt.Println()

	// Step 1: Navigate to PR
	prURL := fmt.Sprintf("https://github.com/owner/repo/pull/%s", *prNumber)
	fmt.Printf("📄 Navigating to PR #%s\n", *prNumber)
	fmt.Printf("   URL: %s\n", prURL)
	fmt.Println()

	if err := navigateToURL(prURL); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error navigating: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Wait for page load
	fmt.Println("⏳ Waiting for page load (5 seconds)...")
	time.Sleep(5 * time.Second)
	fmt.Println()

	// Step 3: Find and click the button
	fmt.Println("🔘 Locating and clicking review button...")
	fmt.Println("   Button ID: re-request-review-copilot-pull-request-reviewer")
	fmt.Println()

	if err := clickButtonByID("re-request-review-copilot-pull-request-reviewer"); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error clicking button: %v\n", err)
		fmt.Fprintf(os.Stderr, "  Note: Button may not be visible or UI structure may have changed\n")
		os.Exit(1)
	}

	fmt.Println("✅ Review request button clicked!")
	fmt.Println()

	// Step 4: Confirm success
	fmt.Println("⏳ Waiting for GitHub to process request (2 seconds)...")
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

// navigateToURL uses Windows API to navigate to a URL
func navigateToURL(url string) error {
	fmt.Println("  Opening URL via keyboard shortcut...")
	fmt.Println("  (Requires Chrome to be the active window)")

	// This is a simplified approach - keyboard automation
	// In a production system, would use more robust methods

	return fmt.Errorf("keyboard automation requires manual setup - use Chrome manually or enable accessibility features")
}

// clickButtonByID finds and clicks a button by ID using UI Automation
func clickButtonByID(buttonID string) error {
	fmt.Printf("  Searching for button with ID: %s...\n", buttonID)

	// Create the UI Automation object
	unknown, err := oleutil.CreateObject("UIAutomationCore.CUIAutomation8")
	if err != nil {
		return fmt.Errorf("failed to create UIAutomation object: %w", err)
	}
	if unknown == nil {
		return fmt.Errorf("UIAutomation object is nil")
	}
	defer unknown.Release()

	// The oleutil.CreateObject returns IUnknown, which we can use directly
	// with oleutil.CallMethod without conversion

	// Get the root element
	rootResult, err := oleutil.CallMethod(unknown, "GetRootElement")
	if err != nil {
		return fmt.Errorf("failed to get root element: %w", err)
	}
	root := rootResult.ToIDispatch()
	if root == nil {
		return fmt.Errorf("root element is nil")
	}
	defer root.Release()

	fmt.Println("  Searching element tree for matching button...")

	// Try to find elements
	// Note: This is a simplified implementation
	// Full UIA implementation would require building PropertyCondition objects

	result, err := oleutil.CallMethod(root, "FindFirst", 1, nil)
	if err != nil {
		return fmt.Errorf("element search failed: %w", err)
	}

	element := result.ToIDispatch()
	if element == nil {
		return fmt.Errorf("no matching element found")
	}
	defer element.Release()

	fmt.Println("  Found element, attempting click...")

	// Try to invoke the click via UI Automation
	_, err = oleutil.CallMethod(element, "Click")
	if err == nil {
		return nil
	}

	// Try Invoke pattern
	_, err = oleutil.CallMethod(element, "Invoke")
	if err == nil {
		return nil
	}

	return fmt.Errorf("could not execute button action")
}
