package cmd

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/Method-Security/pkg/signal"
	"github.com/Method-Security/pkg/writer"
	"github.com/method-security/methodk8s/internal/config"
	"github.com/palantir/pkg/datetime"
	"github.com/palantir/witchcraft-go-logging/wlog/svclog/svc1log"
	"github.com/spf13/cobra"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type MethodK8s struct {
	version      string
	RootFlags    config.RootFlags
	OutputConfig writer.OutputConfig
	OutputSignal signal.Signal
	K8Config     *rest.Config
	RootCmd      *cobra.Command
}

func NewMethodK8s(version string) *MethodK8s {
	methodK8s := MethodK8s{
		version: version,
		RootFlags: config.RootFlags{
			Quiet:   false,
			Verbose: false,
			ServiceAccountConfig: config.ServiceAccountConfig{
				ServiceAccount: false,
				Token:          "",
				CACert:         "",
			},
			KubeConfig: config.KubeConfig{
				Context: "",
				Path:    "",
				URL:     "",
			},
		},
		OutputConfig: writer.NewOutputConfig(nil, writer.NewFormat(writer.SIGNAL)),
		OutputSignal: signal.NewSignal(nil, datetime.DateTime(time.Now()), nil, 0, nil),
		K8Config:     nil,
	}
	return &methodK8s
}

func (a *MethodK8s) InitRootCommand() {
	var outputFormat string
	var outputFile string
	a.RootCmd = &cobra.Command{
		Use:   "methodk8s",
		Short: "Audit K8 resources",
		Long: `The K8s config is defined in order of:
		1. The '--serviceaccount' flag which creates a config via a token, CA-Cert, and URL set as ENV vars or flags
		2. The '--path' flag which passes the path to a .kube/config file
		3. The $KUBECONFIG var which holds the path to a .kube/config file
		4. The '--url' flag which passes the URL of a potential cluster`,
		SilenceUsage: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			var err error

			format, err := validateOutputFormat(outputFormat)
			if err != nil {
				return err
			}
			var outputFilePointer *string
			if outputFile != "" {
				outputFilePointer = &outputFile
			} else {
				outputFilePointer = nil
			}
			a.OutputConfig = writer.NewOutputConfig(outputFilePointer, format)
			cmd.SetContext(svc1log.WithLogger(cmd.Context(), config.InitializeLogging(cmd, &a.RootFlags)))

			k8Config, err := GetK8Config(a)
			if err != nil {
				return err
			}
			a.K8Config = k8Config

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
			completedAt := datetime.DateTime(time.Now())
			a.OutputSignal.CompletedAt = &completedAt
			return writer.Write(
				a.OutputSignal.Content,
				a.OutputConfig,
				a.OutputSignal.StartedAt,
				a.OutputSignal.CompletedAt,
				a.OutputSignal.Status,
				a.OutputSignal.ErrorMessage,
			)
		},
	}

	// Standard flags
	a.RootCmd.PersistentFlags().BoolVarP(&a.RootFlags.Quiet, "quiet", "q", false, "Suppress output")
	a.RootCmd.PersistentFlags().BoolVarP(&a.RootFlags.Verbose, "verbose", "v", false, "Verbose output")
	a.RootCmd.PersistentFlags().StringVarP(&outputFile, "output-file", "f", "", "Path to output file. If blank, will output to STDOUT")
	a.RootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "signal", "Output format (signal, json, yaml). Default value is signal")

	// ServiceAccountConfig flags
	a.RootCmd.PersistentFlags().BoolVarP(&a.RootFlags.ServiceAccountConfig.ServiceAccount, "serviceaccount", "s", false, "Set to true if using service account workflow")
	a.RootCmd.PersistentFlags().StringVarP(&a.RootFlags.ServiceAccountConfig.Token, "token", "t", "", "Base64 Service account token")
	a.RootCmd.PersistentFlags().StringVarP(&a.RootFlags.ServiceAccountConfig.CACert, "cert", "a", "", "Base64 encoded ca certificate")

	// KubeConfig flags
	a.RootCmd.PersistentFlags().StringVarP(&a.RootFlags.KubeConfig.Context, "context", "c", "", "Cluster context (ie. minikube)")
	a.RootCmd.PersistentFlags().StringVarP(&a.RootFlags.KubeConfig.Path, "path", "p", "", "Absolute or relative path to the config file (ie. ~/.kube/config)")
	a.RootCmd.PersistentFlags().StringVarP(&a.RootFlags.KubeConfig.URL, "url", "u", "", "Cluster url (ie. mycluster.com)")

	// Flag rules
	a.RootCmd.MarkFlagsMutuallyExclusive("context", "serviceaccount")
	a.RootCmd.MarkFlagsMutuallyExclusive("context", "token")
	a.RootCmd.MarkFlagsMutuallyExclusive("context", "cert")
	a.RootCmd.MarkFlagsMutuallyExclusive("path", "serviceaccount")
	a.RootCmd.MarkFlagsMutuallyExclusive("path", "token")
	a.RootCmd.MarkFlagsMutuallyExclusive("path", "cert")

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version number of methodk8s",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Println(a.version)
		},
		PersistentPostRunE: func(cmd *cobra.Command, _ []string) error {
			return nil
		},
	}

	a.RootCmd.AddCommand(versionCmd)
}

// A utility function to validate that the provided output format is one of the supported formats: json, yaml, signal.
func validateOutputFormat(output string) (writer.Format, error) {
	var format writer.FormatValue
	switch strings.ToLower(output) {
	case "json":
		format = writer.JSON
	case "yaml":
		return writer.Format{}, errors.New("yaml output format is not supported for methodk8s")
	case "signal":
		format = writer.SIGNAL
	default:
		return writer.Format{}, errors.New("invalid output format. Valid formats are: json, yaml, signal")
	}
	return writer.NewFormat(format), nil
}

// GetK8Config gets the k8s config object from the various auth mechanisms
func GetK8Config(a *MethodK8s) (*rest.Config, error) {
	if a.RootFlags.ServiceAccountConfig.ServiceAccount {
		k8Config, err := CreateConfigFromServiceAccountCreds(a.RootFlags.ServiceAccountConfig.Token, a.RootFlags.ServiceAccountConfig.CACert, a.RootFlags.KubeConfig.URL)
		if err != nil {
			return nil, err
		}
		return k8Config, nil
	} else if a.RootFlags.KubeConfig.Path != "" {
		k8Config, err := CreateConfigFromPath(a.RootFlags.KubeConfig.Path, a.RootFlags.KubeConfig.Context)
		if err != nil {
			return nil, err
		}
		return k8Config, nil

	} else if kubeEnv, exists := os.LookupEnv("KUBECONFIG"); exists && kubeEnv != "" {
		k8Config, err := CreateConfigFromPath(os.Getenv("KUBECONFIG"), a.RootFlags.KubeConfig.Context)
		if err != nil {
			return nil, err
		}
		return k8Config, nil

	} else if a.RootFlags.KubeConfig.URL != "" {
		k8ConfigURL := a.RootFlags.KubeConfig.URL
		k8Config := CreateConfigFromURL(k8ConfigURL)
		return k8Config, nil

	}
	err := errors.New("please provide either: " +
		"Service account creds," +
		"a path to a config file, " +
		"assign $KUBECONFIG to a path to the config file, " +
		"or provide a URL to the cluster")
	return nil, err
}

// CreateConfigFromServiceAccountCreds generates the k8s config object from a service account token, optional(ca cert), and cluster URL
func CreateConfigFromServiceAccountCreds(tokenFlag string, caCertFlag string, urlFlag string) (*rest.Config, error) {
	var err error

	clusterURL := urlFlag
	if clusterURL == "" {
		clusterURL = os.Getenv("K8S_CLUSTER_URL")
	}

	var token []byte
	if tokenFlag != "" {
		token, err = base64.StdEncoding.DecodeString(tokenFlag)
		if err != nil {
			return nil, err
		}
	} else {
		token, err = base64.StdEncoding.DecodeString(os.Getenv("K8S_TOKEN"))
		if err != nil {
			return nil, err
		}
	}

	var caCert []byte
	if caCertFlag != "" {
		caCert, err = base64.StdEncoding.DecodeString(caCertFlag)
		if err != nil {
			return nil, err
		}
	} else {
		caCert, err = base64.StdEncoding.DecodeString(os.Getenv("K8S_CA_CERT"))
		if err != nil {
			return nil, err
		}
	}

	if len(caCert) != 0 {
		return &rest.Config{
			Host: clusterURL,
			TLSClientConfig: rest.TLSClientConfig{
				CAData: caCert,
			},
			BearerToken: string(token),
		}, nil
	}
	return &rest.Config{
		Host: clusterURL,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: true, // Disable TLS verification
		},
		BearerToken: string(token),
	}, nil
}

// CreateConfigFromPath generates the k8s config object from a path to a config file
func CreateConfigFromPath(configPath string, context string) (*rest.Config, error) {
	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: configPath}
	configOverrides := &clientcmd.ConfigOverrides{}

	if context != "" {
		configOverrides.CurrentContext = context
	}
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
}

// CreateConfigFromURL generates the k8s config object from a k8s cluster URL
func CreateConfigFromURL(clusterURL string) *rest.Config {
	return &rest.Config{
		Host: clusterURL,
	}
}
