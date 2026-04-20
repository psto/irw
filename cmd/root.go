package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/psto/irw/internal/commands"
	"github.com/psto/irw/internal/config"
	"github.com/psto/irw/internal/db"
	"github.com/spf13/cobra"
)

var (
	dbPath      string
	database    *sql.DB
	compactMode bool
	rawMode     bool
	nullMode    bool
	cfg         *config.Config
	trackQueue  string
)

var rootCmd = &cobra.Command{
	Use:   "irw",
	Short: "Spaced repetition file tracker",
	Long:  `A spaced repetition system for managing reading and writing queues.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return err
		}
		if dbPath == "" && !cfg.DBPathExplicit {
			fmt.Fprintf(os.Stderr, "Warning: no db_path set. Set in ~/.config/irw/config.json or use --db flag. Default: %s\n", config.DefaultDBPath())
		}
		database, err = db.Connect(cfg, dbPath)
		if err != nil {
			return err
		}
		return db.CreateTables(database)
	},
}

var trackCmd = &cobra.Command{
	Use:   "track <file>",
	Short: "Add a new file to the tracker",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.Track(database, args[0], trackQueue)
	},
}

var untrackCmd = &cobra.Command{
	Use:   "untrack [file]",
	Short: "Remove a file from the tracker",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := ""
		if len(args) > 0 {
			target = args[0]
		}
		return commands.Untrack(database, target)
	},
}

var completeCmd = &cobra.Command{
	Use:   "complete [file]",
	Short: "Mark a file as finished",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := ""
		if len(args) > 0 {
			target = args[0]
		}
		return commands.Complete(database, target)
	},
}

var priorityCmd = &cobra.Command{
	Use:   "priority [file] [p]",
	Short: "Set priority (0-100) for a file",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		target := ""
		priority := 0
		if len(args) > 0 {
			target = args[0]
		}
		if len(args) > 1 {
			fmt.Sscanf(args[1], "%d", &priority)
		}
		return commands.SetPriority(database, target, priority)
	},
}

var statsCmd = &cobra.Command{
	Use:   "stats [type]",
	Short: "Show completion progress",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		trackType := ""
		if len(args) > 0 {
			trackType = args[0]
		}
		commands.ShowStats(database, trackType)
	},
}

var scheduleCmd = &cobra.Command{
	Use:   "schedule",
	Short: "Show upcoming due dates and priorities",
	Run: func(cmd *cobra.Command, args []string) {
		commands.ShowSchedule(database, rawMode, nullMode)
	},
}

var reviewCmd = &cobra.Command{
	Use:   "review [type] [ext]",
	Short: "Start a review session",
	Args:  cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		trackType := ""
		ext := ""
		if len(args) > 0 {
			trackType = args[0]
		}
		if len(args) > 1 {
			ext = args[1]
		}
		return commands.Review(cfg, database, trackType, ext, compactMode)
	},
}

var readCmd = &cobra.Command{
	Use:   "read",
	Short: "Pick a specific file to read",
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.ReadFile(cfg, database)
	},
}

var importCmd = &cobra.Command{
	Use: "import",
	Short: "Sync files from zk and sioyek",
	Run: func(cmd *cobra.Command, args []string) {
		commands.Import(database, cfg)
	},
}

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Delete finished files from database",
	RunE: func(cmd *cobra.Command, args []string) error {
		return commands.PurgeFinished(database)
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", "", "database path (overrides config)")
	reviewCmd.Flags().BoolVar(&compactMode, "compact", false, "use compact terminal output")
	scheduleCmd.Flags().BoolVar(&rawMode, "raw", false, "output raw CSV format")
	scheduleCmd.Flags().BoolVarP(&nullMode, "print0", "0", false, "output null-delimited paths (for xargs -0)")
	trackCmd.Flags().StringVarP(&trackQueue, "queue", "q", "reading", "queue to add file to (reading/writing)")

	rootCmd.AddCommand(trackCmd)
	rootCmd.AddCommand(untrackCmd)
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(priorityCmd)
	rootCmd.AddCommand(statsCmd)
	rootCmd.AddCommand(scheduleCmd)
	rootCmd.AddCommand(reviewCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(importCmd)
	rootCmd.AddCommand(purgeCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
