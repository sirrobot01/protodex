package style

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	SuccessColor = lipgloss.Color("#04B575")
	ErrorColor   = lipgloss.Color("#FF5F56")
	WarningColor = lipgloss.Color("#FFBD2E")
	InfoColor    = lipgloss.Color("#61AFEF")
	PrimaryColor = lipgloss.Color("#7B68EE")

	// Styles
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(WarningColor).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(InfoColor).
			Bold(true)

	PrimaryStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	TitleStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true).
			Underline(true)

	SubtleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666"))

	BoldStyle = lipgloss.NewStyle().
			Bold(true)

	// Box styles for important messages
	BoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1).
			Margin(1)
)

// Helper functions for common patterns
func Success(text string) string {
	return SuccessStyle.Render("" + text)
}

func Error(text string) string {
	return ErrorStyle.Render("" + text)
}

func Warning(text string) string {
	return WarningStyle.Render("  " + text)
}

func Info(text string) string {
	return InfoStyle.Render("ℹ  " + text)
}

func Progress(text string) string {
	return InfoStyle.Render(" " + text)
}

func Download(text string) string {
	return InfoStyle.Render(" " + text)
}

func Upload(text string) string {
	return InfoStyle.Render(" " + text)
}

func Generate(text string) string {
	return PrimaryStyle.Render(" " + text)
}

func Validate(text string) string {
	return InfoStyle.Render(" " + text)
}

func Title(text string) string {
	return TitleStyle.Render(text)
}

func Subtle(text string) string {
	return SubtleStyle.Render(text)
}

func Bold(text string) string {
	return BoldStyle.Render(text)
}

func Box(content string) string {
	return BoxStyle.Render(content)
}

// Spinner for loading states
type Spinner struct {
	frames []string
	delay  time.Duration
}

func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		delay:  100 * time.Millisecond,
	}
}

func (s *Spinner) Start(message string) chan bool {
	done := make(chan bool)
	go func() {
		i := 0
		for {
			select {
			case <-done:
				fmt.Printf("\r%s\n", strings.Repeat(" ", len(message)+2))
				return
			default:
				frame := s.frames[i%len(s.frames)]
				fmt.Printf("\r%s %s", InfoStyle.Render(frame), message)
				time.Sleep(s.delay)
				i++
			}
		}
	}()
	return done
}

// Progress bar for file operations
func ProgressBar(current, total int, width int) string {
	if total == 0 {
		return ""
	}

	percentage := float64(current) / float64(total)
	filled := int(percentage * float64(width))

	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)

	return fmt.Sprintf("%s [%s] %d/%d (%.1f%%)",
		InfoStyle.Render(""),
		bar,
		current,
		total,
		percentage*100,
	)
}

// File count display
func FileCount(count int, noun string) string {
	if count == 1 {
		return fmt.Sprintf("%s %d %s", InfoStyle.Render(""), count, noun)
	}
	return fmt.Sprintf("%s %d %ss", InfoStyle.Render(""), count, noun)
}

// Version display
func Version(pkg, version string) string {
	return fmt.Sprintf("%s %s",
		BoldStyle.Render(pkg),
		PrimaryStyle.Render(version))
}
