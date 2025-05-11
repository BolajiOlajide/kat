package migration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BolajiOlajide/kat/internal/output"
	"github.com/BolajiOlajide/kat/internal/types"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/urfave/cli/v2"
)

type fileContentMsg struct {
	content string
	err     error
}

type addModel struct {
	cliCtx         *cli.Context
	name           string
	description    string
	progressWriter *output.ProgressWriter

	result   addResult
	progress progress.Model
	spinner  spinner.Model
	pending  bool
	done     bool
	err      error
}

type addResult struct {
	cfg       types.Config
	timestamp int64
	name      string
	parents   []int64
}

func NewAddModel(c *cli.Context, name string) *addModel {
	description := c.String("description")

	s := spinner.New(
		spinner.WithStyle(
			lipgloss.NewStyle().Foreground(lipgloss.Color("63")),
		),
		spinner.WithSpinner(spinner.Dot),
	)

	return &addModel{
		cliCtx:      c,
		name:        name,
		description: description,
		progress:    progress.New(progress.WithDefaultGradient()),
		spinner:     s,
		// progressWriter: progressWriter,
	}
}

func (m *addModel) SetProgressWriter(progressWriter *output.ProgressWriter) {
	m.progressWriter = progressWriter
}

func (m *addModel) Init() tea.Cmd {
	// let's start the process
	m.pending = true
	return m.spinner.Tick
}

func (m *addModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.pending = false
		// Use the common quit handler
		return handleQuitKeyPress(msg, m)

	case tea.WindowSizeMsg:
		m.progress.Width = min(msg.Width-padding*2-4, maxWidth)
		return m, nil

	case output.ProgressError:
		m.err = msg.Err
		return m, tea.Quit

	case output.ProgresCount:
		cmd := m.progress.SetPercent(float64(msg))
		return m, cmd

	// FrameMsg is sent when the progress bar wants to animate itself
	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	case errMsg:
		m.err = msg
		return m, tea.Quit

	case fileContentMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, tea.Quit
		}
		// Process the file content if needed
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	default:
		return m, readConfigFile
	}
}

func readConfigFile() tea.Msg {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fileContentMsg{err: err}
	}

	// .config/gh/config.yml
	configPath := filepath.Join(homeDir, ".config", "gh", "config.yml")
	content, err := os.ReadFile(configPath)
	if err != nil {
		return fileContentMsg{err: err}
	}

	return fileContentMsg{content: string(content)}
}

func (m *addModel) View() string {
	if m.err != nil {
		return lipgloss.JoinVertical(0.2, fmt.Sprintf("❌ %s", errorStyle(m.err.Error())))
	}

	if m.pending {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.JoinHorizontal(
				lipgloss.Left,
				m.spinner.View(),
				" ",
				"Getting Kat configuration",
				"\n",
			),
			m.progress.View(),
			infoStyle("Press 'ctrl+c' or 'ctrl+q' to quit\n"),
		)
	}

	var out []string
	out = append(out, successStyle("Migration successfully created."))
	if m.result.cfg.Verbose {
		out = append(
			out, infoStyle(
				fmt.Sprintf(
					"Directory: %s/%d_%s", m.result.cfg.Migration.Directory, m.result.timestamp, m.result.name,
				),
			),
		)

		if len(m.result.parents) > 0 {
			out = append(out, infoStyle(fmt.Sprintf("Parents: %v", m.result.parents)))
		}
	}

	fmt.Println("hellloooo", out)
	return lipgloss.JoinVertical(padding, out...)
}

// Add creates a new directory with stub migration files in the given schema and returns the
// names of the newly created files. If there was an error, the filesystem is rolled-back.
// func (a *addModel) execute() tea.Cmd {
// 	cfg, err := config.GetKatConfigFromCtx(a.cliCtx)
// 	if err != nil {
// 		return err
// 	}

// 	timestamp := time.Now().UTC().Unix()
// 	sanitizedName := nonAlphaNumericOrUnderscore.ReplaceAllString(
// 		strings.ReplaceAll(strings.ToLower(a.name), " ", "_"), "_",
// 	)
// 	migrationDirName := fmt.Sprintf("%d_%s", timestamp, sanitizedName)
// 	m := types.Migration{
// 		Up:        filepath.Join(cfg.Migration.Directory, fmt.Sprintf("%s/up.sql", migrationDirName)),
// 		Down:      filepath.Join(cfg.Migration.Directory, fmt.Sprintf("%s/down.sql", migrationDirName)),
// 		Metadata:  filepath.Join(cfg.Migration.Directory, fmt.Sprintf("%s/metadata.yaml", migrationDirName)),
// 		Timestamp: timestamp,
// 	}

// 	// Get the current parent migrations from the database
// 	parents, err := getCurrentParentMigrations(a.cliCtx.Context, cfg)
// 	if err != nil {
// 		return errors.Wrap(err, "getting current parent migrations")
// 	}

// 	metadata := types.MigrationMetadata{
// 		Name:        sanitizedName,
// 		Timestamp:   timestamp,
// 		Description: a.description,
// 		Parents:     parents,
// 	}
// 	if err := saveMigration(m, metadata); err != nil {
// 		return errors.Wrap(err, "saving migration metadata")
// 	}

// 	return nil
// }

// 	fmt.Printf("%sMigration created successfully!%s\n", output.StyleSuccess, output.StyleReset)
// 	if cfg.Verbose {
// 		fmt.Printf("%sUp query file: %s%s\n", output.StyleInfo, m.Up, output.StyleReset)
// 		fmt.Printf("%sDown query file: %s%s\n", output.StyleInfo, m.Down, output.StyleReset)
// 		fmt.Printf("%sMetadata file: %s%s\n", output.StyleInfo, m.Metadata, output.StyleReset)

// 		if len(parents) > 0 {
// 			fmt.Printf("%sParent migrations:%s\n", output.StyleInfo, output.StyleReset)
// 			fmt.Printf("%s  - %v%s\n", output.StyleInfo, parents, output.StyleReset)
// 		} else {
// 			fmt.Printf("%sNo parent migrations (first migration)%s\n", output.StyleInfo, output.StyleReset)
// 		}
// 	}

// 	return nil
// }

// getCurrentParentMigrations retrieves the list of already applied migrations to use as parents
// func getCurrentParentMigrations(ctx context.Context, cfg types.Config) ([]int64, error) {
// 	// Connect to the database
// 	dbConn, err := cfg.Database.ConnString()
// 	if err != nil {
// 		return nil, errors.Wrap(err, "getting database connection string")
// 	}

// 	db, err := database.New(dbConn)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "connecting to database")
// 	}
// 	defer db.Close()

// 	// Check if migration table exists
// 	query := sqlf.Sprintf(
// 		"SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = %s)",
// 		cfg.Migration.TableName,
// 	)

// 	var tableExists bool
// 	row := db.QueryRow(ctx, query)
// 	if err := row.Scan(&tableExists); err != nil {
// 		// If we can't determine if the table exists, assume no parents
// 		return nil, nil
// 	}

// 	if !tableExists {
// 		// If migration table doesn't exist yet, this is the first migration
// 		return nil, nil
// 	}

// 	// Query the existing migrations - extract timestamps from names
// 	selectQuery := sqlf.Sprintf(
// 		"SELECT name FROM %s ORDER BY migration_time DESC",
// 		sqlf.Sprintf(cfg.Migration.TableName),
// 	)

// 	rows, err := db.Query(ctx, selectQuery)
// 	if err != nil {
// 		return nil, errors.Wrap(err, "querying migrations")
// 	}
// 	defer rows.Close()

// 	// Collect the timestamps from migration names
// 	// Migration name format is "TIMESTAMP_name"
// 	var parents []int64
// 	for rows.Next() {
// 		var name string
// 		if err := rows.Scan(&name); err != nil {
// 			return nil, errors.Wrap(err, "scanning migration name")
// 		}

// 		// Extract timestamp from the migration name
// 		timestampStr := strings.Split(name, "_")[0]
// 		timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
// 		if err != nil {
// 			return nil, errors.Wrap(err, "parsing timestamp from migration name")
// 		}
// 		parents = append(parents, timestamp)
// 	}

// 	return parents, nil
// }
