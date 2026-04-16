package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/psto/irw/internal/db"
	"golang.org/x/term"
)

func getTerminalWidth() int {
	if width, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && width > 0 {
		return width
	}
	return 120
}

func RenderStats(trackType string, active, finished, due int, completion float64) string {
	var b strings.Builder

	header := fmt.Sprintf("## 📊 %s queue", strings.Title(trackType))
	rendered, _ := glamour.Render(header, "dark")
	b.WriteString(rendered)

	t := table.New().
		Width(getTerminalWidth()).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("63"))).
		Headers("Active", "Finished", "Due", "Completion").
		Row(
			fmt.Sprintf("%d", active),
			fmt.Sprintf("%d", finished),
			fmt.Sprintf("%d", due),
			fmt.Sprintf("%.1f%%", completion),
		)

	b.WriteString(t.String())
	b.WriteString("\n")
	return b.String()
}

func RenderSessions(sessions []db.SessionStats) string {
	var b strings.Builder

	header := "## 📅 Daily Review Activity (All)"
	rendered, _ := glamour.Render(header, "dark")
	b.WriteString(rendered)

	t := table.New().
		Width(getTerminalWidth()).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("63"))).
		Headers("Date", "Time", "Items", "Fin", "Avg/Item")

	for _, s := range sessions {
		hours := s.Duration / 3600
		mins := (s.Duration % 3600) / 60
		secs := s.Duration % 60
		timeStr := fmt.Sprintf("%02d:%02d:%02d", hours, mins, secs)
		avgStr := fmt.Sprintf("%.2fm", s.AvgPer)

		t.Row(s.Date, timeStr, fmt.Sprintf("%d", s.Reviewed), fmt.Sprintf("%d", s.Finished), avgStr)
	}

	b.WriteString(t.String())
	b.WriteString("\n")
	return b.String()
}

func RenderReviewItem(filename string, remaining int) string {
	md := fmt.Sprintf(`📖 **%s**

* **[Enter]** Next Interval
* **[p]** Set Priority
* **[f]** Finish File
* **[s]** Skip
* **[z]** Postpone
* **[q]** Quit

📊 %d due
`, filename, remaining)

	rendered, _ := glamour.Render(md, "dark")
	return rendered
}

func RenderReviewItemCompact(filename string, remaining int) string {
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Bold(true)

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("8"))

	header := style.Render("📖 "+filename) + " " + dimStyle.Render(fmt.Sprintf("(%d due)", remaining))
	hints := dimStyle.Render("[Enter]next [p]riority [f]inish [s]kip [z]postpone [q]uit")

	return header + "\n" + hints
}

func RenderQueueEmpty() string {
	rendered, _ := glamour.Render("# 🎉 Queue Empty!", "dark")
	return rendered
}

func RenderSchedule(items []db.ScheduleItem) string {
	width := getTerminalWidth()

	t := table.New().
		Width(width).
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("63"))).
		Headers("Due", "Int", "AF", "Priority", "Type", "File")

	for _, item := range items {
		filename := filepath.Base(item.Path)
		t.Row(
			item.DueDate,
			fmt.Sprintf("%.1f", item.Interval),
			fmt.Sprintf("%.1f", item.Afactor),
			fmt.Sprintf("%d", item.Priority),
			item.Type,
			filename,
		)
	}

	return t.String()
}
