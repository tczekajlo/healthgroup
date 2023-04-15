package main

import (
	"fmt"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/tczekajlo/healthgroup/internal/config"
	"github.com/tczekajlo/healthgroup/internal/log"
	"github.com/tczekajlo/healthgroup/internal/server"
	"github.com/tczekajlo/healthgroup/internal/version"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
	"k8s.io/client-go/util/homedir"
)

func main() {
	f := &config.Flags{}

	flag.BoolVar(&f.InCluster, "in-cluster", false, "use in-cluster config. Use always in a case when the app is running on a Kubernetes cluster")
	flag.StringVar(&f.ConfigFile, "config", "", "config file (default is $HOME/.healthgroup.yaml)")

	if home := homedir.HomeDir(); home != "" {
		flag.StringVar(&f.Kubeconfig, "kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to the kubeconfig file")
	} else {
		flag.StringVar(&f.Kubeconfig, "kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	viper.SetEnvPrefix("hg")
	viper.AutomaticEnv()

	if err := run(f); err != nil {
		fmt.Fprintf(os.Stderr, "an error occurred: %s\n", err)
		os.Exit(1)
	}
}

func run(flags *config.Flags) error {
	logger, err := log.NewAtLevel(viper.GetString("log_level"))
	if err != nil {
		return err
	}
	logger.Info("healthgroup", zap.String("version", version.Version))

	defer func() {
		err = logger.Sync()
	}()

	if _, err := maxprocs.Set(maxprocs.Logger(logger.Sugar().Debugf)); err != nil {
		return err
	}

	config := config.New(
		config.WithFile(flags.ConfigFile),
		config.WithLogger(logger),
		config.WithFlags(flags),
	)
	if err := config.ReadConfig(); err != nil {
		return err
	}

	return server.NewHTTP(config, logger)
}
