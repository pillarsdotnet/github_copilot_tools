package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
	"unsafe"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// Windows API constants for keyboard input
const (
	INPUT_KEYBOARD = 1
	KEYEVENTF_KEYUP = 0x0002
	VK_CONTROL = 0xA2
	VK_L = 0x4C
	VK_RETURN = 0x0D
)

// INPUT structure for SendInput
type KeyboardInput struct {
	Type uint32
	Ki   KeyboardInputData
	Pad  uint64
}

type KeyboardInputData struct {
	VKey         uint16
	Scan         uint16
	Flags        uint32
	Time         uint32
	ExtraInfo    uintptr
}

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procSendInput     = user32.NewLazyProc("SendInput")
	procFindWindowW   = user32.NewLazyProc("FindWindowW")
	procSetForeground = user32.NewLazyProc("SetForegroundWindow")
	procGetForeground = user32.NewLazyProc("GetForegroundWindow")
)

func main() {
	prNumber := flag.String("pr", "", "GitHub PR number (e.g., 116)")
	owner := flag.String("owner", "", "GitHub repository owner (e.g., owner)")
	repo := flag.String("repo", "", "GitHub repository name (e.g., repo)")
	flag.Parse()

	if *prNumber == "" || *owner == "" || *repo == "" {
		fmt.Fprintf(os.Stderr, "Usage: request-copilot-review-windows -pr <PR_NUMBER> -owner <OWNER> -repo <REPO>\n")
		fmt.Fprintf(os.Stderr, "Example: request-copilot-review-windows -pr 116 -owner owner -repo repo\n")
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

	// Step 1: Find Chrome window
	fmt.Println("🔍 Finding Chrome window...")
	chromeHWND := findChromeWindow()
	if chromeHWND == 0 {
		fmt.Fprintf(os.Stderr, "✗ Error: Chrome window not found\n")
		fmt.Fprintf(os.Stderr, "  Make sure Chrome is open and visible\n")
		os.Exit(1)
	}
	fmt.Printf("✓ Found Chrome window (HWND: %d)\n", chromeHWND)
	fmt.Println()

	// Step 2: Navigate to PR
	prURL := fmt.Sprintf("https://github.com/%s/%s/pull/%s", *owner, *repo, *prNumber)
	fmt.Printf("📄 Navigating to PR #%s\n", *prNumber)
	fmt.Printf("   Repository: %s/%s\n", *owner, *repo)
	fmt.Printf("   URL: %s\n", prURL)
	fmt.Println()

	if err := navigateToPR(chromeHWND, prURL); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error navigating: %v\n", err)
		os.Exit(1)
	}

	// Step 3: Wait for page load
	fmt.Println("⏳ Waiting for page load (5 seconds)...")
	time.Sleep(5 * time.Second)
	fmt.Println()

	// Step 4: Find and click the button
	fmt.Println("🔘 Locating and clicking review button...")
	fmt.Println("   Button ID: re-request-review-copilot-pull-request-reviewer")
	fmt.Println()

	if err := clickReviewButton(); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error clicking button: %v\n", err)
		fmt.Fprintf(os.Stderr, "  Note: Button may not be visible or UI structure may have changed\n")
		os.Exit(1)
	}

	fmt.Println("✅ Review request button clicked!")
	fmt.Println()

	// Step 5: Confirm success
	fmt.Println("⏳ Waiting for GitHub to process request (2 seconds)...")
	time.Sleep(2 * time.Second)

	fmt.Println()
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("  ✓ Copilot review request submitted successfully")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Printf("GitHub will queue a new Copilot review for %s/%s#%s\n", *owner, *repo, *prNumber)
	fmt.Println()
	fmt.Println("Expected timeline:")
	fmt.Println("  • 1-2 minutes: Copilot analyzes the PR")
	fmt.Println("  • Refresh GitHub to see the new review")
	fmt.Println()
}

// findChromeWindow finds the Chrome browser window
func findChromeWindow() uintptr {
	// Look for Chrome windows by class name
	classNames := []string{
		"Chrome_WidgetWin_1",
		"Chrome_WidgetWin_0",
	}

	for _, className := range classNames {
		ret, _, _ := procFindWindowW.Call(
			uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className))),
			0)
		if ret != 0 {
			return ret
		}
	}

	return 0
}

// navigateToPR navigates to the PR using keyboard shortcuts and clipboard
func navigateToPR(hwnd uintptr, url string) error {
	// Set Chrome window to foreground
	procSetForeground.Call(hwnd)
	time.Sleep(300 * time.Millisecond)

	fmt.Println("  Focusing address bar (Ctrl+L)...")

	// Send Ctrl+L to focus address bar
	if err := sendKeyPress(VK_CONTROL, true); err != nil {
		return fmt.Errorf("failed to send Ctrl key down: %w", err)
	}
	if err := sendKeyPress(VK_L, false); err != nil {
		return fmt.Errorf("failed to send L key: %w", err)
	}
	if err := sendKeyUp(VK_CONTROL); err != nil {
		return fmt.Errorf("failed to send Ctrl key up: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Println("  Typing URL...")

	// Send the URL one character at a time using keyboard events
	for _, ch := range url {
		if err := sendChar(uint16(ch)); err != nil {
			return fmt.Errorf("failed to send character '%c': %w", ch, err)
		}
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(300 * time.Millisecond)

	fmt.Println("  Pressing Enter to navigate...")

	// Send Enter to navigate
	if err := sendKeyPress(VK_RETURN, false); err != nil {
		return fmt.Errorf("failed to send Enter: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	return nil
}

// sendKeyPress sends a key down and up
func sendKeyPress(vkey uint16, isControl bool) error {
	// Send key down
	input := KeyboardInput{
		Type: INPUT_KEYBOARD,
		Ki: KeyboardInputData{
			VKey: vkey,
			Scan: 0,
			Flags: 0,
		},
	}

	ret, _, err := procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input))

	if ret == 0 {
		return fmt.Errorf("SendInput failed for key down: %w", err)
	}

	time.Sleep(50 * time.Millisecond)

	// Send key up
	input.Ki.Flags = KEYEVENTF_KEYUP
	ret, _, err = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input))

	if ret == 0 {
		return fmt.Errorf("SendInput failed for key up: %w", err)
	}

	return nil
}

// sendKeyUp sends a key up event
func sendKeyUp(vkey uint16) error {
	input := KeyboardInput{
		Type: INPUT_KEYBOARD,
		Ki: KeyboardInputData{
			VKey: vkey,
			Scan: 0,
			Flags: KEYEVENTF_KEYUP,
		},
	}

	ret, _, err := procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input))

	if ret == 0 {
		return fmt.Errorf("SendInput failed: %w", err)
	}

	return nil
}

// sendChar sends a character using keyboard event
func sendChar(char uint16) error {
	input := KeyboardInput{
		Type: INPUT_KEYBOARD,
		Ki: KeyboardInputData{
			VKey: 0,
			Scan: char,
			Flags: 0x0004, // KEYEVENTF_UNICODE
		},
	}

	ret, _, err := procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input))

	if ret == 0 {
		return fmt.Errorf("SendInput failed: %w", err)
	}

	time.Sleep(50 * time.Millisecond)

	input.Ki.Flags = KEYEVENTF_KEYUP | 0x0004
	ret, _, err = procSendInput.Call(
		1,
		uintptr(unsafe.Pointer(&input)),
		unsafe.Sizeof(input))

	if ret == 0 {
		return fmt.Errorf("SendInput failed for char up: %w", err)
	}

	return nil
}

// clickReviewButton finds and clicks the review button using UI Automation
func clickReviewButton() error {
	fmt.Println("  Initializing UI Automation...")

	// Create UI Automation object
	unknown, err := oleutil.CreateObject("UIAutomationCore.CUIAutomation8")
	if err != nil {
		return fmt.Errorf("failed to create UIAutomation object: %w", err)
	}
	if unknown == nil {
		return fmt.Errorf("UIAutomation object is nil")
	}
	defer unknown.Release()

	// Get root element
	rootResult, err := oleutil.CallMethod(unknown, "GetRootElement")
	if err != nil {
		return fmt.Errorf("failed to get root element: %w", err)
	}
	root := rootResult.ToIDispatch()
	if root == nil {
		return fmt.Errorf("root element is nil")
	}
	defer root.Release()

	fmt.Println("  Searching for button by AutomationId...")

	// Create a PropertyCondition to find button by AutomationId
	// AutomationId property ID is 30011
	automationIDPropID := 30011
	buttonID := "re-request-review-copilot-pull-request-reviewer"

	// Create condition: (AutomationId == "re-request-review-copilot-pull-request-reviewer")
	conditionResult, err := oleutil.CallMethod(unknown, "CreatePropertyCondition",
		automationIDPropID, buttonID)
	if err != nil {
		return fmt.Errorf("failed to create property condition: %w", err)
	}
	condition := conditionResult.ToIDispatch()
	if condition == nil {
		return fmt.Errorf("property condition is nil")
	}
	defer condition.Release()

	// Find the first element matching the condition (TreeScope_Subtree = 7)
	findResult, err := oleutil.CallMethod(root, "FindFirst", 7, condition)
	if err != nil {
		return fmt.Errorf("failed to find element: %w", err)
	}

	element := findResult.ToIDispatch()
	if element == nil {
		return fmt.Errorf("button element not found - may not be visible or ID mismatch")
	}
	defer element.Release()

	fmt.Println("  Found button element, invoking click...")

	// Try to invoke the Invoke pattern (for buttons)
	invokeResult, err := oleutil.CallMethod(element, "GetCurrentPattern", 5) // 5 = InvokePattern
	if err == nil && invokeResult != nil {
		invokePattern := invokeResult.ToIDispatch()
		if invokePattern != nil {
			defer invokePattern.Release()

			_, err := oleutil.CallMethod(invokePattern, "Invoke")
			if err == nil {
				return nil
			}
		}
	}

	// Fallback: Try Click method directly
	_, err = oleutil.CallMethod(element, "Click")
	if err == nil {
		return nil
	}

	return fmt.Errorf("could not invoke button - tried both Invoke pattern and Click")
}
