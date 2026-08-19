package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

// Windows API constants for keyboard input
const (
	INPUT_KEYBOARD = 1
	KEYEVENTF_KEYUP = 0x0002
	VK_CONTROL = 0xA2
	VK_SHIFT = 0xA0
	VK_L = 0x4C
	VK_V = 0x56
	VK_I = 0x49
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
	user32            = syscall.MustLoadDLL("user32.dll")
	procSendInput     = user32.MustFindProc("SendInput")
	procFindWindowW   = user32.MustFindProc("FindWindowW")
	procSetForeground = user32.MustFindProc("SetForegroundWindow")
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
	fmt.Println("  GitHub Copilot Review Request (DevTools Console)")
	fmt.Println("═══════════════════════════════════════════════════════════════")
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

	// Step 4: Click button via DevTools Console
	fmt.Println("🔘 Opening DevTools Console and executing JavaScript...")
	fmt.Println("   Method: Ctrl+Shift+I → Console → Paste → Execute")
	fmt.Println()

	if err := clickReviewButtonViaConsole(chromeHWND); err != nil {
		fmt.Fprintf(os.Stderr, "✗ Error clicking button: %v\n", err)
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

// navigateToPR navigates to the PR using clipboard + Ctrl+V
func navigateToPR(hwnd uintptr, url string) error {
	// Set Chrome window to foreground
	fmt.Println("  Setting Chrome window to foreground...")
	for i := 0; i < 3; i++ {
		procSetForeground.Call(hwnd)
		time.Sleep(300 * time.Millisecond)
	}

	fmt.Println("  Focusing address bar (Ctrl+L)...")
	time.Sleep(1 * time.Second)

	// Send Ctrl+L to focus address bar
	if err := sendKeyCombo(VK_CONTROL, VK_L); err != nil {
		return fmt.Errorf("failed to send Ctrl+L: %w", err)
	}

	time.Sleep(1 * time.Second)

	fmt.Println("  Copying URL to clipboard and pasting...")

	// Copy URL to clipboard
	if err := copyToClipboard(url); err != nil {
		return fmt.Errorf("failed to copy URL to clipboard: %w", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Send Ctrl+V to paste
	if err := sendKeyCombo(VK_CONTROL, VK_V); err != nil {
		return fmt.Errorf("failed to send Ctrl+V: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Println("  Pressing Enter to navigate...")

	// Send Enter to navigate
	if err := sendKeyPress(VK_RETURN); err != nil {
		return fmt.Errorf("failed to send Enter: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	return nil
}

// clickReviewButtonViaConsole opens DevTools and executes JavaScript
func clickReviewButtonViaConsole(hwnd uintptr) error {
	jsCode := "document.getElementById('re-request-review-copilot-pull-request-reviewer').click()"

	// Set Chrome to foreground
	procSetForeground.Call(hwnd)
	time.Sleep(300 * time.Millisecond)

	fmt.Println("  Opening DevTools (Ctrl+Shift+I)...")

	// Send Ctrl+Shift+I to open DevTools
	if err := sendKeyTriple(VK_CONTROL, VK_SHIFT, VK_I); err != nil {
		return fmt.Errorf("failed to open DevTools: %w", err)
	}

	time.Sleep(2 * time.Second) // Wait for DevTools to open

	fmt.Println("  DevTools should now be open with Console tab visible...")
	fmt.Println("  Copying JavaScript code to clipboard...")

	// Copy JavaScript to clipboard
	if err := copyToClipboard(jsCode); err != nil {
		return fmt.Errorf("failed to copy JavaScript to clipboard: %w", err)
	}

	time.Sleep(300 * time.Millisecond)

	fmt.Println("  Pasting JavaScript (Ctrl+V)...")

	// First paste attempt
	if err := sendKeyCombo(VK_CONTROL, VK_V); err != nil {
		return fmt.Errorf("failed to paste JavaScript: %w", err)
	}

	time.Sleep(1 * time.Second) // Wait for paste warning to appear

	fmt.Println("  Typing 'allow pasting' and pressing Enter...")

	// Type "allow pasting"
	for _, ch := range "allow pasting" {
		if err := typeCharacter(ch); err != nil {
			return fmt.Errorf("failed to type character: %w", err)
		}
		time.Sleep(50 * time.Millisecond)
	}

	time.Sleep(300 * time.Millisecond)

	// Press Enter
	if err := sendKeyPress(VK_RETURN); err != nil {
		return fmt.Errorf("failed to press Enter: %w", err)
	}

	time.Sleep(300 * time.Millisecond)

	fmt.Println("  Pasting JavaScript again (Ctrl+V)...")

	// Second paste attempt
	if err := sendKeyCombo(VK_CONTROL, VK_V); err != nil {
		return fmt.Errorf("failed to paste JavaScript again: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	fmt.Println("  Pressing Enter to execute JavaScript...")

	// Press Enter to execute
	if err := sendKeyPress(VK_RETURN); err != nil {
		return fmt.Errorf("failed to execute JavaScript: %w", err)
	}

	time.Sleep(1 * time.Second)

	fmt.Println("  Closing DevTools (Ctrl+Shift+I)...")

	// Close DevTools with Ctrl+Shift+I
	if err := sendKeyTriple(VK_CONTROL, VK_SHIFT, VK_I); err != nil {
		return fmt.Errorf("failed to close DevTools: %w", err)
	}

	time.Sleep(500 * time.Millisecond)

	return nil
}

// typeCharacter sends a single character via keyboard
func typeCharacter(ch rune) error {
	// For simplicity, use VK codes for common characters
	// This is a simplified approach - a full implementation would map all characters
	vkMap := map[rune]uint16{
		'a': 0x41, 'b': 0x42, 'c': 0x43, 'd': 0x44, 'e': 0x45,
		'f': 0x46, 'g': 0x47, 'h': 0x48, 'i': 0x49, 'j': 0x4A,
		'k': 0x4B, 'l': 0x4C, 'm': 0x4D, 'n': 0x4E, 'o': 0x4F,
		'p': 0x50, 'q': 0x51, 'r': 0x52, 's': 0x53, 't': 0x54,
		'u': 0x55, 'v': 0x56, 'w': 0x57, 'x': 0x58, 'y': 0x59,
		'z': 0x5A, ' ': 0x20,
	}

	vk, ok := vkMap[ch]
	if !ok {
		return fmt.Errorf("unsupported character: %c", ch)
	}

	return sendKeyPress(vk)
}

// copyToClipboard puts text on the Windows clipboard by writing to a temp file
func copyToClipboard(text string) error {
	tmpFile := filepath.Join(os.TempDir(), "copilot_clip.txt")

	if err := ioutil.WriteFile(tmpFile, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	// Use PowerShell to read file and set clipboard
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf("Get-Content '%s' | Set-Clipboard", tmpFile))

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set clipboard: %w", err)
	}

	// Clean up temp file
	_ = os.Remove(tmpFile)

	return nil
}

// sendKeyTriple sends three keys together (like Ctrl+Shift+I)
func sendKeyTriple(vkey1, vkey2, vkey3 uint16) error {
	if err := sendKeyDown(vkey1); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	if err := sendKeyDown(vkey2); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	if err := sendKeyDown(vkey3); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	if err := sendKeyUp(vkey3); err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	if err := sendKeyUp(vkey2); err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	if err := sendKeyUp(vkey1); err != nil {
		return err
	}

	return nil
}

// sendKeyCombo sends two keys together (like Ctrl+L)
func sendKeyCombo(vkey1, vkey2 uint16) error {
	if err := sendKeyDown(vkey1); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	if err := sendKeyDown(vkey2); err != nil {
		return err
	}

	time.Sleep(100 * time.Millisecond)

	if err := sendKeyUp(vkey2); err != nil {
		return err
	}

	time.Sleep(50 * time.Millisecond)

	if err := sendKeyUp(vkey1); err != nil {
		return err
	}

	return nil
}

// sendKeyPress sends a single key down and up
func sendKeyPress(vkey uint16) error {
	if err := sendKeyDown(vkey); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return sendKeyUp(vkey)
}

// sendKeyDown sends a key down event
func sendKeyDown(vkey uint16) error {
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
		return fmt.Errorf("SendInput failed: %w", err)
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
