package migration

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/types"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type addModel struct {
	cliCtx      context.Context
	cfg         types.Config
	name        string
	description string

	errMsg error

	spinner spinner.Model
}

func NewAddModel(ctx context.Context, cfg types.Config, name, description string) *addModel {
	return &addModel{
		cliCtx:      ctx,
		name:        name,
		description: description,
		cfg:         cfg,

		spinner: spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("205"))),
		),
	}
}

func (m *addModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, addExec(m.name, m.description, m.cfg.Migration.Directory))
}

func (m *addModel) View() string {
	if m.errMsg != nil {
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			"❌ ",
			m.errMsg.Error(),
		)
	}

	return m.spinner.View() + " Hey hey!"
}

func (m *addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlQ:
			return m, tea.Quit
		default:
			return m, nil
		}

	case errorMsg:
		m.errMsg = msg
		// Create a custom quit message to display the error properly
		return m, tea.Quit

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func addExec(name, description, migrationsDir string) func() tea.Msg {
	return func() tea.Msg {
		time.Sleep(time.Second * 2)
		timestamp := time.Now().UTC().Unix()
		sanitizedName := nonAlphaNumericOrUnderscore.ReplaceAllString(
			strings.ReplaceAll(strings.ToLower(name), " ", "_"), "",
		)
		migrationDirName := fmt.Sprintf("%d_%s", timestamp, sanitizedName)

		md := types.Migration{
			Up:        filepath.Join(migrationsDir, fmt.Sprintf("%s/up.sql", migrationDirName)),
			Down:      filepath.Join(migrationsDir, fmt.Sprintf("%s/down.sql", migrationDirName)),
			Metadata:  filepath.Join(migrationsDir, fmt.Sprintf("%s/metadata.yaml", migrationDirName)),
			Timestamp: timestamp,
		}

		if err := saveMigration(md, sanitizedName, description, migrationDirName); err != nil {
			return errorMsg(err)
		}

		return nil
	}
}

// Add creates a new directory with stub migration files in the given schema and returns the
// names of the newly created files. If there was an error, the filesystem is rolled-back.
func Add(name string, cfg types.Config) error {
	timestamp := time.Now().UTC().Unix()
	sanitizedName := nonAlphaNumericOrUnderscore.ReplaceAllString(
		strings.ReplaceAll(strings.ToLower(name), " ", "_"), "",
	)
	migrationName := fmt.Sprintf("%d_%s", timestamp, sanitizedName)

	m := types.Migration{
		Up:        filepath.Join(cfg.Migration.Directory, fmt.Sprintf("%s/up.sql", migrationName)),
		Down:      filepath.Join(cfg.Migration.Directory, fmt.Sprintf("%s/down.sql", migrationName)),
		Metadata:  filepath.Join(cfg.Migration.Directory, fmt.Sprintf("%s/metadata.yaml", migrationName)),
		Timestamp: timestamp,
	}

	fmt.Printf("%sMigration created successfully!%s\n", output.StyleSuccess, output.StyleReset)
	if cfg.Verbose {
		fmt.Printf("%sUp query file: %s%s\n", output.StyleInfo, m.Up, output.StyleReset)
		fmt.Printf("%sDown query file: %s%s\n", output.StyleInfo, m.Down, output.StyleReset)
		fmt.Printf("%sMetadata file: %s%s\n", output.StyleInfo, m.Metadata, output.StyleReset)
	}

	return nil
}
