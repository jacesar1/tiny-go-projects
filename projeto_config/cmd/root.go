package cmd

import (
	"fmt"
	"os"
	"projeto_config/internal/gcp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// Version da aplicação (formato semântico: MAJOR.MINOR.PATCH)
	Version = "1.0.0"

	defaultOrgID          = "727440331682"
	defaultParentFolderID = "fldr-scge"
	defaultBillingAccount = "01F7C9-60D131-20DC44"
)

var rootCmd = &cobra.Command{
	Use:     "projeto_config",
	Short:   "Automacao de projetos GCP da Axia Energia",
	Version: Version,
	Long: `CLI para criar e atualizar projetos GCP seguindo o padrao da Axia Energia.

Exemplos:
  projeto_config create projeto benner-cloud
	projeto_config create projeto benner-cloud --all
	projeto_config create projeto benner-cloud --all --env qld --env prd
	projeto_config create projeto benner-cloud --all --interactive-envs
  projeto_config update projeto benner-cloud --labels
	projeto_config update projeto benner-cloud --networks
	projeto_config update projeto benner-cloud --service-accounts
	projeto_config update projeto benner-cloud --all --env qld
	projeto_config update projeto benner-cloud --all --interactive-envs
	projeto_config update projeto benner-cloud --apis --optional-api secretmanager
	projeto_config update projeto benner-cloud --apis --optional-api secretmanager,firestore
	projeto_config get projeto benner-cloud
	projeto_config describe projeto benner-cloud
	projeto_config delete projeto benner-cloud --yes

Ajuda detalhada por recurso:
  projeto_config create projeto -h
  projeto_config update projeto -h`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("org-id", defaultOrgID, "ID da organizacao GCP")
	rootCmd.PersistentFlags().String("parent-folder", defaultParentFolderID, "ID/nome da pasta pai (usado no create)")
	rootCmd.PersistentFlags().String("billing-account", defaultBillingAccount, "Billing account vinculada no create")
	rootCmd.PersistentFlags().Bool("show-gcloud-commands", true, "Exibe resumo dos comandos gcloud ao final de cada passo")

	_ = viper.BindPFlag("org-id", rootCmd.PersistentFlags().Lookup("org-id"))
	_ = viper.BindPFlag("parent-folder", rootCmd.PersistentFlags().Lookup("parent-folder"))
	_ = viper.BindPFlag("billing-account", rootCmd.PersistentFlags().Lookup("billing-account"))
	_ = viper.BindPFlag("show-gcloud-commands", rootCmd.PersistentFlags().Lookup("show-gcloud-commands"))

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		gcp.SetGCloudCommandSummaryEnabled(viper.GetBool("show-gcloud-commands"))
	}

	rootCmd.AddCommand(newCreateCommand())
	rootCmd.AddCommand(newUpdateCommand())
	rootCmd.AddCommand(newGetCommand())
	rootCmd.AddCommand(newDescribeCommand())
	rootCmd.AddCommand(newDeleteCommand())
}

func initConfig() {
	viper.SetEnvPrefix("PROJETO_CONFIG")
	viper.AutomaticEnv()

	viper.SetDefault("org-id", defaultOrgID)
	viper.SetDefault("parent-folder", defaultParentFolderID)
	viper.SetDefault("billing-account", defaultBillingAccount)
	viper.SetDefault("show-gcloud-commands", true)
}
