package main

import (
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/genkiroid/cert"

	dexterCmd "github.com/gini/dexter/cmd"
	"github.com/gookit/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientCmdApi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	BANNER = `
	#    # #    # #####  #   #
	#   #  #    # #    #  # #
	####   #    # #####    #
	#  #   #    # #    #   #
	#   #  #    # #    #   #
	#    #  ####  #####    #
`
)

var (
	servers []string
	rootCmd = &cobra.Command{
		Use:   "sre-kuby",
		Short: "A OpenId connect authentication helper for Kubernetes",
		Long: fmt.Sprintf(`%s
sre-kube is a tool to generate the kube configuration based on your google gsuite account..`, BANNER),
		RunE: rootCommand,
	}
)

func init() {
	defaultServers := []string{
		"dev:k8s-api-dev.example.de",
		"staging:k8s-api-staging.example.de",
		"phdp:k8s-api-prod.example.de",
	}
	rootCmd.PersistentFlags().StringArrayVarP(&servers, "servers", "s", defaultServers, "The k8s api servers to include in the config. In the format name:server. (for example dev:k8s-api-dev.example.de)")
	rootCmd.MarkFlagRequired("servers")
}

// Execute executes the root command.
func rootCommand(cocmd *cobra.Command, args []string) error {
	//set the args for the dexter auth command call
	home := os.Getenv("HOME")

	if _, err := os.Stat(home + "/.kube/config_auth"); os.IsNotExist(err) {
		color.Danger.Printf("The 'config_auth' file couldn't be found. Either ways you miss to download it from 1password vault 'GSuite' secret 'K8S oidc configuration'.\nOr you didn't put it in the '%s/.kube/' folder.\n", home)
		return err
	}

	os.Args = []string{"dexters", "auth", "-k", home + "/.kube/config_auth"}
	fmt.Printf("Args %s", os.Args)
	err := dexterCmd.Execute()
	if err != nil {
		return err

	}
	authKubeConfig := dexterCmd.AuthCmd.PersistentFlags().Lookup("kube-config").Value.String()
	dir := path.Dir(authKubeConfig)

	kubeConfig := dir + "/config_new"
	copy(authKubeConfig, kubeConfig)

	clientCfg, _ := clientcmd.LoadFromFile(kubeConfig)
	var user string
	for key := range clientCfg.AuthInfos {
		if strings.HasPrefix(key, "USERNAME") {
			continue
		}
		user = key
	}

	fmt.Printf("servers %s\n", servers)
	cert.SkipVerify = true

	for _, server := range servers {
		nameServer := strings.Split(server, ":")
		if len(nameServer) != 2 {
			return errors.New("the server was not in the format name:server")
		}
		certificate := cert.NewCert(nameServer[1] + ":6443")
		if len(certificate.Error) > 0 {
			fmt.Printf("Skip server '%s' because it wasn't reachable. ('%s')\n", nameServer[1], certificate.Error)
			continue
		}

		var certificateBody string
		for _, cert := range certificate.CertChain() {
			certificateBody = certificateBody + string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw}))
		}
		clientCfg.Clusters[nameServer[0]] = &clientCmdApi.Cluster{
			Server: "https://" + nameServer[1] + ":6443",
			// TODO: use CertificateAuthorityData: []byte(certificateBody),
			InsecureSkipTLSVerify: true,
		}
		clientCfg.Contexts[nameServer[0]] = &clientCmdApi.Context{
			Cluster:  nameServer[0],
			AuthInfo: user,
		}
		clientCfg.CurrentContext = nameServer[0]
	}

	err = mergeK8sConfig(clientCfg, kubeConfig)
	if err != nil {
		return err
	}

	// ask to overwrite the .kube/config
	if askForConfirmation(fmt.Sprintf("Do you want to overwrite the '%s/.kube/config'.\nThe original config will be saved in the '%s/.kube/config_bak' file", home, home)) == true {
		copy(home+"/.kube/config", home+"/.kube/config_bak")
		fmt.Printf("Backup old kubectl config to those location:'%s/.kube/config_bak.'\n", home)
		copy(kubeConfig, home+"/.kube/config")
	} else {
		fmt.Printf("The generated kubectl config location:'%s'.\n", kubeConfig)
	}
	return nil
}

func mergeK8sConfig(clientCfg *clientCmdApi.Config, kubeConfig string) error {
	tempKubeConfig, err := ioutil.TempFile("", "")
	defer os.Remove(tempKubeConfig.Name())

	if err != nil {
		return err
	}

	// write snipped to temporary file
	clientcmd.WriteToFile(*clientCfg, tempKubeConfig.Name())

	// setup the order for the file load
	loadingRules := clientcmd.ClientConfigLoadingRules{
		Precedence: []string{tempKubeConfig.Name(), kubeConfig},
	}

	// merge the configs
	mergedConfig, err := loadingRules.Load()

	if err != nil {
		return fmt.Errorf("failed to merge configurations: %s", err)
	}

	// write the merged data to the k8s config
	err = clientcmd.WriteToFile(*mergedConfig, kubeConfig)

	if err != nil {
		return fmt.Errorf("failed to write merged configuration: %s", err)
	}
	return nil
}

func copy(src string, dst string) {
	// Read all content of src to data
	data, _ := ioutil.ReadFile(src)

	// Write data to dst
	_ = ioutil.WriteFile(dst, data, 0644)
}

// askForConfirmation uses Scanln to parse user input. A user must type in "yes" or "no" and
// then press enter. It has fuzzy matching, so "y", "Y", "yes", "YES", and "Yes" all count as
// confirmations. If the input is not recognized, it will ask again. The function does not return
// until it gets a valid response from the user. Typically, you should use fmt to print out a question
// before calling askForConfirmation. E.g. fmt.Println("WARNING: Are you sure? (yes/no)")
func askForConfirmation(msg string) bool {
	var response string
	fmt.Printf("%s (y/n):", msg)
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation(msg)
	}
}

func containsString(slice []string, element string) bool {
	for _, elem := range slice {
		if elem == element {
			return true
		}
	}
	return false
}

func main() {
	// set log format & level
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.InfoLevel)

	rootCmd.Execute()
}
