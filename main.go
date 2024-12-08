package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/charmbracelet/log"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/golang-migrate/migrate/source/file"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var httpMetricsServer *http.Server

var (
	DodudaVersion     = "v0.1.0"
	DodualmShort      = "dodualm - Dofus Almanax API"
	DodualmLong       = "The Dofus Almanax API."
	DodudaVersionHelp = DodualmShort + "\n" + DodudaVersion + "\nhttps://github.com/dofusdude/dodualm"

	ApiPort     string
	ServerTz    string
	ApiScheme   string
	ApiHostName string
	MeiliHost   string
	MeiliKey    string

	rootCmd = &cobra.Command{
		Use:           "dodualm",
		Short:         DodualmShort,
		Long:          DodualmLong,
		SilenceErrors: true,
		SilenceUsage:  false,
		Run:           rootCommand,
	}

	migrateCmd = &cobra.Command{
		Use:           "migrate",
		Short:         "Run migrations on the database.",
		SilenceErrors: true,
		SilenceUsage:  false,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Running migrate command")
		},
	}

	migrateDownCmd = &cobra.Command{
		Use:   "down",
		Short: "migrate from v2 to v1",
		Long:  `Command to downgrade database from v2 to v1`,
		Run:   migrateDown,
	}

	migrateUpCmd = &cobra.Command{
		Use:   "down",
		Short: "migrate from v2 to v1",
		Long:  `Command to downgrade database from v2 to v1`,
		Run:   migrateUp,
	}
)

func migrateUp(cmd *cobra.Command, args []string) {
	database := NewDatabaseRepository(context.Background(), ".")

	dbDriver, err := sqlite3.WithInstance(database.Db, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("instance error: %v \n", err)
	}

	fileSource, err := (&file.File{}).Open("file://migrations")
	if err != nil {
		log.Fatalf("opening file error: %v \n", err)
	}

	m, err := migrate.NewWithInstance("file", fileSource, "myDB", dbDriver)
	if err != nil {
		log.Fatalf("migrate error: %v \n", err)
	}

	if err = m.Up(); err != nil {
		log.Fatalf("migrate up error: %v \n", err)
	}

	log.Print("Migrate up done with success")
}

func migrateDown(cmd *cobra.Command, args []string) {
	database := NewDatabaseRepository(context.Background(), ".")

	dbDriver, err := sqlite3.WithInstance(database.Db, &sqlite3.Config{})
	if err != nil {
		log.Fatalf("instance error: %v \n", err)
	}

	fileSource, err := (&file.File{}).Open("file://migrations")
	if err != nil {
		log.Fatalf("opening file error: %v \n", err)
	}

	m, err := migrate.NewWithInstance("file", fileSource, "myDB", dbDriver)
	if err != nil {
		log.Fatalf("migrate error: %v \n", err)
	}

	if err = m.Down(); err != nil {
		log.Fatalf("migrate down error: %v \n", err)
	}

	log.Print("Migrate down done with success")
}

func rootCommand(cmd *cobra.Command, args []string) {
	if version, _ := cmd.Flags().GetBool("version"); version {
		fmt.Println(DodudaVersion)
		return
	}

	metrics, err := cmd.Flags().GetBool("metrics")
	if err != nil {
		log.Fatal(err)
	}

	httpDataServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", ApiPort),
		Handler: Router(),
	}

	if metrics {
		apiPort, _ := strconv.Atoi(ApiPort)
		metricsPort := apiPort + 1
		httpMetricsServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", metricsPort),
			Handler: promhttp.Handler(),
		}

		go func() {
			log.Info("Metrics server started", "port", metricsPort)
			if err := httpMetricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
	}

	log.Info("Almanax server started", "port", ApiPort)
	if err := httpDataServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func main() {
	viper.SetDefault("LOG_LEVEL", "info")
	viper.AutomaticEnv()

	rootCmd.Flags().Bool("version", false, "Print the dodualm version.")
	rootCmd.Flags().Bool("metrics", false, "Toggle Prometheus metrics export.")

	rootCmd.AddCommand(migrateCmd)
	migrateCmd.AddCommand(migrateDownCmd)

	viper.SetDefault("MEILI_PORT", "7700")
	viper.SetDefault("MEILI_MASTER_KEY", "masterKey")
	viper.SetDefault("MEILI_PROTOCOL", "http")
	viper.SetDefault("MEILI_HOST", "127.0.0.1")
	viper.SetDefault("API_PORT", "3000")
	viper.SetDefault("API_SCHEME", "http")
	viper.SetDefault("API_HOSTNAME", "localhost")
	viper.SetDefault("SERVER_TZ", "Europe/Berlin")

	ApiScheme = viper.GetString("API_SCHEME")
	ApiHostName = viper.GetString("API_HOSTNAME")
	ApiPort = viper.GetString("API_PORT")
	MeiliKey = viper.GetString("MEILI_MASTER_KEY")
	MeiliHost = fmt.Sprintf("%s://%s:%s", viper.GetString("MEILI_PROTOCOL"), viper.GetString("MEILI_HOST"), viper.GetString("MEILI_PORT"))
	ServerTz = getEnv("SERVER_TZ", "Europe/Berlin")

	err := rootCmd.Execute()
	if err != nil && err.Error() != "" {
		fmt.Fprintln(os.Stderr, err)
	}
}
