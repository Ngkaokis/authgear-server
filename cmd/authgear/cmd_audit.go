package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/authgear/authgear-server/pkg/util/sqlmigrate"
)

func init() {
	cmdAudit.AddCommand(cmdAuditDatabase)
	cmdAuditDatabase.AddCommand(cmdAuditDatabaseMigrate)

	cmdAuditDatabaseMigrate.AddCommand(cmdAuditDatabaseMigrateNew)
	cmdAuditDatabaseMigrate.AddCommand(cmdAuditDatabaseMigrateUp)
	cmdAuditDatabaseMigrate.AddCommand(cmdAuditDatabaseMigrateDown)
	cmdAuditDatabaseMigrate.AddCommand(cmdAuditDatabaseMigrateStatus)

	for _, cmd := range []*cobra.Command{cmdAuditDatabaseMigrateUp, cmdAuditDatabaseMigrateDown, cmdAuditDatabaseMigrateStatus} {
		ArgDatabaseURL.Bind(cmd.Flags(), viper.GetViper())
		ArgDatabaseSchema.Bind(cmd.Flags(), viper.GetViper())
	}
}

var AuditMigrationSet = sqlmigrate.NewMigrateSet("_audit_migration", "migrations/audit")

var cmdAudit = &cobra.Command{
	Use:   "audit database",
	Short: "Audit log commands",
}

var cmdAuditDatabase = &cobra.Command{
	Use:   "database migrate",
	Short: "Audit log database commands",
}

var cmdAuditDatabaseMigrate = &cobra.Command{
	Use:   "migrate [new|status|up|down]",
	Short: "Migrate database schema",
}

var cmdAuditDatabaseMigrateNew = &cobra.Command{
	Use:    "new",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		name := strings.Join(args, "_")
		_, err = AuditMigrationSet.Create(name)
		if err != nil {
			return
		}

		return
	},
}

var cmdAuditDatabaseMigrateUp = &cobra.Command{
	Use:   "up",
	Short: "Migrate database schema to latest version",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			return
		}
		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
		if err != nil {
			return
		}

		_, err = AuditMigrationSet.Up(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, 0)
		if err != nil {
			return
		}

		return
	},
}

var cmdAuditDatabaseMigrateDown = &cobra.Command{
	Use:    "down",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			return
		}
		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
		if err != nil {
			return
		}

		if len(args) == 0 {
			err = fmt.Errorf("number of migrations to revert not specified; specify 'all' to revert all migrations")
			return
		}

		var numMigrations int
		if args[0] == "all" {
			numMigrations = 0
		} else {
			numMigrations, err = strconv.Atoi(args[0])
			if err != nil {
				err = fmt.Errorf("invalid number of migrations specified: %s", err)
				return
			} else if numMigrations <= 0 {
				err = fmt.Errorf("no migrations specified to revert")
				return
			}
		}

		_, err = AuditMigrationSet.Down(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		}, numMigrations)
		if err != nil {
			return
		}

		return
	},
}

var cmdAuditDatabaseMigrateStatus = &cobra.Command{
	Use:   "status",
	Short: "Get database schema migration status",
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		dbURL, err := ArgDatabaseURL.GetRequired(viper.GetViper())
		if err != nil {
			return
		}
		dbSchema, err := ArgDatabaseSchema.GetRequired(viper.GetViper())
		if err != nil {
			return
		}

		plans, err := AuditMigrationSet.Status(sqlmigrate.ConnectionOptions{
			DatabaseURL:    dbURL,
			DatabaseSchema: dbSchema,
		})
		if err != nil {
			return
		}

		if len(plans) != 0 {
			os.Exit(1)
		}

		return
	},
}
